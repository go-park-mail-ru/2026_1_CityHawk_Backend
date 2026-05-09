DROP TRIGGER IF EXISTS set_organizer_application_updated_at ON organizer_application;
DROP INDEX IF EXISTS idx_organizer_application_status_created_at;
DROP INDEX IF EXISTS organizer_application_active_user_key;
DROP TABLE IF EXISTS organizer_application;
