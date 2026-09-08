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
