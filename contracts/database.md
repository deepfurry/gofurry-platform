# Database contract

PostgreSQL 18.x is the canonical source of truth. `app` is the application schema;
`river` is isolated job infrastructure. Redis is ephemeral, never canonical storage.

Goose SQL migrations under `server/db/migrations` are the sole application schema
source. No duplicate `schema.sql`, ORM, GORM or AutoMigrate. Never modify a released
or shared-environment-applied migration in place: add a new migration. Runtime
startup must not migrate. Explicit migration order is Goose then official River
`rivermigrate` using its explicit `Schema` option; never copy/rewrite River SQL.

The explicit migrator first ensures `app` exists so Goose can create
`app.goose_db_version` before applying migration 1. This bookkeeping bootstrap is
necessary on a fresh database and respects the prepared migrator's lack of CREATE
on `public`. All subsequent application schema evolution remains Goose-owned.

`pg_trgm` is required in `public`; `vector` must not be enabled. Migration 1 remains
immutable. Migration 2 owns identity, profile, credential and session tables in
`app`. Goose may create namespaces, but River owns objects inside `river`.

sqlc consumes Goose migrations and `server/db/queries`, generating committed pgx/v5
code under `server/internal/database/sqlc`. Never hand-edit generated code.

Development roles:

| Role | Purpose |
| --- | --- |
| `gfp_migrator` | Explicit Goose/River migrations and owned-object grants |
| `gfp_api` | Public API runtime |
| `gfp_admin` | Admin API runtime |
| `gfp_worker` | Worker and River runtime |
| `gfp_readonly` | Operator inspection only |

Only the migrator accesses migration machinery. Worker gets River object runtime
privileges, not ownership/DDL. Public/Admin enqueue grants wait until actually used.
Cluster roles and shared server configuration are operator-owned; missing privileges
are a stop condition, never grounds for SSH, superuser use or widening Redis ACLs.

P0-1A grants API only the SELECT/INSERT/UPDATE rights its use cases need; Admin and
Worker receive no identity DML. Readonly may SELECT. IDs are generated in Go with
standard-library UUIDv7. Restrictive FKs, CHECKs and UNIQUE constraints enforce the
account/identity/profile/session invariants; public views never SELECT credentials.
