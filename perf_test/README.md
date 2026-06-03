# ДЗ 4: оптимизация работы СУБД

## Цель

В этой работе задокументирован воспроизводимый цикл нагрузочного тестирования и
оптимизации backend-сервиса CityHawk:

1. Провести нагрузочный тест endpoint'а создания основной сущности.
2. Провести нагрузочный тест endpoint'а чтения после заполнения базы 100 тыс.
   основными сущностями.
3. Проанализировать метрики PostgreSQL и планы выполнения запросов.
4. Выполнить оптимизацию.
5. Повторить тест чтения и сравнить результаты.

Основная бизнес-сущность проекта - `event`, то есть событие. Именно события
показываются в ленте, на карте, в подборках и на странице деталей.

## Тестируемое API

Endpoint создания:

```text
POST /api/events
```

Endpoint чтения:

```text
GET /api/events?limit=12&offset=0&sort=dateAsc
```

В read-тест также включены типовые варианты того же endpoint'а:

```text
GET /api/events?sort=titleAsc
GET /api/events?query=Perf+Event+09
GET /api/events?cityId=10000000-0000-0000-0000-000000000001
GET /api/events?categoryId=20000000-0000-0000-0000-000000000001
GET /api/events?tag=30000000-0000-0000-0000-000000000001
```

## Инструменты

Генератор нагрузки: `vegeta`.

Анализ базы данных:

- `pg_stat_statements`
- `EXPLAIN (ANALYZE, BUFFERS)`
- PostgreSQL slow logs и `auto_explain`
- Prometheus, Grafana и `postgres_exporter`

Наблюдаемость PostgreSQL включается через `db/postgres/postgresql.conf` и
`docker-compose.yml`.

## Файлы

```text
perf_test/
  README.md                         этот отчет
  init.sql                          baseline DDL до оптимизационной миграции 0030
  scripts/
    perf_seed.sql                   детерминированные справочники для теста
    prepare_db.sh                   загрузка справочников
    auth.sh                         создание/логин тестового пользователя
    generate_create_targets.py      генерация vegeta targets для POST /api/events
    generate_read_targets.py        генерация vegeta targets для GET /api/events
    run_create_test.sh              запуск create-теста
    run_read_test.sh                запуск read-теста
    collect_db_stats.sql            сбор pg_stat_statements и счетчиков БД
    explain_read_events.sql         representative EXPLAIN для read-запроса
    reset_pg_stat_statements.sql    сброс статистики между итерациями
  results/                          сырые результаты vegeta и БД
  reports/                          vegeta plots и скриншоты
```

Оптимизационная миграция:

```text
db/migrations/0030_perf_event_indexes.up.sql
db/migrations/0030_perf_event_indexes.down.sql
```

## Окружение

Тест выполнялся на выделенной VM, как требуется в задании.

Параметры окружения:

```text
Дата: 2026-06-03
Хост: 2026-1-cityhawk VM
CPU: заполнить с VM
RAM: заполнить с VM
Диск: заполнить с VM
OS: Ubuntu
Docker: заполнить из `docker version`
PostgreSQL: 16
Go: заполнить из `go version`
Vegeta: заполнить из `vegeta -version`
```

Команды для проверки окружения:

```bash
uname -a
docker version
docker compose version
vegeta -version
```

## Подготовка baseline

Запуск приложения на VM:

```bash
docker compose up -d postgres photon cityhawk-auth-service cityhawk-profile-service cityhawk-events-service cityhawk-support-service cityhawk-social-service cityhawk-backend
```

Опционально запуск мониторинга:

```bash
docker compose --profile monitoring up -d
```

Подготовка детерминированных справочников:

```bash
export DATABASE_URL='postgres://cityhawk_migrator:cityhawk_migrator@localhost:5432/cityhawk?sslmode=disable'
./perf_test/scripts/prepare_db.sh
```

Создание и логин нагрузочного пользователя:

```bash
export BASE_URL='http://localhost:8080'
./perf_test/scripts/auth.sh
```

