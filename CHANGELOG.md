# Changelog

## Unreleased

- Add P0-1B migration 3 for single-use auth challenges and transactional security
  events, with minimal API privileges and no new dependency or product domain.
- Require session-bound HMAC CSRF and exact Origin for authenticated unsafe Public
  requests; keep raw session tokens HttpOnly and discard CSRF on rotation.
- Add email verification, password reset/change, reauthentication and owned public
  session listing/revocation, with atomic password/challenge/session/event updates.
- Add consumer-owned post-commit mail delivery and private local capture; production
  rejects local capture and no production mail provider is included.
- Add recovery/verification pages and account security controls using generated
  clients, memory-only fragment handling and no-referrer responses.
- Extend disposable security, privacy, concurrency and rollback tests, and the real
  development auth smoke with private capture and narrowly scoped fixture cleanup.

- Add P0-1A accounts, public profiles, email identities, password credentials and
  PostgreSQL public sessions through migration 2, with explicit runtime grants.
- Add standard-library UUIDv7 IDs, easyhash v1.2.0 explicit Argon2id hashing,
  dummy verification and race-safe login hash upgrades.
- Add Public register/login/logout/me/profile APIs, opaque HttpOnly cookies,
  bounded session resolution and exact-origin protection for unsafe requests.
- Add anonymous Astro login/register/account shells with React forms using the
  generated Public client; public profile responses exclude all private fields.
- Add disposable database/concurrency/HTTP/privacy tests and a real-development
  authentication smoke with narrowly scoped temporary-identity cleanup.
- Extend Agent contracts and secret/boundary audits for the implemented identity
  scope. OAuth and final auth hardening remain later phases.

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
