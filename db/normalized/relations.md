## Принятые допущения

1. Во всех отношениях используются только атомарные атрибуты. Составные типы данных (`array`, `json` и т.д.) не применяются.
2. Во всех сущностях с суррогатным ключом атрибут `id` является первичным ключом.
3. В `USER_ACCOUNT` атрибут `email` уникален.
4. В `CATEGORY` и `TAG` атрибут `name` уникален.
5. В `SHARE_LINK` атрибут `share_token` уникален.
6. Для `CITY` дополнительно предполагается альтернативный ключ `UNIQUE(country_name, name)`.
7. В таблицах-связках с одним атрибутом `*_id` этот атрибут одновременно является первичным и внешним ключом, обеспечивая связь один-к-одному с базовой сущностью.
8. В `REFRESH_SESSION` атрибут `token_hash` является первичным ключом; хранится именно хеш refresh-токена, а не исходное значение токена.

## Общий вывод по нормальным формам

Все отношения находятся в 1НФ, потому что каждое поле хранит одно атомарное значение, а повторяющиеся группы вынесены в отдельные отношения.

Все отношения находятся во 2НФ, потому что:
- в отношениях с простым ключом все неключевые атрибуты зависят от полного ключа;
- в отношениях с составным ключом (`EVENT_CATEGORY`, `EVENT_TAG`, `COLLECTION_EVENT`, `FAVORITE_EVENT`, `USER_FOLLOW`) служебные атрибуты зависят только от всей комбинации ключевых атрибутов, а не от части ключа.

Все отношения находятся в 3НФ, потому что транзитивные зависимости устранены: данные о городах, местах, событиях, подборках, приглашениях и уведомлениях вынесены в отдельные отношения и не дублируются в зависимых сущностях.

Все отношения находятся в НФБК, потому что каждая нетривиальная функциональная зависимость имеет слева суперключ соответствующего отношения. Там, где есть альтернативные ключи (`email`, `name`, `share_token`, `(country_name, name)`), они также являются суперключами.

---

## Relation CITY

Краткое описание: справочник городов, используемый в профилях пользователей и у мест проведения.

### Functional dependencies

```text
Relation CITY:

{id} -> name, country_name, timezone, created_at, updated_at

{country_name, name} -> id, timezone, created_at, updated_at
```

### Обоснование нормальных форм

`CITY` находится в 1НФ, потому что все атрибуты атомарны.

`CITY` находится во 2НФ, потому что ключ `id` простой, а при использовании альтернативного ключа `{country_name, name}` неключевые атрибуты зависят от всего ключа целиком.

`CITY` находится в 3НФ, потому что нет транзитивных зависимостей между неключевыми атрибутами.

`CITY` находится в НФБК, потому что детерминанты `{id}` и `{country_name, name}` являются суперключами.

---

## Relation USER_ACCOUNT

Краткое описание: учетная запись пользователя системы.

### Functional dependencies

```text
Relation USER_ACCOUNT:

{id} -> email, username, user_surname, password_hash, birthday, city_id, avatar_url, created_at, updated_at

{email} -> id, username, user_surname, password_hash, birthday, city_id, avatar_url, created_at, updated_at
```

### Обоснование нормальных форм

`USER_ACCOUNT` находится в 1НФ, потому что все поля атомарны.

`USER_ACCOUNT` находится во 2НФ, потому что ключи `id` и `email` не составные и все остальные атрибуты зависят от полного ключа.

`USER_ACCOUNT` находится в 3НФ, потому что данные о городе пользователя вынесены в `CITY`, а значит нет транзитивной зависимости вида `id -> city_id -> city_name`.

`USER_ACCOUNT` находится в НФБК, потому что все детерминанты `{id}` и `{email}` являются суперключами.

---

## Relation REFRESH_SESSION

Краткое описание: серверные сессии refresh-токенов, используемые для ротации и отзыва токенов аутентификации.

### Functional dependencies

```text
Relation REFRESH_SESSION:

{token_hash} -> user_id, expires_at, created_at
```

### Обоснование нормальных форм

`REFRESH_SESSION` находится в 1НФ, потому что все атрибуты атомарны, а значение `token_hash` хранится как одиночная строка.

