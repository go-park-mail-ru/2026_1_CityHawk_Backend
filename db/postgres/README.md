# PostgreSQL runtime config

`postgresql.conf` is mounted by `docker-compose.yml` and contains the server
settings required for the database administration homework.

Chosen values:

- `max_connections = 100`: six backend processes use `DB_POOL_MAX_CONNS=10` by
  default, so the application can consume about 60 connections. The rest is left
  for migrations, monitoring and manual maintenance.
- `listen_addresses = '*'`: inside Docker the server must accept connections
  from other containers. The host port is bound to `127.0.0.1` in Compose to
  avoid exposing PostgreSQL on all host interfaces during local development.
- `log_min_duration_statement = '1000ms'`: requests above one second are already
  suspicious for the CityHawk API UX and useful for DOS/noisy-query detection.
- `pg_stat_statements` and `auto_explain` are preloaded so query statistics and
  execution plans are available after restart.
- `log_line_prefix` includes timestamp, pid, user, database, application name
  and client address so the logs can be parsed by PGBadger.

PostgreSQL writes logs to the `log` directory inside `PGDATA`; this avoids file
permission issues with extra local bind mounts.
