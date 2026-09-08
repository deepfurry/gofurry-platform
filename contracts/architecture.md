# Architecture contract

- One Go module, three runtime processes: Public API, Admin API, Worker.
- Public/Admin remain separate transport/security boundaries; no inter-process HTTP
  calls to share business behavior. Composition belongs in `cmd/*`.
- Future Application owns transactions; Domain stays ordinary Go without transport
  DTOs, Fiber, Redis or River. No ceremonial service/repository wrappers around sqlc.
- River imports and types stay under `server/internal/jobs`.
- PostgreSQL is canonical; Redis is ephemeral and all created keys use `gfp:`.
- OpenAPI is spec-first. P0-0 exposes only GET `/health/live` and `/health/ready`.
  Liveness never probes dependencies. Readiness requires PostgreSQL; Redis failure
  reports degraded state while PostgreSQL remains ready.
- Go 1.27+, Node 24, pnpm workspace; Astro/React 19 public SSR, React 19/Vite admin.
  Tailwind v4 handles layout; SCSS handles appearance.
- No product packages/tables, unused dependencies, ORM/AutoMigrate, MongoDB, NATS,
  vector support, extra brokers/search services or observability stack in P0-0.
