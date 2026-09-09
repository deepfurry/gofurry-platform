-- +goose Up
ALTER TABLE app.security_events DROP CONSTRAINT security_events_event_type_check;
ALTER TABLE app.security_events ADD CONSTRAINT security_events_event_type_check CHECK (event_type IN (
    'account_registered', 'login_succeeded', 'logout',
    'email_verification_requested', 'email_verified',
    'password_reset_requested', 'password_reset_completed', 'password_changed',
    'reauthenticated', 'session_revoked', 'other_sessions_revoked',
    'oauth_login_succeeded', 'oauth_identity_linked', 'oauth_identity_unlinked', 'oauth_reauthenticated'
));
ALTER TABLE app.sessions ADD CONSTRAINT sessions_auth_method_check
    CHECK (auth_method IN ('password', 'password_reset', 'google', 'github'));
CREATE UNIQUE INDEX auth_identities_user_provider_idx ON app.auth_identities(user_id, provider)
    WHERE provider IN ('google', 'github');

-- SELECT FOR UPDATE requires an UPDATE privilege, even when no column changes.
GRANT UPDATE (updated_at) ON app.users TO gfp_api;
GRANT UPDATE (email) ON app.auth_identities TO gfp_api;
GRANT DELETE ON app.auth_identities TO gfp_api;

-- +goose Down
REVOKE DELETE ON app.auth_identities FROM gfp_api;
REVOKE UPDATE (email) ON app.auth_identities FROM gfp_api;
REVOKE UPDATE (updated_at) ON app.users FROM gfp_api;
DROP INDEX app.auth_identities_user_provider_idx;
ALTER TABLE app.sessions DROP CONSTRAINT sessions_auth_method_check;
-- Down intentionally refuses to erase OAuth security history if rows exist.
ALTER TABLE app.security_events DROP CONSTRAINT security_events_event_type_check;
ALTER TABLE app.security_events ADD CONSTRAINT security_events_event_type_check CHECK (event_type IN (
    'account_registered', 'login_succeeded', 'logout',
    'email_verification_requested', 'email_verified',
    'password_reset_requested', 'password_reset_completed', 'password_changed',
    'reauthenticated', 'session_revoked', 'other_sessions_revoked'
));
