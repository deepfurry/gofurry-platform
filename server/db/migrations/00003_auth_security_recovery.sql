-- +goose Up
CREATE TABLE app.auth_challenges (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE RESTRICT,
    auth_identity_id uuid NOT NULL REFERENCES app.auth_identities(id) ON DELETE RESTRICT,
    purpose text NOT NULL CHECK (purpose IN ('email_verify', 'password_reset')),
    token_hash text NOT NULL UNIQUE,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL CHECK (expires_at > created_at),
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CHECK (NOT (consumed_at IS NOT NULL AND invalidated_at IS NOT NULL))
);
CREATE UNIQUE INDEX auth_challenges_active_identity_purpose_idx
    ON app.auth_challenges(auth_identity_id, purpose)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;
CREATE INDEX auth_challenges_user_id_idx ON app.auth_challenges(user_id);
CREATE INDEX auth_challenges_latest_identity_purpose_idx
    ON app.auth_challenges(auth_identity_id, purpose, created_at DESC);
CREATE INDEX auth_challenges_expiry_idx ON app.auth_challenges(expires_at)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE TABLE app.security_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id uuid REFERENCES app.users(id) ON DELETE RESTRICT,
    session_id uuid,
    event_type text NOT NULL CHECK (event_type IN (
        'account_registered', 'login_succeeded', 'logout',
        'email_verification_requested', 'email_verified',
        'password_reset_requested', 'password_reset_completed', 'password_changed',
        'reauthenticated', 'session_revoked', 'other_sessions_revoked'
    )),
    occurred_at timestamptz NOT NULL
);
CREATE INDEX security_events_user_id_idx ON app.security_events(user_id);

REVOKE ALL ON app.auth_challenges, app.security_events
    FROM PUBLIC, gfp_api, gfp_admin, gfp_worker, gfp_readonly;
REVOKE ALL ON SEQUENCE app.security_events_id_seq
    FROM PUBLIC, gfp_api, gfp_admin, gfp_worker, gfp_readonly;
GRANT SELECT, INSERT, UPDATE ON app.auth_challenges TO gfp_api;
GRANT INSERT ON app.security_events TO gfp_api;
-- GENERATED ALWAYS AS IDENTITY insertion does not require direct sequence access.
GRANT UPDATE (verified_at, updated_at) ON app.auth_identities TO gfp_api;
GRANT SELECT ON app.auth_challenges, app.security_events TO gfp_readonly;

-- +goose Down
REVOKE UPDATE (verified_at, updated_at) ON app.auth_identities FROM gfp_api;
DROP TABLE app.security_events;
DROP TABLE app.auth_challenges;