Сброс статистики запросов перед каждой итерацией:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/reset_pg_stat_statements.sql
```

## Итерация 1: создание 100 тыс. событий

Команда:

```bash
COUNT=100000 RATE=200 RESULT_PREFIX=baseline-create ./perf_test/scripts/run_create_test.sh
```

Ожидаемое поведение:

- `POST /api/events` создает 100 тыс. строк `event` через публичное API.
- Каждое событие содержит одну категорию, один тег, один URL изображения, одно
  основное место и одну сессию.
- Сессии распределяются по 2048 местам и непересекающимся временным слотам,
  чтобы удовлетворять constraint `event_session_place_time_excl`.

Артефакты:

```text
perf_test/results/baseline-create.bin
perf_test/results/baseline-create.txt
perf_test/results/baseline-create.json
perf_test/reports/baseline-create.html
```

Результат:

```text
Основной create-прогон, COUNT=100000 RATE=200:

Requests      [total, rate, throughput]  100000, 200.00, 181.03
Duration      [total, attack, wait]      8m23.387213159s, 8m19.995433871s, 3.391779288s
Latencies     [mean, 50, 95, 99, max]    1.877587155s, 1.837538605s, 4.620437972s, 5.203054522s, 5.423837861s
Bytes In      [total, mean]              4546682, 45.47
Bytes Out     [total, mean]              59994228, 599.94
Success       [ratio]                    91.13%
Status Codes  [code:count]               0:2  201:91127  500:8871
Error Set:
500 Internal Server Error
Post "http://localhost:8080/api/events": EOF

Дозагрузка недостающих событий, COUNT=2 RATE=1:

Requests      [total, rate, throughput]  2, 2.00, 1.98
Duration      [total, attack, wait]      1.007600486s, 1.000012235s, 7.588251ms
Latencies     [mean, 50, 95, 99, max]    7.683488ms, 7.683488ms, 7.778725ms, 7.778725ms, 7.778725ms
Bytes In      [total, mean]              92, 46.00
Bytes Out     [total, mean]              1200, 600.00
Success       [ratio]                    100.00%
Status Codes  [code:count]               201:2
Error Set:

Финальная проверка БД:

SELECT count(*) FROM event WHERE title LIKE 'Perf Event %%';
count = 100000
```

Endpoint создания достиг практического предела VM при `RATE=200`: p99 превысил
5 секунд, сервис вернул `8871` ответов HTTP 500. Недостающие события были
дозагружены отдельным low-rate прогоном. Требование создать 100 тыс. основных
сущностей выполнено, что подтверждено финальным count в базе.

## Итерация 1: baseline чтения

Команда:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=baseline-read ./perf_test/scripts/run_read_test.sh
```

