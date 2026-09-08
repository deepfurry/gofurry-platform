-- name: CreateUser :exec
INSERT INTO app.users (id, created_at, updated_at) VALUES ($1, $2, $2);

-- name: CreateProfile :exec
INSERT INTO app.user_profiles (user_id, created_at, updated_at) VALUES ($1, $2, $2);

-- name: GetMe :one
SELECT u.id, u.account_state, u.created_at, p.handle, p.display_name, p.bio,
       p.search_engine_indexing, a.email, a.verified_at
FROM app.users u JOIN app.user_profiles p ON p.user_id = u.id
JOIN app.auth_identities a ON a.user_id = u.id AND a.provider = 'email'
WHERE u.id = $1 AND u.account_state = 'active' AND u.deleted_at IS NULL;

-- name: UpdateProfile :execrows
UPDATE app.user_profiles p SET
    handle = CASE WHEN sqlc.arg(set_handle)::boolean THEN sqlc.narg(handle)::text ELSE p.handle END,
    display_name = CASE WHEN sqlc.arg(set_display_name)::boolean THEN sqlc.narg(display_name)::text ELSE p.display_name END,
    bio = CASE WHEN sqlc.arg(set_bio)::boolean THEN sqlc.narg(bio)::text ELSE p.bio END,
    search_engine_indexing = CASE WHEN sqlc.arg(set_indexing)::boolean THEN sqlc.arg(search_engine_indexing)::boolean ELSE p.search_engine_indexing END,
    updated_at = sqlc.arg(now)::timestamptz
WHERE p.user_id = sqlc.arg(user_id) AND EXISTS (
    SELECT 1 FROM app.users u WHERE u.id = p.user_id AND u.account_state = 'active' AND u.deleted_at IS NULL
);

-- name: GetPublicProfileByHandle :one
SELECT p.handle, p.display_name, p.bio
FROM app.user_profiles p JOIN app.users u ON u.id = p.user_id
WHERE p.handle = $1 AND u.account_state = 'active' AND u.deleted_at IS NULL;
