// Package db holds the authoritative Goose migration files.
package db

import "embed"

// Migrations contains only application-owned Goose SQL, never River SQL.
//
//go:embed migrations/*.sql
var Migrations embed.FS