Сбор статистики БД:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/baseline-db-stats.txt
psql "$DATABASE_URL" -f perf_test/scripts/explain_read_events.sql > perf_test/results/baseline-explain-read-events.txt
```

Артефакты:

```text
perf_test/results/baseline-read.bin
perf_test/results/baseline-read.txt
perf_test/results/baseline-read.json
perf_test/reports/baseline-read.html
perf_test/results/baseline-db-stats.txt
perf_test/results/baseline-explain-read-events.txt
```

Результат:

```text
Requests      [total, rate, throughput]  5000, 20.00, 0.00
Duration      [total, attack, wait]      4m39.950655221s, 4m9.95022971s, 30.000425511s
Latencies     [mean, 50, 95, 99, max]    29.963377006s, 30.000627419s, 30.001316485s, 30.001713638s, 30.002338935s
Bytes In      [total, mean]              80, 0.02
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.00%
Status Codes  [code:count]               0:4998  500:2
Error Set:
500 Internal Server Error
context deadline exceeded / EOF для вариантов GET /api/events
```

После генерации 100 тыс. событий endpoint чтения не давал стабильных ответов
даже при `RATE=20`. Vegeta упиралась в timeout 30 секунд. Это стало главным
бутылочным горлышком для дальнейшей оптимизации.

## Анализ бутылочного горлышка

Основной read-запрос строится в `internal/place/repository/postgres.go` функцией
`buildListEventsQuery`.

Ожидаемые проблемные места после 100 тыс. событий:

- сканирование и сортировка `event_session` для поиска ближайшей активной
  сессии;
- `DISTINCT ON (event_id)` по `event_image` для выбора первой картинки;
- повторяющиеся `EXISTS`-проверки по `event_id` в `event_session`;
- сортировка `event` по `created_at`, `title` или времени ближайшей сессии;
- фильтрация по городу через `event_place`, `place` и `event_session`;
- trigram-поиск по `lower(title)`.

При анализе baseline-плана нужно смотреть на:

- sequential scan по большим таблицам;
- внешние сортировки или большие in-memory sort;
- большое число прочитанных shared blocks;
- nested loop, умножающие работу на количество событий;
- высокий `total_exec_time` в `pg_stat_statements`.

## Оптимизация

Первая оптимизация вынесена в миграцию:

```text
db/migrations/0030_perf_event_indexes.up.sql
```

Она добавляет индексы под горячие пути:

```sql
idx_event_created_at_id
idx_event_title_id
idx_event_author_created_at
idx_event_session_event_end_start
idx_event_session_start_at_id
idx_event_session_end_at_event
idx_event_image_event_created_id
idx_event_place_place_event
idx_place_city_id
idx_city_lower_name
idx_event_lower_title_trgm
```

Применение после сохранения baseline-результатов:

```bash
psql "$DATABASE_URL" -f db/migrations/0030_perf_event_indexes.up.sql
psql "$DATABASE_URL" -c 'ANALYZE;'
psql "$DATABASE_URL" -f perf_test/scripts/reset_pg_stat_statements.sql
```

Так как после index-only оптимизации endpoint чтения все еще уходил в timeout,
была применена вторая оптимизация в `internal/place/repository/postgres.go`:
запрос списка событий теперь сначала выбирает страницу `event.id` в CTE
`page_events`, а затем подтягивает изображения, теги, ближайшие сессии, места и
favorite counts только для этой маленькой страницы. Это снижает объем работы,
которая раньше выполнялась по связанным таблицам для всех 100 тыс. событий до
`LIMIT`.

Index-only оптимизация тестировалась как `optimized-read`. Она не решила timeout
на выбранной нагрузке. После этого была изменена форма запроса, и вариант с
pagination-first query shape тестировался как `optimized-read-v2`.

## Итерация 2: чтение после оптимизации

Команда:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=optimized-read ./perf_test/scripts/run_read_test.sh
```

Сбор статистики БД:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/optimized-db-stats.txt
psql "$DATABASE_URL" -f perf_test/scripts/explain_read_events.sql > perf_test/results/optimized-explain-read-events.txt
```

Артефакты:

```text
perf_test/results/optimized-read.bin
perf_test/results/optimized-read.txt
perf_test/results/optimized-read.json
perf_test/reports/optimized-read.html
perf_test/results/optimized-db-stats.txt
perf_test/results/optimized-explain-read-events.txt
```

Результат:

```text
Index-only оптимизация, COUNT=5000 RATE=20:

Requests      [total, rate, throughput]  5000, 20.00, 0.00
Duration      [total, attack, wait]      4m39.95140021s, 4m9.950580291s, 30.000819919s
Latencies     [mean, 50, 95, 99, max]    29.935932865s, 30.000629951s, 30.001310652s, 30.001711856s, 30.002416664s
Bytes In      [total, mean]              240, 0.05
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.00%
Status Codes  [code:count]               0:4994  500:6
Error Set:
500 Internal Server Error
context deadline exceeded / EOF для вариантов GET /api/events

Pagination-first rewrite, COUNT=5000 RATE=20:

