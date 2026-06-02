DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cityhawk_migrator') THEN
        CREATE ROLE cityhawk_migrator LOGIN PASSWORD 'cityhawk_migrator';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cityhawk_app') THEN
        CREATE ROLE cityhawk_app LOGIN PASSWORD 'cityhawk_app';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cityhawk_monitor') THEN
        CREATE ROLE cityhawk_monitor LOGIN PASSWORD 'cityhawk_monitor';
    END IF;
END
$$;

GRANT CONNECT ON DATABASE cityhawk TO cityhawk_migrator;
GRANT CREATE ON DATABASE cityhawk TO cityhawk_migrator;
GRANT CONNECT ON DATABASE cityhawk TO cityhawk_app;
GRANT CONNECT ON DATABASE cityhawk TO cityhawk_monitor;
GRANT pg_monitor TO cityhawk_monitor;

GRANT USAGE, CREATE ON SCHEMA public TO cityhawk_migrator;
GRANT USAGE ON SCHEMA public TO cityhawk_app;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO cityhawk_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cityhawk_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO cityhawk_app;

ALTER DEFAULT PRIVILEGES FOR ROLE cityhawk_migrator IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO cityhawk_app;

ALTER DEFAULT PRIVILEGES FOR ROLE cityhawk_migrator IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO cityhawk_app;

ALTER DEFAULT PRIVILEGES FOR ROLE cityhawk_migrator IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO cityhawk_app;