`REFRESH_SESSION` находится во 2НФ, потому что первичный ключ `token_hash` простой, и остальные атрибуты зависят от него целиком.

`REFRESH_SESSION` находится в 3НФ, потому что данные пользователя не дублируются в строке сессии, а вынесены в `USER_ACCOUNT`; транзитивных зависимостей между неключевыми атрибутами нет.

`REFRESH_SESSION` находится в НФБК, потому что единственный детерминант `{token_hash}` является суперключом.

---

## Relation PLACE

Краткое описание: место проведения события с адресом и координатами.

### Functional dependencies

```text
Relation PLACE:

{id} -> city_id, name, address_line, latitude, longitude, description, created_at, updated_at
```

### Обоснование нормальных форм

`PLACE` находится в 1НФ, потому что все атрибуты атомарны.

`PLACE` находится во 2НФ, потому что первичный ключ `id` простой.

`PLACE` находится в 3НФ, потому что сведения о городе вынесены в `CITY` и не дублируются в строке места.

`PLACE` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation CATEGORY

Краткое описание: справочник категорий событий.

### Functional dependencies

```text
Relation CATEGORY:

{id} -> name, created_at, updated_at

{name} -> id, created_at, updated_at
```

### Обоснование нормальных форм

`CATEGORY` находится в 1НФ, потому что все атрибуты атомарны.

`CATEGORY` находится во 2НФ, потому что ключи `id` и `name` простые.

`CATEGORY` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`CATEGORY` находится в НФБК, потому что детерминанты `{id}` и `{name}` являются суперключами.

---

## Relation TAG

Краткое описание: справочник тегов событий.

### Functional dependencies

```text
Relation TAG:

{id} -> name, created_at, updated_at

{name} -> id, created_at, updated_at
```

### Обоснование нормальных форм

`TAG` находится в 1НФ, потому что все атрибуты атомарны.

`TAG` находится во 2НФ, потому что ключи `id` и `name` простые.

`TAG` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`TAG` находится в НФБК, потому что детерминанты `{id}` и `{name}` являются суперключами.

---

## Relation EVENT

Краткое описание: основная карточка события без конкретного сеанса.

### Functional dependencies

```text
Relation EVENT:

{id} -> author_user_id, title, short_description, full_description, age_limit, source_url, created_at, updated_at
```

### Обоснование нормальных форм

`EVENT` находится в 1НФ, потому что все атрибуты атомарны.

`EVENT` находится во 2НФ, потому что первичный ключ `id` простой.

`EVENT` находится в 3НФ, потому что время, место, категории, теги и изображения вынесены в отдельные отношения.

`EVENT` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation EVENT_SESSION

Краткое описание: конкретный сеанс события в заданном месте и временном интервале.

### Functional dependencies

```text
Relation EVENT_SESSION:

{id} -> event_id, place_id, start_at, end_at, price, created_at, updated_at
```

### Обоснование нормальных форм

`EVENT_SESSION` находится в 1НФ, потому что все атрибуты атомарны.

`EVENT_SESSION` находится во 2НФ, потому что первичный ключ `id` простой.

`EVENT_SESSION` находится в 3НФ, потому что информация о событии и месте хранится в `EVENT` и `PLACE`, а не дублируется здесь.

`EVENT_SESSION` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation EVENT_IMAGE

Краткое описание: изображение, относящееся к событию.

### Functional dependencies

```text
Relation EVENT_IMAGE:

{id} -> event_id, image_url, created_at, updated_at
```

### Обоснование нормальных форм

`EVENT_IMAGE` находится в 1НФ, потому что одно изображение хранится в одной строке.

`EVENT_IMAGE` находится во 2НФ, потому что первичный ключ `id` простой.

`EVENT_IMAGE` находится в 3НФ, потому что атрибуты зависят только от идентификатора изображения.

`EVENT_IMAGE` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation EVENT_CATEGORY

Краткое описание: связь многие-ко-многим между событиями и категориями.

### Functional dependencies

```text
Relation EVENT_CATEGORY:

{event_id, category_id} -> created_at, updated_at
```

