# Architecture

Public/Admin are distinct transport and security boundaries, sharing one Go module.

```text
apps/web / apps/admin → generated API clients → Public/Admin transport
                                                ↓
                                      future Application/Domain
                                                ↓
                                           sqlc / pgx
                                                ↓
                                           PostgreSQL
worker → jobs adapter → shared Application/Domain (when implemented)
```

P0-0 contains only health transport and infrastructure. No product domains or tables
are scaffolded. `cmd/*` composes configuration, dependencies, startup, signals and
bounded cleanup; reusable behavior lives in `internal/`.

Worker never calls Public/Admin HTTP. Future domain code must not import Fiber,
transport DTOs, Redis or River. River types stay inside Jobs infrastructure and its
objects live in `river`; business data will live in `app`. PostgreSQL is canonical,
Redis holds disposable state. No ORM, AutoMigrate, vectors, MongoDB or message broker.

OpenAPI owns Go transport and TypeScript clients. Goose migrations plus SQL queries
own sqlc output. Generated files are committed, reviewed, and never manually edited.

Public web uses anonymous Astro Node SSR with isolated React interaction; Admin is
a React SPA. Apps use API-client facades, never deep generated imports or direct DB
access. Tailwind v4 owns layout; shared SCSS and SCSS Modules own visual appearance.
