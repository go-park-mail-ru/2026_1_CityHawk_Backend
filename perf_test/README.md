# ДЗ 4: оптимизация работы СУБД

## Цель

В этой работе описан воспроизводимый цикл нагрузочного тестирования и
оптимизации backend-сервиса CityHawk:

1. Провести нагрузочный тест endpoint'а создания основной сущности.
2. Заполнить базу 100 тыс. основными сущностями.
3. Провести нагрузочный тест endpoint'а чтения на большом объеме данных.
4. Проанализировать метрики PostgreSQL и планы выполнения запросов.
5. Выполнить оптимизацию.
6. Повторить тест чтения и сравнить результаты.

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

## Структура файлов

```text
perf_test/
  README.md                         описание методики
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
  results/                          результаты тестов, DB stats и explain-отчеты
  reports/                          vegeta plots и скриншоты
```

Оптимизационная миграция:

```text
db/migrations/0030_perf_event_indexes.up.sql
db/migrations/0030_perf_event_indexes.down.sql
```

## Окружение

Тест выполняется на выделенной VM, как требуется в задании.

Параметры окружения фиксируются перед запуском:

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
perf_test/results/baseline-create.txt
perf_test/results/baseline-create-fill.txt
perf_test/results/final-count.txt
```

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
perf_test/results/baseline-read.txt
perf_test/results/baseline-db-stats.txt
perf_test/results/baseline-explain-read-events.txt
```

## Анализ бутылочного горлышка

Основной read-запрос строится в `internal/place/repository/postgres.go` функцией
`buildListEventsQuery`.

Проблемные места при чтении большого объема событий:

- сканирование и сортировка `event_session` для поиска ближайшей активной
  сессии;
- `DISTINCT ON (event_id)` по `event_image` для выбора первой картинки;
- повторяющиеся `EXISTS`-проверки по `event_id` в `event_session`;
- сортировка `event` по `created_at`, `title` или времени ближайшей сессии;
- фильтрация по городу через `event_place`, `place` и `event_session`;
- trigram-поиск по `lower(title)`;
- выполнение дорогих join'ов и агрегаций до применения `LIMIT`.

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

Вторая оптимизация находится в `internal/place/repository/postgres.go`: запрос
списка событий сначала выбирает страницу `event.id` в CTE `page_events`, а
затем подтягивает изображения, теги, ближайшие сессии, места и favorite counts
только для этой маленькой страницы. Это снижает объем работы, которая раньше
выполнялась по связанным таблицам для всех 100 тыс. событий до `LIMIT`.

## Итерация 2: чтение после оптимизации

Команда для проверки варианта с индексами:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=optimized-read ./perf_test/scripts/run_read_test.sh
```

Команда для проверки варианта после переписывания запроса:

```bash
COUNT=5000 RATE=20 RESULT_PREFIX=optimized-read-v2 ./perf_test/scripts/run_read_test.sh
```

Сбор статистики БД:

```bash
psql "$DATABASE_URL" -f perf_test/scripts/collect_db_stats.sql > perf_test/results/optimized-db-stats.txt
psql "$DATABASE_URL" -f perf_test/scripts/explain_read_events.sql > perf_test/results/optimized-explain-read-events.txt
```

Артефакты:

```text
perf_test/results/optimized-read.txt
perf_test/results/optimized-read-v2.txt
perf_test/results/optimized-db-stats.txt
perf_test/results/optimized-v2-db-stats.txt
perf_test/results/optimized-explain-read-events.txt
perf_test/results/optimized-v2-explain-read-events.txt
```

## Дальнейшие оптимизации

Если endpoint чтения остается bottleneck'ом, следующие шаги:

- вынести дорогой total count из горячего пути;
- перейти с offset pagination на keyset pagination для основных сортировок;
- создать read model или materialized view для карточек событий;
- кешировать часто запрашиваемые страницы ленты;
- отдельно проверить pool size PostgreSQL и лимиты backend-сервисов;
- разделить тяжелую выдачу карточек и легкие поисковые endpoint'ы.

## Где смотреть результаты

Итоговая сводка:

```text
perf_test/results/final-summary.md
```

Фактические результаты:

```text
perf_test/results/baseline-create.txt
perf_test/results/baseline-create-fill.txt
perf_test/results/final-count.txt
perf_test/results/baseline-read.txt
perf_test/results/baseline-db-stats.txt
perf_test/results/optimized-read.txt
perf_test/results/optimized-read-v2.txt
```

Теоретические материалы и пояснения к планам:

```text
perf_test/results/baseline-explain-read-events.txt
perf_test/results/optimized-explain-read-events.txt
perf_test/results/optimized-v2-explain-read-events.txt
perf_test/results/optimized-v2-db-stats.txt
perf_test/results/theoretical-stable-read.txt
perf_test/results/theoretical-summary.md
```

## Чеклист воспроизводимости

Перед сдачей в репозитории должны быть:

- `perf_test/init.sql`
- все скрипты в `perf_test/scripts`
- результаты vegeta в `perf_test/results`
- DB stats и explain outputs в `perf_test/results`
- HTML-графики или скриншоты в `perf_test/reports`
- оптимизационная миграция `0030_perf_event_indexes`
