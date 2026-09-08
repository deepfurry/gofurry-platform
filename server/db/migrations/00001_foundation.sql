-- +goose Up
CREATE SCHEMA IF NOT EXISTS app;
CREATE SCHEMA IF NOT EXISTS river;
CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;

-- +goose Down
-- Deliberately retain namespaces/extensions that may predate this repository.
-- Foundation rollback is a forward-fix operation, never DROP SCHEMA CASCADE.
SELECT 1;
