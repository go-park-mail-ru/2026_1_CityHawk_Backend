-- Baseline DDL snapshot for HW 4 before performance optimization migration 0030.
-- Run from the repository root with:
-- psql "$DATABASE_URL" -f perf_test/init.sql
--
-- This file intentionally stops before db/migrations/0030_perf_event_indexes.up.sql.
-- It preserves the original database definition used for the baseline load test.

\ir ../db/migrations/0001_init.up.sql
\ir ../db/migrations/0003_search_trgm.up.sql
\ir ../db/migrations/0004_media_paths.up.sql
\ir ../db/migrations/0005_support_tickets.up.sql
\ir ../db/migrations/0006_support_ticket_messages.up.sql
\ir ../db/migrations/0007_user_roles.up.sql
\ir ../db/migrations/0008_profile_social_schema.up.sql
\ir ../db/migrations/0009_organizer_applications.up.sql
\ir ../db/migrations/0010_invitations_notifications_share_links.up.sql
\ir ../db/migrations/0011_event_place.up.sql
\ir ../db/migrations/0020_optional_profile_names.up.sql
\ir ../db/migrations/0026_tag_groups.up.sql
\ir ../db/migrations/0029_pg_stat_statements.up.sql
