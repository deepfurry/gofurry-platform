-- name: LockUserAuthState :one
SELECT id FROM app.users WHERE id=$1 AND account_state='active' AND deleted_at IS NULL FOR UPDATE;

-- name: FindProviderIdentity :one
SELECT id, user_id, provider, provider_subject, email, created_at FROM app.auth_identities
WHERE provider=sqlc.arg(provider) AND provider_subject=sqlc.arg(subject) AND provider IN ('google','github');

-- name: FindProviderIdentityForUpdate :one
SELECT id, user_id, provider, provider_subject, email, created_at FROM app.auth_identities
WHERE provider=sqlc.arg(provider) AND provider_subject=sqlc.arg(subject) AND provider IN ('google','github') FOR UPDATE;

-- name: ListProviderIdentitiesForUser :many
SELECT id, provider, provider_subject, email, created_at FROM app.auth_identities
WHERE user_id=$1 AND provider IN ('google','github') ORDER BY provider;

-- name: CreateProviderIdentity :exec
INSERT INTO app.auth_identities(id,user_id,provider,provider_subject,email,verified_at,created_at,updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$7);

-- name: UpdateProviderIdentityMetadata :exec
UPDATE app.auth_identities SET email=sqlc.narg(email),updated_at=sqlc.arg(now)
WHERE id=sqlc.arg(id) AND user_id=sqlc.arg(user_id) AND provider IN ('google','github');

-- name: DeleteProviderIdentity :execrows
DELETE FROM app.auth_identities WHERE user_id=$1 AND provider=$2 AND provider IN ('google','github');

-- name: FindEmailIdentityBySubject :one
SELECT user_id FROM app.auth_identities WHERE provider='email' AND provider_subject=$1;

-- name: CreateEmailIdentityForOAuthAccount :exec
INSERT INTO app.auth_identities(id,user_id,provider,provider_subject,email,verified_at,created_at,updated_at)
VALUES ($1,$2,'email',sqlc.arg(email)::text,sqlc.arg(email)::text,sqlc.narg(verified_at),sqlc.arg(now),sqlc.arg(now));

-- name: GetPasswordChangeTime :one
SELECT password_updated_at FROM app.password_credentials WHERE user_id=$1;

-- name: HasPasswordCredential :one
SELECT EXISTS(SELECT 1 FROM app.password_credentials WHERE user_id=$1);

-- name: InitializeOAuthProfile :exec
UPDATE app.user_profiles SET display_name=sqlc.narg(display_name) WHERE user_id=$1;

-- name: RevokeSessionsByAuthMethod :exec
UPDATE app.sessions SET revoked_at=sqlc.arg(now)
WHERE user_id=sqlc.arg(user_id) AND kind='public' AND auth_method=sqlc.arg(auth_method) AND revoked_at IS NULL;
