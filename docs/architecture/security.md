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

## Auth Challenges

Email verification, password reset, and email change may use a unified challenge model.

Store only token hashes.

Tokens are random, short-lived, and single-use.

## Enumeration Resistance

Login and password-reset flows do not reveal account existence unnecessarily.

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