### Обоснование нормальных форм

`EVENT_CATEGORY` находится в 1НФ, потому что все значения атомарны.

`EVENT_CATEGORY` находится во 2НФ, потому что служебные атрибуты зависят от всего составного ключа `{event_id, category_id}`, а не от его части.

`EVENT_CATEGORY` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`EVENT_CATEGORY` находится в НФБК, потому что единственный детерминант `{event_id, category_id}` является суперключом.

---

## Relation EVENT_TAG

Краткое описание: связь многие-ко-многим между событиями и тегами.

### Functional dependencies

```text
Relation EVENT_TAG:

{event_id, tag_id} -> created_at, updated_at
```

### Обоснование нормальных форм

`EVENT_TAG` находится в 1НФ, потому что все значения атомарны.

`EVENT_TAG` находится во 2НФ, потому что служебные атрибуты зависят от полного составного ключа `{event_id, tag_id}`.

`EVENT_TAG` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`EVENT_TAG` находится в НФБК, потому что единственный детерминант `{event_id, tag_id}` является суперключом.

---

## Relation COLLECTION

Краткое описание: пользовательская подборка событий.

### Functional dependencies

```text
Relation COLLECTION:

{id} -> author_user_id, title, description, is_public, created_at, updated_at
```

### Обоснование нормальных форм

`COLLECTION` находится в 1НФ, потому что все атрибуты атомарны.

`COLLECTION` находится во 2НФ, потому что первичный ключ `id` простой.

`COLLECTION` находится в 3НФ, потому что список событий подборки вынесен в отдельное отношение `COLLECTION_EVENT`.

`COLLECTION` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation COLLECTION_IMAGE

Краткое описание: изображение, относящееся к пользовательской подборке.

### Functional dependencies

```text
Relation COLLECTION_IMAGE:

{id} -> collection_id, image_url, created_at, updated_at
```

### Обоснование нормальных форм

`COLLECTION_IMAGE` находится в 1НФ, потому что одно изображение хранится в одной строке.

`COLLECTION_IMAGE` находится во 2НФ, потому что первичный ключ `id` простой.

`COLLECTION_IMAGE` находится в 3НФ, потому что атрибуты зависят только от идентификатора изображения.

`COLLECTION_IMAGE` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation COLLECTION_EVENT

Краткое описание: связь многие-ко-многим между подборками и событиями.

### Functional dependencies

```text
Relation COLLECTION_EVENT:

{collection_id, event_id} -> created_at, updated_at
```

### Обоснование нормальных форм

`COLLECTION_EVENT` находится в 1НФ, потому что все значения атомарны.

`COLLECTION_EVENT` находится во 2НФ, потому что служебные атрибуты зависят от полного составного ключа `{collection_id, event_id}`.

`COLLECTION_EVENT` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`COLLECTION_EVENT` находится в НФБК, потому что единственный детерминант `{collection_id, event_id}` является суперключом.

---

## Relation FAVORITE_EVENT

Краткое описание: отметка события как избранного пользователем.

### Functional dependencies

```text
Relation FAVORITE_EVENT:

{user_id, event_id} -> created_at, updated_at
```

### Обоснование нормальных форм

`FAVORITE_EVENT` находится в 1НФ, потому что все значения атомарны.

`FAVORITE_EVENT` находится во 2НФ, потому что служебные атрибуты зависят от полного составного ключа `{user_id, event_id}`.

`FAVORITE_EVENT` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`FAVORITE_EVENT` находится в НФБК, потому что единственный детерминант `{user_id, event_id}` является суперключом.

---

## Relation USER_FOLLOW

Краткое описание: подписка одного пользователя на другого пользователя.

### Functional dependencies

```text
Relation USER_FOLLOW:

{follower_user_id, followed_user_id} -> created_at, updated_at
```

### Обоснование нормальных форм

`USER_FOLLOW` находится в 1НФ, потому что все значения атомарны.

`USER_FOLLOW` находится во 2НФ, потому что служебные атрибуты зависят от полного составного ключа `{follower_user_id, followed_user_id}`.

`USER_FOLLOW` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`USER_FOLLOW` находится в НФБК, потому что единственный детерминант `{follower_user_id, followed_user_id}` является суперключом.

