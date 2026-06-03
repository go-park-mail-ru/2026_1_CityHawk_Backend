# DB administration homework notes

## Service accounts

Application services use `cityhawk_app`, not the PostgreSQL bootstrap/admin
account. The role script is stored in `db/admin/001_create_service_roles.sql`
because it prepares operational database roles and is not part of the ordinary
application schema.

Roles:

- `cityhawk_admin`: local Docker bootstrap role from `POSTGRES_USER`; used to
  create the database on first container initialization.
- `cityhawk_migrator`: migration role with schema creation rights.
- `cityhawk_app`: runtime service account with DML privileges on application
  tables, sequence usage and function execution.
- `cityhawk_monitor`: monitoring role with `pg_monitor` for
  `postgres_exporter`.

## Connection pool and max_connections

The Go services use `pgxpool`. Pool settings are read from env and applied in
`internal/platform/postgres/postgres.go`.

Default local values:

```env
DB_POOL_MAX_CONNS=10
DB_POOL_MIN_CONNS=1
DB_POOL_MAX_CONN_LIFETIME=30m
DB_POOL_MAX_CONN_IDLE_TIME=5m
```

There are six database-using processes in Docker Compose: the HTTP backend and
five gRPC services. With `DB_POOL_MAX_CONNS=10`, the application can occupy up
to roughly 60 PostgreSQL connections. PostgreSQL has `max_connections = 100`,
leaving about 40 connections for migrations, monitoring, manual `psql` sessions
and short operational bursts.

If services are scaled, the value should be recalculated as:

```text
max_connections >= sum(service_replicas * DB_POOL_MAX_CONNS) + admin_margin
```

## listen_addresses

`db/postgres/postgresql.conf` uses:

```conf
listen_addresses = '*'
```

Inside Docker, PostgreSQL must listen on the container network so backend
containers and `postgres_exporter` can connect. For local host security,
`docker-compose.yml` binds the host port to `127.0.0.1:${DB_PORT:-5432}:5432`,
so the database is not exposed on every host interface.

## Client timeouts

The app sets PostgreSQL runtime parameters on each pool connection:

```env
DB_STATEMENT_TIMEOUT=5s
DB_LOCK_TIMEOUT=1s
```

`statement_timeout=5s` protects the API from unexpectedly long database work.
For CityHawk endpoints, responses above a few seconds are already poor UX, and
this also limits the impact of accidental heavy queries or simple DOS attempts.

`lock_timeout=1s` avoids letting write requests wait too long behind blocked
transactions. It favors returning an error over tying up request workers and
database connections.

## Slow query logging and PGBadger

The mounted PostgreSQL config enables:

```conf
logging_collector = on
log_min_duration_statement = '1000ms'
log_line_prefix = '%m [%p] %q%u@%d app=%a client=%h '
```

Queries slower than one second are logged. The prefix includes timestamp, pid,
user, database, application name and client address, which gives PGBadger enough
context to build reports. PostgreSQL writes logs into `PGDATA/log`.

Example report command inside or against copied logs:

```bash
pgbadger /path/to/postgresql-*.log -o pgbadger-report.html
```

## pg_stat_statements and auto_explain

`db/postgres/postgresql.conf` preloads:

```conf
shared_preload_libraries = 'pg_stat_statements,auto_explain'
```

`pg_stat_statements` is created by migration
`db/migrations/0029_pg_stat_statements.up.sql`. `auto_explain` logs execution
plans for queries slower than one second:

```conf
auto_explain.log_min_duration = '1000ms'
auto_explain.log_analyze = on
auto_explain.log_buffers = on
auto_explain.log_nested_statements = on
```

## Monitoring

Docker Compose includes `postgres-exporter` in the `monitoring` profile.
Prometheus scrapes it with the `postgres-exporter` job, and the Grafana dashboard
contains panels for:

- PostgreSQL connections by state.
- Transactions per second.
- Locks and deadlocks.
- Database size.
- Cache hit ratio.
- Prometheus target health including `postgres-exporter`.

Start monitoring with:

```bash
docker compose --profile monitoring up -d
```

Grafana is available at `http://localhost:3000`, Prometheus at
`http://localhost:9090`.
