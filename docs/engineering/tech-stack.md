# GoFurry International — Technology Stack

## Frontend

### Public
- Astro
- React 19 Islands
- Node 24 LTS
- Vite 8 / Rolldown
- Tailwind CSS v4 for layout
- SCSS / SCSS Modules for visual styling
- Base UI primitives
- Lucide icons
- React Hook Form + Zod where needed

### Admin
- React 19
- Vite SPA
- TanStack Router
- TanStack Query v5
- React state first
- Zustand only when justified
- React Hook Form + Zod
- Base UI
- Tailwind layout + SCSS visual styling

### Workspace
- pnpm
- pnpm workspace
- no Turbo/Nx initially

## Backend

- Go 1.27+
- Fiber v3
- OpenAPI contract-first
- oapi-codegen
- pgx/v5 + pgxpool
- sqlc
- goose SQL migrations
- go-redis/v9
- River OSS
- log/slog
- golang.org/x/oauth2
- Argon2id through `github.com/gofurry/easyhash` v1.2.0: explicit `WithArgon2id()`
  for new credentials and an explicit Argon2id `VerifyAndUpgrade` policy. The
  library's default bcrypt preference is not GoFurry's password policy.
- AWS SDK v2 S3 client

## Database

- PostgreSQL 18.x
- Go standard-library UUIDv7 generation
- PostgreSQL FTS
- `pg_trgm`
- pgvector later, only when semantic search/recommendation is justified

## Storage

```text
PostgreSQL → canonical state
Redis      → ephemeral state / cache / counters
R2 / S3    → media
River      → durable background jobs stored in PostgreSQL
```

## Infrastructure

- Hong Kong VPS preferred
- Tokyo fallback
- Docker Compose
- Host Nginx
- Cloudflare
- GHCR
- GitHub Actions

## Explicit Non-selections

P0/P1 do not depend on:

```text
MongoDB
NATS
Kubernetes
gRPC
Service Mesh
Kafka
RabbitMQ
Elasticsearch/OpenSearch
ClickHouse
Vault
Distributed Tracing platform
```

## Runtime Alternatives

### Deno
Preferred alternative if GoFurry later chooses to leave Node for Astro SSR.

### Bun
Re-evaluate when Astro/runtime integration risk is low enough.

Node 24 LTS remains the production baseline.
