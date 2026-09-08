# Runtime artifacts

Run `pnpm build:images` from the repository root with a local Docker engine. Builds
are local only and never push. CI builds the same four Dockerfiles.

| Image | Runtime |
| --- | --- |
| `gofurry-server:p0-0-local` | `gofurry-api` (default), `gofurry-admin`, or `gofurry-worker` |
| `gofurry-web:p0-0-local` | Astro Node SSR, port 4321 |
| `gofurry-admin:p0-0-local` | Unprivileged Nginx SPA, port 8080 |
| `gofurry-postgres:p0-0-local` | PostgreSQL 18 with pg_trgm available |

All backend binaries read injected process environment. Images contain no private
config. `.dockerignore` excludes local secrets and generated build artifacts.
Server/Web run as non-root; Admin uses an unprivileged runtime. PostgreSQL follows
its upstream entrypoint and keeps PGDATA independent from image contents.
Web uses tini to forward container signals to Node and reap child processes.

The server image's command selects a runtime binary; API and Admin require separate
HTTP_ADDR values when sharing a network namespace. Migrations are explicit developer
or CI commands and never run at application startup.

P0-0 defines build artifacts only. Same-origin `/api/*` forwarding belongs to a future
deployment reverse proxy; the Admin static container returns 503 on that prefix until
routing is configured. No VPS, Cloudflare, certificates, production credentials,
image publishing or deployment is provisioned here.