Requests      [total, rate, throughput]  5000, 20.00, 0.01
Duration      [total, attack, wait]      4m39.950555106s, 4m9.95023258s, 30.000322526s
Latencies     [mean, 50, 95, 99, max]    29.966801045s, 30.000631888s, 30.001296391s, 30.001732848s, 30.002455233s
Bytes In      [total, mean]              23618, 4.72
Bytes Out     [total, mean]              0, 0.00
Success       [ratio]                    0.04%
Status Codes  [code:count]               0:4998  200:2
Error Set:
context deadline exceeded / EOF для вариантов GET /api/events
```

Index-only оптимизация сделала новые индексы видимыми в PostgreSQL stats,
особенно для `event_session`, `event_image`, `event_place` и `place`, но ее
оказалось недостаточно для выбранной read-нагрузки. Pagination-first rewrite
позволил получить небольшое количество успешных ответов, однако endpoint все
еще оставался нестабильным при `RATE=20`. Вывод: бутылочное горлышко связано не
только с покрытием индексами, но и с формой endpoint'а/запроса и текущей
производительностью сервиса и БД.

## Сравнение

| Сценарий | RPS/throughput | p50 | p95 | p99 | Успешность | Комментарий |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Создание baseline, 100 тыс. событий | 181.03 throughput | 1.84s | 4.62s | 5.20s | 91.13% основной прогон; итоговый count 100000 | `RATE=200` перегрузил путь создания; недостающие строки дозагружены отдельно |
| Дозагрузка создания | 1.98 throughput | 7.68ms | 7.78ms | 7.78ms | 100.00% | Дозагрузка недостающих событий |
| Чтение baseline | 0.00 throughput | 30.00s | 30.00s | 30.00s | 0.00% | До миграции 0030 |
| Чтение после индексов | 0.00 throughput | 30.00s | 30.00s | 30.00s | 0.00% | Миграции 0030 оказалось недостаточно |
| Чтение после переписывания запроса | 0.01 throughput | 30.00s | 30.00s | 30.00s | 0.04% | Запрос сначала выбирает страницу событий; endpoint все еще насыщен |

Сравнение БД:

| Метрика | Baseline | После индексов | Изменение |
| --- | ---: | ---: | ---: |
| total_exec_time самого тяжелого запроса в pg_stat_statements | 33233.08ms | 29920.14ms | -3312.94ms |
| mean_exec_time самого тяжелого запроса в pg_stat_statements | 6.65ms | 5.98ms | -0.67ms |
| DB shared blocks read | 9952 | 15057 | +5105 |
| DB shared blocks hit | 42643996 | 74667006 | +32023010 |
| Deadlocks | 0 | 0 | без изменений |

## Вывод

Нагрузка успешно создала 100 тыс. основных сущностей через публичное API.
Путь создания при `RATE=200` был близок к пределу VM: p95 составил `4.62s`, p99 -
`5.20s`, сервис возвращал ошибки. Low-rate дозагрузка завершила формирование
набора данных.

Read path стал основным бутылочным горлышком. При 100 тыс. сгенерированных
событий `GET /api/events` уходил в timeout 30 секунд даже при `RATE=20`.
Индексы из миграции `0030` использовались PostgreSQL, но index-only оптимизация
не восстановила стабильный read throughput. Вторая оптимизация изменила запрос:
сначала выбирается страница событий, затем подтягиваются связанные данные только
для этой страницы. Несмотря на это, endpoint все еще оставался насыщенным на VM.

Практический вывод: для стабильной работы CityHawk на 100 тыс. событий под
конкурентной нагрузкой нужна дальнейшая оптимизация read path: более узкие
запросы endpoint'ов, отказ от дорогого total count в горячем пути, кэширование
или предрассчитанная read model/materialized view для карточек событий.

## Чеклист воспроизводимости

Перед сдачей в репозитории должны быть:

- `perf_test/init.sql`
- все скрипты в `perf_test/scripts`
- сырые результаты vegeta в `perf_test/results`
- DB stats и explain outputs в `perf_test/results`
- HTML-графики или скриншоты в `perf_test/reports`
- заполненные таблицы сравнения в этом README
- оптимизационная миграция `0030_perf_event_indexes`
