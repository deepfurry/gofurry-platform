# Local development

## Toolchain and commands

Use Go 1.27.1+ (one `server/go.mod`, no `go.work`), Node 24 LTS and pnpm 10.11.0.
Go module tools pin oapi-codegen/sqlc; Goose and River migrators use their pinned
official Go APIs. No global Go tools, psql or redis-cli are needed.

| Command | Purpose |
| --- | --- |
| `pnpm install --frozen-lockfile` | Install the committed workspace dependency graph |
| `pnpm generate` | OpenAPI → Fiber v3/TypeScript; Goose/queries → sqlc |
| `pnpm lint` | ESLint, gofmt verification and go vet |
| `pnpm typecheck` | Strict TypeScript and Astro diagnostics |
| `pnpm test` | Native Node generated-client tests and Go behavior tests |
| `pnpm build` | Public SSR, Admin SPA and all three Go runtime binaries |
| `pnpm check` | Secret/boundary audit, generation drift, lint/types/tests/build |
| `pnpm check:generated` | Regenerate and compare exact bytes and file membership |
| `pnpm audit:repository` | Secret, forbidden dependency and architecture checks |
| `pnpm migrate:dev` | Explicit Goose → official River → Worker object grants |
| `pnpm smoke:dev` | Real pgx/Redis and River execution checks with private config |
| `pnpm smoke:auth:dev` | Temporary account through real Public HTTP/application handlers, with fixture cleanup |
| `pnpm integration:ci` | Fresh, guarded loopback disposable PostgreSQL/Redis tests |
| `pnpm build:images` | Build four local Docker images, without publishing |

