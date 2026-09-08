# GoFurry International — Authentication & Security Architecture

## Authentication Model

```text
Email + Password
Google OIDC
GitHub OAuth
       │
       ▼
auth_identities
       │
       ▼
User Account
       │
       ▼
Public Profile
```

Authentication identity is private and separate from public profile identity.

Implemented P0-1A/B scope is local authentication, basic profiles, public sessions,
email verification and password recovery. Provider/Admin/abuse-control sections below
describe later P0-1C/D work, not capabilities already deployed.

## OAuth

Go backend owns OAuth flows.

Use Authorization Code flows with state and PKCE where applicable.

Temporary OAuth flow state belongs in Redis with short TTL and one-time consumption.

Do not retain provider access/refresh tokens unless a future feature explicitly requires provider API access.

## Account Linking

Do not auto-link accounts merely because provider emails match.

Link/unlink requires explicit user intent and re-authentication.

Do not unlink the last authentication method.

## Email / Password

Store passwords with:

```text
Argon2id
```

Favor length and password-manager compatibility over composition rules.

Email verification is required before high-cost contribution actions.

## Sessions

Use opaque server-side sessions.

Cookies:

```text
__Host-gofurry_session
__Host-gofurry_admin_session
```

Properties:

```text
Secure
HttpOnly
Path=/
No Domain
```

Public:

```text
SameSite=Lax
```

Admin:

```text
SameSite=Strict
```

Database stores only a hash of the raw session token.

PostgreSQL is the canonical session store.

## Lifetimes

Initial configurable defaults:

### Public
```text
absolute ≈ 30 days
idle     ≈ 14 days
```

### Admin
```text
absolute ≈ 8 hours
idle     ≈ 1 hour
```

Rotate sessions after login, password reset/change, recovery, or major auth changes.

## CSRF

Cookie-authenticated unsafe methods require:

```text
session-bound CSRF token
+
Origin verification
```

P0-1B uses base64url HMAC-SHA256(`CSRF_SECRET`, raw session token). GET `/auth/csrf`
returns only that HMAC with no-store; raw session tokens stay HttpOnly. Every
authenticated unsafe route checks exact Origin and `X-CSRF-Token`. Anonymous unsafe
routes check Origin only. Production requires a private secret of at least 32 bytes.
Browser mutation helpers request a fresh value each time; session rotation makes old
values invalid. CSRF never replaces PostgreSQL authentication or persists in storage.

## Auth Challenges

Email verification, password reset, and email change may use a unified challenge model.

Store only token hashes.

Tokens are random, short-lived, and single-use.

P0-1B implements only `email_verify` and `password_reset`. easyhash v1.2.0 supplies
GenerateToken, HashToken and VerifyToken; deterministic self-described SHA-256 hashes
support indexed lookup. Verification lasts 24 hours; reset lasts 30 minutes. A partial
unique index permits one outstanding identity/purpose challenge. Reissue first
invalidates the prior row and observes a 60-second cooldown. Consumption rechecks
purpose, state, expiry and active account inside the transaction.

Credential-row locking serializes login, challenge issuance/consumption and session
mutations; password KDFs run outside transactions. Reset/change revoke all public
sessions and create one replacement atomically. Reset does not set email verified.
Reauthentication verifies the current password, replaces just the current session
and refreshes `authenticated_at`. Foreign/nonexistent session IDs return the same 404.

Challenge delivery is an explicit post-commit exception to general queued mail:
auth owns the mail interface; `internal/mail` writes private local captures. Raw tokens
never enter PostgreSQL, Redis, River, logs or normal API responses. Capture links use
fragments, browser pages immediately clear them, and Referrer-Policy is no-referrer.
Registration keeps its committed account/session even on mail failure. Local capture
is rejected in production; no production provider is included in P0-1B.

## Enumeration Resistance

Login and password-reset flows do not reveal account existence unnecessarily.

Reset requests perform bounded dummy Argon2id work and return the same 202/body for
existing, absent, OAuth-only, disabled and deleted accounts and mail failure. Delivery
failures do not serialize sender errors. Current login behavior and registration
duplicate-email behavior remain as specified in P0-1A.

Rate-limit by multiple dimensions rather than IP alone.

## Turnstile

Use selectively for:

- registration
- password reset
- repeated failed login
- suspicious Resource submission
- spam-risk reports

Do not challenge Save / Want / Have.

## Public / Admin Separation

Admin:

```text
Cloudflare Access
+ mandatory MFA
+ GoFurry Admin Login
+ separate Admin Session
+ GoFurry Admin Authorization
```

Cloudflare identity is not GoFurry authorization.

Public sessions are never accepted as Admin sessions.

## Re-authentication

Require step-up re-authentication for sensitive account changes.

## Security Events

Security Events are separate from business Audit logs.

Examples:

```text
login_success
login_failure
oauth_failure
password_changed
password_reset
provider_linked
provider_unlinked
session_revoked
admin_login
challenge_failed
```

The implemented P0-1B event types are `account_registered`, `login_succeeded`,
`logout`, `email_verification_requested`, `email_verified`, `password_reset_requested`,
`password_reset_completed`, `password_changed`, `reauthenticated`, `session_revoked`
and `other_sessions_revoked`. The broader list above describes future events.
`app.security_events` allows only a generated bigint ID, user ID, historical session
ID, closed event type and timestamp. No metadata/email/IP/user-agent field exists.
Event insertion commits in the same transaction as the corresponding state change.

Do not log passwords, hashes, raw session tokens, CSRF tokens, OAuth codes/tokens, PKCE verifiers, reset tokens, cookies, or authorization headers.

## CORS

Browser API is same-origin:

```text
gofurry.com/api/*
admin.gofurry.com/api/*
```

Do not open broad CORS in P0.

## Proxy Trust

Go trusts forwarding/IP headers only through the controlled Cloudflare → Nginx path.

## Secrets

Use environment/Docker-secret style deployment.

No Vault requirement for P0.
