-- +goose Up
CREATE TABLE app.users (
    id uuid PRIMARY KEY,
    account_state text NOT NULL DEFAULT 'active' CHECK (account_state IN ('active', 'disabled')),
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz
);

CREATE TABLE app.user_profiles (
    user_id uuid PRIMARY KEY REFERENCES app.users(id) ON DELETE RESTRICT,
    handle text UNIQUE CHECK (handle IS NULL OR (handle = lower(handle) AND handle ~ '^[a-z0-9][a-z0-9_-]{2,31}$')),
    display_name text CHECK (char_length(display_name) <= 80),
    bio text CHECK (char_length(bio) <= 500),
    search_engine_indexing boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE app.auth_identities (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE RESTRICT,
    provider text NOT NULL CHECK (provider IN ('email', 'google', 'github')),
    provider_subject text NOT NULL,
    email text,
    verified_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (provider, provider_subject)
);
CREATE INDEX auth_identities_user_id_idx ON app.auth_identities(user_id);

CREATE TABLE app.password_credentials (
    user_id uuid PRIMARY KEY REFERENCES app.users(id) ON DELETE RESTRICT,
    password_hash text NOT NULL,
    password_updated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE app.sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE RESTRICT,
    kind text NOT NULL CHECK (kind IN ('public', 'admin')),
    auth_method text NOT NULL,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    authenticated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    absolute_expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);
CREATE INDEX sessions_user_id_idx ON app.sessions(user_id);
CREATE INDEX sessions_active_user_idx ON app.sessions(user_id, kind) WHERE revoked_at IS NULL;
CREATE INDEX sessions_expiry_idx ON app.sessions(LEAST(idle_expires_at, absolute_expires_at)) WHERE revoked_at IS NULL;

-- Override any prepared default privileges; runtime roles never own these objects.
REVOKE ALL ON app.users, app.user_profiles, app.auth_identities, app.password_credentials, app.sessions
    FROM PUBLIC, gfp_api, gfp_admin, gfp_worker, gfp_readonly;
GRANT USAGE ON SCHEMA app TO gfp_api, gfp_readonly;
GRANT SELECT, INSERT ON app.users, app.auth_identities TO gfp_api;
GRANT SELECT, INSERT, UPDATE ON app.user_profiles TO gfp_api;
GRANT SELECT, INSERT ON app.password_credentials, app.sessions TO gfp_api;
GRANT UPDATE (password_hash, password_updated_at, updated_at) ON app.password_credentials TO gfp_api;
GRANT UPDATE (last_seen_at, idle_expires_at, revoked_at) ON app.sessions TO gfp_api;
GRANT SELECT ON app.users, app.user_profiles, app.auth_identities, app.password_credentials, app.sessions TO gfp_readonly;

-- +goose Down
DROP TABLE app.sessions;
DROP TABLE app.password_credentials;
DROP TABLE app.auth_identities;
DROP TABLE app.user_profiles;
DROP TABLE app.users;
