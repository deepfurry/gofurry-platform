package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Capabilities struct {
	CreateSchema bool `json:"create_schema"`
	CreatePublic bool `json:"create_public"`
	AppExists    bool `json:"app_exists"`
	RiverExists  bool `json:"river_exists"`
	Trigram      bool `json:"pg_trgm"`
	Vector       bool `json:"vector"`
}

func Inspect(ctx context.Context, pool *pgxpool.Pool, role, database string) (Capabilities, error) {
	var actualRole, actualDatabase string
	if err := pool.QueryRow(ctx, "SELECT current_user, current_database()").Scan(&actualRole, &actualDatabase); err != nil {
		return Capabilities{}, SafeError("PostgreSQL identity check", err)
	}
	if actualRole != role || actualDatabase != database {
		return Capabilities{}, errors.New("PostgreSQL identity does not match expected role/database (values withheld)")
	}
	var c Capabilities
	err := pool.QueryRow(ctx, `SELECT
		has_database_privilege(current_user, current_database(), 'CREATE'),
		has_schema_privilege(current_user, 'public', 'CREATE'),
		EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'app'),
		EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'river'),
		EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_trgm'),
		EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')`).Scan(
		&c.CreateSchema, &c.CreatePublic, &c.AppExists, &c.RiverExists, &c.Trigram, &c.Vector)
	if err != nil {
		return Capabilities{}, SafeError("PostgreSQL capability check", err)
	}
	return c, nil
}