`pnpm check` never implicitly migrates or accesses private Infra. Run the separate
integration/image gates when the active spec requires them. Generated source is
committed and never edited manually. sqlc reads the Goose migrations directly;
there is no second schema definition. Tool-only transitive dependencies (for example
sqlc's MySQL/SQLite parsers and gRPC client) are not application architecture/runtime
integrations. No observability framework is configured or imported by application code.

## Private launch inputs

Preserve the existing ignored files:

```text
server/env/api.local
server/env/admin.local
server/env/worker.local
server/env/migrator.local
.local/readonly.env
```

The four `server/env/*.example` files show the public contract using localhost
placeholders. For a new workstation, obtain private credentials through the operator
path and create missing local files only. Never overwrite existing credentials.
Do not print, commit, log or paste these files or real addresses/connection URLs.
The readonly operator input is not an application launch input.

Runtime binaries read process environment only and validate required values. Node's
standard `parseEnv` loads `.local` only in the developer launcher, which refuses CI.
For older prepared files, it supplies missing non-secret defaults: API
`127.0.0.1:8080`, Admin API `127.0.0.1:8081`, Worker/Migrator `RIVER_SCHEMA=river`.
Explicit values win and are still validated. No private file is rewritten.

Start separate terminals:

```text
pnpm dev:api
pnpm dev:admin-api
pnpm dev:worker
pnpm dev:web
pnpm dev:admin
```

Web development serves on port 4321; Admin serves on 4322. Development `/api/*`
proxies strip `/api` and forward to the corresponding Go API. Browser clients use
the generated facade. Both APIs expose GET `/health/live` and `/health/ready`.
Live does not fan out; readiness requires PostgreSQL and reports Redis degradation.
Anonymous public SSR never reads per-user state; `/foundation` is prerendered.

## Local authentication (P0-1A)

Public Web provides `/register`, `/login` and `/account` as anonymous Astro shells
with React islands. Account data is fetched in the browser through `/api/me`; it
never enters shared SSR HTML. The public profile lookup is GET `/api/users/{handle}`.
There is no public profile HTML route yet. Admin remains health-only.

Public API adds POST `/auth/register`, `/auth/login`, `/auth/logout`, GET `/me`,
PATCH `/me/profile` and GET `/users/{handle}`. PATCH retains omitted fields; null
clears handle/display name/bio. Indexing is an explicit boolean and defaults false.
Public profile JSON contains only handle, display name and bio.

API configuration includes `PUBLIC_ORIGIN`. Development defaults to
`http://localhost:4321`; tests supply an explicit value. Production requires an
HTTPS origin without path/query/fragment/userinfo. Unsafe auth/profile requests
must carry that exact Origin. Do not configure broad CORS. Existing local files
need no edits to use the development default. Other processes do not need it.

easyhash v1.2.0 hashes new passwords with explicit `WithArgon2id()`: 64 MiB memory,
time cost 3, parallelism 2, library-default salt/key lengths. Its default Hash and
DefaultPolicy prefer bcrypt, so GoFurry overrides the policy to Argon2id before
`VerifyAndUpgrade`. Successful legacy/bcrypt verification upgrades with a CAS;
concurrent upgrades retry verification once. Equal/stronger current Argon2id hashes
are retained. Rehash does not change the actual password-change timestamp.
Unknown identities perform dummy Argon2id verification. Passwords accept 15–128
Unicode code points, including spaces, without trimming or normalization.

Business IDs use Go 1.27's standard `uuid.NewV7()`. Browser sessions contain 32
random bytes encoded with unpadded base64url; PostgreSQL stores only the SHA-256
hash of that encoded token. Public sessions expire after 30 days absolute or 14
days idle. Activity is touched at most every 10 minutes and never extends absolute
expiry. Revoked/expired sessions and disabled/deleted accounts are rejected.
Production uses `__Host-gofurry_session`, Secure, HttpOnly, SameSite=Lax, Path=/,
without Domain. Local HTTP uses `gofurry_session`. Login always issues a new session;
logout revokes the current session and expires the same cookie. No browser storage
contains credentials or session tokens. Auth/profile database work has a bounded
request context; private responses use `Cache-Control: no-store`.

This is not production-auth complete. P0-1B supplies verification/recovery and full
synchronizer CSRF; P0-1C/P0-1D cover the remaining planned provider/security work.
Rate limits, Turnstile, security events, Admin auth and roles are not implemented.

## Migrations and shared Infra

Shared development Infra is accessed only using private configuration. No SSH,
server/container administration, cluster-role changes or Redis ACL changes are part
of ordinary work. Migrator may apply repository-owned changes only to `gfp_dev`.

`pnpm migrate:dev` checks the migrator identity/database, ensures Goose's bookkeeping
namespace `app`, runs Goose with `app.goose_db_version`, then official `rivermigrate`
with `Schema: river`, and grants Worker only pinned River runtime object access.
Goose owns namespace/extension foundation; River owns all SQL inside `river`.
Migration 1 retains pre-existing schemas/extensions on down; recovery uses new
forward migrations. Once applied to shared Infra, do not edit it in place.
Migration 2 adds the five identity/auth tables with restrictive FKs, CHECK/UNIQUE
constraints and explicit API DML. Admin/Worker get no identity DML. Always pass
disposable migration and auth tests before applying new migrations to shared dev.

River 0.47.0 refuses to start with zero registered workers (`client.go`, `Start`).
The only P0-0 job is `infrastructure.probe.v1`: an explicit smoke request that runs
the system readiness query. Normal worker startup does not enqueue or schedule it.
Smoke waits for completion, removes its own job, and checks bounded Worker shutdown.
Public/Admin have no River enqueue grants. No future product jobs are scaffolded.

The official paths used are [oapi-codegen Fiber v3](https://github.com/oapi-codegen/oapi-codegen/blob/v2.8.0/docs/fiber-v3-server.md)
and [River explicit alternate schema](https://riverqueue.com/docs/alternate-schema).

`pnpm smoke:dev` verifies all four prepared PostgreSQL identities, schema/extension
state and sqlc query, Redis PING and a unique TTL-bound `gfp:*` SET/GET/DEL, then
Worker River enqueue/execution/completion/cleanup. Driver errors retain only safe
operation labels and PostgreSQL SQLSTATE. Missing privileges are a stop condition.

After disposable acceptance, `pnpm smoke:auth:dev` checks `gfp_api`/`gfp_dev`, then
registers, logs in, reads/updates a profile and logs out through the actual Fiber
handlers/application code with real pgx connections. It does not require an already
running API listener. No email/password/cookie/hash/URL is printed. The prepared
`gfp_migrator` connection deletes only that process's randomly named temporary
identity and dependent records in a restrictive-FK-safe transaction; no runtime
DELETE grant, shared server administration or production user deletion is involved.

## Disposable CI and containers

GitHub Actions provisions fresh PostgreSQL 18 and Redis 8 service containers. It
never loads private files or uses Tailscale. `integration:ci` requires `CI=true` and
`GFP_DISPOSABLE_INFRA=1`, uses fixed loopback endpoints and creates `gfp_ci` plus
restricted roles. Its published passwords are disposable fixtures, not developer
credentials. Running it against an existing initialized CI database fails deliberately;
use fresh containers for each acceptance run. Only this explicitly guarded fixture
setup creates cluster roles, on the disposable server.

CI reuses `pnpm check`, repeats generation with Git drift/untracked-file checks,
runs fresh migrations twice, driver smoke and P0-1A auth/database/HTTP/privacy tests,
then builds all four images. Third
party Actions are pinned to commit SHAs. No deployment, tag, release or image push.

See [deploy/README.md](../deploy/README.md) for images and runtime ports. Local image
acceptance needs a working Docker engine; lack of Docker must be reported as an
unexecuted gate, never a passing build.

To rerun only auth integration against an already initialized disposable fixture,
set `CI=true`, `GFP_DISPOSABLE_INFRA=1`, `GFP_AUTH_INTEGRATION=1` and run
`go -C server test -count=1 -run TestIntegration ./internal/transport/public`.
Tests hard-code the disposable loopback database and never read developer URLs.
The ordinary Go test suite skips these integration tests until explicitly enabled.
