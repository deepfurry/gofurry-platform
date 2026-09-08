package auth

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/deepfurry/gofurry-platform/server/internal/database"
	"github.com/deepfurry/gofurry-platform/server/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrChallengeInvalid = errors.New("challenge invalid")
	ErrReauthFailed     = errors.New("reauthentication failed")
	ErrSessionNotFound  = errors.New("session not found")
	ErrMailUnavailable  = errors.New("challenge delivery unavailable")
)

type eventType string

const (
	accountRegistered     eventType = "account_registered"
	loginSucceeded        eventType = "login_succeeded"
	loggedOut             eventType = "logout"
	verificationRequested eventType = "email_verification_requested"
	emailVerified         eventType = "email_verified"
	resetRequested        eventType = "password_reset_requested"
	resetCompleted        eventType = "password_reset_completed"
	passwordChanged       eventType = "password_changed"
	reauthenticated       eventType = "reauthenticated"
	sessionRevoked        eventType = "session_revoked"
	otherSessionsRevoked  eventType = "other_sessions_revoked"
)

// Events accept identifiers and a closed event type, never arbitrary metadata.
func recordEvent(ctx context.Context, q *sqlc.Queries, event eventType, userID, sessionID uuid.UUID, now time.Time) error {
	session := pgtype.UUID{Bytes: sessionID, Valid: sessionID != uuid.Nil()}
	err := q.InsertSecurityEvent(ctx, sqlc.InsertSecurityEventParams{UserID: dbID(userID), SessionID: session, EventType: string(event), OccurredAt: timestamp(now)})
	if err != nil {
		return database.SafeError("record security event", err)
	}
	return nil
}

// Every existing-account authentication mutation takes this lock first. It
// serializes login/rotation/revocation/challenges without holding it during KDFs
// or mail delivery. This also prevents an old-password login escaping a reset.
func lockCredential(ctx context.Context, q *sqlc.Queries, userID uuid.UUID) (string, error) {
	row, err := q.LockPasswordCredential(ctx, dbID(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUnauthenticated
	}
	if err != nil {
		return "", database.SafeError("lock authentication state", err)
	}
	return row.PasswordHash, nil
}

func requireActor(ctx context.Context, q *sqlc.Queries, actor Actor, now time.Time) error {
	if actor.SessionKind != "public" {
		return ErrUnauthenticated
	}
	_, err := q.GetActivePublicSessionByID(ctx, sqlc.GetActivePublicSessionByIDParams{ID: dbID(actor.SessionID), UserID: dbID(actor.UserID), Now: timestamp(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUnauthenticated
	}
	if err != nil {
		return database.SafeError("revalidate authentication", err)
	}
	return nil
}

type SessionView struct {
	ID                                                                       uuid.UUID
	AuthMethod                                                               string
	AuthenticatedAt, CreatedAt, LastSeenAt, IdleExpiresAt, AbsoluteExpiresAt time.Time
	Current                                                                  bool
}

func (a *App) Sessions(ctx context.Context, actor Actor) ([]SessionView, error) {
	q, now := sqlc.New(a.pool), a.now().UTC()
	if err := requireActor(ctx, q, actor, now); err != nil {
		return nil, err
	}
	rows, err := q.ListActivePublicSessions(ctx, sqlc.ListActivePublicSessionsParams{UserID: dbID(actor.UserID), Now: timestamp(now)})
	if err != nil {
		return nil, database.SafeError("list sessions", err)
	}
	result := make([]SessionView, 0, len(rows))
	for _, row := range rows {
		id := uuid.UUID(row.ID.Bytes)
		result = append(result, SessionView{ID: id, AuthMethod: row.AuthMethod,
			AuthenticatedAt: row.AuthenticatedAt.Time, CreatedAt: row.CreatedAt.Time, LastSeenAt: row.LastSeenAt.Time,
			IdleExpiresAt: row.IdleExpiresAt.Time, AbsoluteExpiresAt: row.AbsoluteExpiresAt.Time, Current: id == actor.SessionID})
	}
	return result, nil
}

func (a *App) RevokeSession(ctx context.Context, actor Actor, target uuid.UUID) error {
	return a.revoke(ctx, actor, target, sessionRevoked)
}
func (a *App) RevokeOthers(ctx context.Context, actor Actor) error {
	return a.revoke(ctx, actor, actor.SessionID, otherSessionsRevoked)
}
func (a *App) revoke(ctx context.Context, actor Actor, target uuid.UUID, event eventType) error {
	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return database.SafeError("begin session revocation", err)
	}
	defer tx.Rollback(ctx)
	q := sqlc.New(tx)
	if _, err = lockCredential(ctx, q, actor.UserID); err != nil {
		return err
	}
	now := a.now().UTC()
	if err = requireActor(ctx, q, actor, now); err != nil {
		return err
	}
	if event == otherSessionsRevoked {
		err = q.RevokeOtherPublicSessions(ctx, sqlc.RevokeOtherPublicSessionsParams{UserID: dbID(actor.UserID), CurrentID: dbID(actor.SessionID), Now: timestamp(now)})
	} else {
		var count int64
		count, err = q.RevokePublicSessionByID(ctx, sqlc.RevokePublicSessionByIDParams{ID: dbID(target), UserID: dbID(actor.UserID), Now: timestamp(now)})
		if err == nil && count != 1 {
			return ErrSessionNotFound
		}
	}
	if err != nil {
		return database.SafeError("revoke public sessions", err)
	}
	if err = recordEvent(ctx, q, event, actor.UserID, target, now); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return database.SafeError("commit session revocation", err)
	}
	return nil
}