---

## Relation EVENT_INVITATION

Краткое описание: базовая сущность приглашения от одного пользователя другому.

### Functional dependencies

```text
Relation EVENT_INVITATION:

{id} -> sender_user_id, recipient_user_id, status, message_text, responded_at, created_at, updated_at
```

### Обоснование нормальных форм

`EVENT_INVITATION` находится в 1НФ, потому что все атрибуты атомарны.

`EVENT_INVITATION` находится во 2НФ, потому что первичный ключ `id` простой.

`EVENT_INVITATION` находится в 3НФ, потому что объект приглашения вынесен отдельно, а его цель хранится не здесь, а в специальных отношениях.

`EVENT_INVITATION` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation EVENT_INVITATION_EVENT

Краткое описание: связь приглашения с событием как целевой сущностью.

### Functional dependencies

```text
Relation EVENT_INVITATION_EVENT:

{invitation_id} -> event_id
```

### Обоснование нормальных форм

`EVENT_INVITATION_EVENT` находится в 1НФ, потому что оба атрибута атомарны.

`EVENT_INVITATION_EVENT` находится во 2НФ, потому что ключ `invitation_id` простой.

`EVENT_INVITATION_EVENT` находится в 3НФ, потому что нет транзитивных зависимостей между неключевыми атрибутами.

`EVENT_INVITATION_EVENT` находится в НФБК, потому что детерминант `{invitation_id}` является суперключом.

---

## Relation EVENT_INVITATION_SESSION

Краткое описание: связь приглашения с конкретным сеансом события как целевой сущностью.

### Functional dependencies

```text
Relation EVENT_INVITATION_SESSION:

{invitation_id} -> event_session_id
```

### Обоснование нормальных форм

`EVENT_INVITATION_SESSION` находится в 1НФ, потому что оба атрибута атомарны.

`EVENT_INVITATION_SESSION` находится во 2НФ, потому что ключ `invitation_id` простой.

`EVENT_INVITATION_SESSION` находится в 3НФ, потому что нет транзитивных зависимостей.

`EVENT_INVITATION_SESSION` находится в НФБК, потому что детерминант `{invitation_id}` является суперключом.

---

## Relation SHARE_LINK

Краткое описание: базовая сущность публичной ссылки для шаринга.

### Functional dependencies

```text
Relation SHARE_LINK:

{id} -> creator_user_id, share_token, created_at, updated_at

{share_token} -> id, creator_user_id, created_at, updated_at
```

### Обоснование нормальных форм

`SHARE_LINK` находится в 1НФ, потому что все поля атомарны.

`SHARE_LINK` находится во 2НФ, потому что ключи `id` и `share_token` простые.

`SHARE_LINK` находится в 3НФ, потому что целевой объект ссылки вынесен в отдельные отношения, а не хранится вместе с базовыми атрибутами ссылки.

`SHARE_LINK` находится в НФБК, потому что детерминанты `{id}` и `{share_token}` являются суперключами.

---

## Relation SHARE_LINK_EVENT

Краткое описание: связь ссылки шаринга с событием.

### Functional dependencies

```text
Relation SHARE_LINK_EVENT:

{share_link_id} -> event_id
```

### Обоснование нормальных форм

`SHARE_LINK_EVENT` находится в 1НФ, потому что оба атрибута атомарны.

`SHARE_LINK_EVENT` находится во 2НФ, потому что ключ `share_link_id` простой.

`SHARE_LINK_EVENT` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`SHARE_LINK_EVENT` находится в НФБК, потому что детерминант `{share_link_id}` является суперключом.

---

## Relation SHARE_LINK_COLLECTION

Краткое описание: связь ссылки шаринга с подборкой.

### Functional dependencies

```text
Relation SHARE_LINK_COLLECTION:

{share_link_id} -> collection_id
```

### Обоснование нормальных форм

`SHARE_LINK_COLLECTION` находится в 1НФ, потому что оба атрибута атомарны.

`SHARE_LINK_COLLECTION` находится во 2НФ, потому что ключ `share_link_id` простой.

