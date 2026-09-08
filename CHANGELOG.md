# Changelog

## Unreleased

- Establish P0-0 Agent context, engineering contracts and implementation guidance.
- Add pnpm workspace with Astro/React public SSR, React/Vite admin SPA, shared SCSS
  design tokens, and Public/Admin generated fetch-client facades.
- Add Go API/Admin/Worker runtimes, environment validation, structured redacted
  logging, bounded health checks and graceful shutdown.
- Add separate OpenAPI contracts, Fiber v3/oapi-codegen and Orval generation,
  Goose foundation and pgx/sqlc readiness query.
- Isolate official River migrations and Worker runtime in the `river` schema,
  with an explicit infrastructure probe and least-privilege object grants.
- Add real driver smoke commands, generated drift and secret/boundary checks,
  disposable-infrastructure CI, and four runtime Dockerfile definitions.
