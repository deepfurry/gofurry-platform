-- name: CreateEmailIdentity :exec
INSERT INTO app.auth_identities (id, user_id, provider, provider_subject, email, created_at, updated_at)
VALUES ($1, $2, 'email', sqlc.arg(email)::text, sqlc.arg(email)::text, sqlc.arg(now), sqlc.arg(now));

-- name: CreatePasswordCredential :exec
INSERT INTO app.password_credentials (user_id, password_hash, password_updated_at, created_at, updated_at)
VALUES ($1, $2, $3, $3, $3);

-- name: FindLocalCredentialByEmailSubject :one
SELECT u.id, u.account_state, u.deleted_at, c.password_hash
FROM app.auth_identities a JOIN app.users u ON u.id = a.user_id
JOIN app.password_credentials c ON c.user_id = u.id
WHERE a.provider = 'email' AND a.provider_subject = $1;

-- name: CompareAndSwapPasswordHash :execrows
UPDATE app.password_credentials SET password_hash = sqlc.arg(new_hash), updated_at = sqlc.arg(now)
WHERE user_id = sqlc.arg(user_id) AND password_hash = sqlc.arg(old_hash);

-- name: GetPasswordCredentialByUser :one
SELECT c.user_id, c.password_hash FROM app.password_credentials c
JOIN app.users u ON u.id = c.user_id
WHERE c.user_id = $1 AND u.account_state = 'active' AND u.deleted_at IS NULL;

-- name: LockPasswordCredential :one
SELECT c.user_id, c.password_hash FROM app.password_credentials c
JOIN app.users u ON u.id = c.user_id
WHERE c.user_id = $1 AND u.account_state = 'active' AND u.deleted_at IS NULL
FOR UPDATE OF c;

-- name: ChangePasswordHash :execrows
UPDATE app.password_credentials SET password_hash = sqlc.arg(new_hash),
    password_updated_at = sqlc.arg(now), updated_at = sqlc.arg(now)
WHERE user_id = sqlc.arg(user_id) AND password_hash = sqlc.arg(old_hash);

-- name: SetPasswordHashForReset :exec
UPDATE app.password_credentials SET password_hash = sqlc.arg(new_hash),
    password_updated_at = sqlc.arg(now), updated_at = sqlc.arg(now)
WHERE user_id = sqlc.arg(user_id);

-- name: GetLocalEmailIdentityForUser :one
SELECT a.id, a.user_id, a.email, a.verified_at
FROM app.auth_identities a JOIN app.users u ON u.id = a.user_id
WHERE a.user_id = $1 AND a.provider = 'email' AND a.email IS NOT NULL
  AND u.account_state = 'active' AND u.deleted_at IS NULL;

-- name: MarkEmailIdentityVerified :execrows
UPDATE app.auth_identities SET verified_at = sqlc.arg(now), updated_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id) AND provider = 'email' AND verified_at IS NULL;

-- name: InvalidateActiveChallenges :exec
UPDATE app.auth_challenges SET invalidated_at = sqlc.arg(now)
WHERE user_id = sqlc.arg(user_id) AND purpose = sqlc.arg(purpose)
  AND consumed_at IS NULL AND invalidated_at IS NULL;

-- name: CreateAuthChallenge :exec
INSERT INTO app.auth_challenges (id, user_id, auth_identity_id, purpose, token_hash, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetLatestChallenge :one
SELECT id, created_at FROM app.auth_challenges
WHERE auth_identity_id = $1 AND purpose = $2
ORDER BY created_at DESC LIMIT 1;

-- name: FindActiveChallengeByTokenHash :one
SELECT ch.id, ch.user_id, ch.auth_identity_id, ch.token_hash
FROM app.auth_challenges ch JOIN app.auth_identities a ON a.id = ch.auth_identity_id AND a.user_id = ch.user_id
JOIN app.users u ON u.id = ch.user_id
WHERE ch.token_hash = sqlc.arg(token_hash) AND ch.purpose = sqlc.arg(purpose)
  AND ch.consumed_at IS NULL AND ch.invalidated_at IS NULL AND ch.expires_at > sqlc.arg(now)
  AND a.provider = 'email' AND u.account_state = 'active' AND u.deleted_at IS NULL;

-- name: FindActiveChallengeByTokenHashForUpdate :one
SELECT ch.id, ch.user_id, ch.auth_identity_id, ch.token_hash
FROM app.auth_challenges ch JOIN app.auth_identities a ON a.id = ch.auth_identity_id AND a.user_id = ch.user_id
JOIN app.users u ON u.id = ch.user_id
WHERE ch.token_hash = sqlc.arg(token_hash) AND ch.purpose = sqlc.arg(purpose)
  AND ch.consumed_at IS NULL AND ch.invalidated_at IS NULL AND ch.expires_at > sqlc.arg(now)
  AND a.provider = 'email' AND u.account_state = 'active' AND u.deleted_at IS NULL
FOR UPDATE OF ch;

-- name: ConsumeChallenge :execrows
UPDATE app.auth_challenges SET consumed_at = sqlc.arg(now)
WHERE id = sqlc.arg(id) AND consumed_at IS NULL AND invalidated_at IS NULL AND expires_at > sqlc.arg(now);

-- name: InsertSecurityEvent :exec
INSERT INTO app.security_events (user_id, session_id, event_type, occurred_at) VALUES ($1, $2, $3, $4);