`SHARE_LINK_COLLECTION` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`SHARE_LINK_COLLECTION` находится в НФБК, потому что детерминант `{share_link_id}` является суперключом.

---

## Relation NOTIFICATION

Краткое описание: базовая сущность уведомления для получателя.

### Functional dependencies

```text
Relation NOTIFICATION:

{id} -> recipient_user_id, notification_type, is_read, read_at, created_at, updated_at
```

### Обоснование нормальных форм

`NOTIFICATION` находится в 1НФ, потому что все атрибуты атомарны.

`NOTIFICATION` находится во 2НФ, потому что первичный ключ `id` простой.

`NOTIFICATION` находится в 3НФ, потому что автор, событие, сеанс, приглашение и подборка вынесены в отдельные отношения и не образуют транзитивных зависимостей внутри строки уведомления.

`NOTIFICATION` находится в НФБК, потому что единственный детерминант `{id}` является суперключом.

---

## Relation NOTIFICATION_ACTOR

Краткое описание: связь уведомления с пользователем-инициатором.

### Functional dependencies

```text
Relation NOTIFICATION_ACTOR:

{notification_id} -> author_user_id
```

### Обоснование нормальных форм

`NOTIFICATION_ACTOR` находится в 1НФ, потому что оба атрибута атомарны.

`NOTIFICATION_ACTOR` находится во 2НФ, потому что ключ `notification_id` простой.

`NOTIFICATION_ACTOR` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`NOTIFICATION_ACTOR` находится в НФБК, потому что детерминант `{notification_id}` является суперключом.

---

## Relation NOTIFICATION_EVENT

Краткое описание: связь уведомления с событием.

### Functional dependencies

```text
Relation NOTIFICATION_EVENT:

{notification_id} -> event_id
```

### Обоснование нормальных форм

`NOTIFICATION_EVENT` находится в 1НФ, потому что оба атрибута атомарны.

`NOTIFICATION_EVENT` находится во 2НФ, потому что ключ `notification_id` простой.

`NOTIFICATION_EVENT` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`NOTIFICATION_EVENT` находится в НФБК, потому что детерминант `{notification_id}` является суперключом.

---

## Relation NOTIFICATION_EVENT_SESSION

Краткое описание: связь уведомления с конкретным сеансом события.

### Functional dependencies

```text
Relation NOTIFICATION_EVENT_SESSION:

{notification_id} -> event_session_id
```

### Обоснование нормальных форм

`NOTIFICATION_EVENT_SESSION` находится в 1НФ, потому что оба атрибута атомарны.

`NOTIFICATION_EVENT_SESSION` находится во 2НФ, потому что ключ `notification_id` простой.

`NOTIFICATION_EVENT_SESSION` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`NOTIFICATION_EVENT_SESSION` находится в НФБК, потому что детерминант `{notification_id}` является суперключом.

---

## Relation NOTIFICATION_INVITATION

Краткое описание: связь уведомления с приглашением.

### Functional dependencies

```text
Relation NOTIFICATION_INVITATION:

{notification_id} -> invitation_id
```

### Обоснование нормальных форм

`NOTIFICATION_INVITATION` находится в 1НФ, потому что оба атрибута атомарны.

`NOTIFICATION_INVITATION` находится во 2НФ, потому что ключ `notification_id` простой.

`NOTIFICATION_INVITATION` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`NOTIFICATION_INVITATION` находится в НФБК, потому что детерминант `{notification_id}` является суперключом.

---

## Relation NOTIFICATION_COLLECTION

Краткое описание: связь уведомления с подборкой.

### Functional dependencies

```text
Relation NOTIFICATION_COLLECTION:

{notification_id} -> collection_id
```

### Обоснование нормальных форм

`NOTIFICATION_COLLECTION` находится в 1НФ, потому что оба атрибута атомарны.

`NOTIFICATION_COLLECTION` находится во 2НФ, потому что ключ `notification_id` простой.

`NOTIFICATION_COLLECTION` находится в 3НФ, потому что транзитивные зависимости отсутствуют.

`NOTIFICATION_COLLECTION` находится в НФБК, потому что детерминант `{notification_id}` является суперключом.
