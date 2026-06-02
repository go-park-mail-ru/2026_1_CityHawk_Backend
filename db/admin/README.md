# Database administration scripts

`001_create_service_roles.sql` lives here because it is not an application
schema migration. It prepares database roles used by the application runtime
and by schema migrations.

Roles:

- `cityhawk_migrator` owns DDL work: migrations, extensions, tables, indexes and
  functions.
- `cityhawk_app` is the service account used by the backend services. It has
  DML access to existing and future application tables, sequence usage and
  function execution, but no schema creation rights.
- `cityhawk_monitor` is the monitoring account used by `postgres_exporter`. It
  gets `pg_monitor`, but no application DML grants.

The default passwords are local-development values. Production deployments
should override them and apply the same grants with secret-managed passwords.
