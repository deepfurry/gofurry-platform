package database

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDriverErrorsAreRedacted(t *testing.T) {
	for _, err := range []error{errors.New("sentinel-private-address/password"), &pgconn.PgError{Code: "42501", Message: "sentinel-private-address/password"}} {
		got := SafeError("operation", err).Error()
		if strings.Contains(got, "sentinel") {
			t.Fatal("driver text leaked")
		}
	}
	if !strings.Contains(SafeError("grant", &pgconn.PgError{Code: "42501"}).Error(), "42501") {
		t.Fatal("SQLSTATE lost")
	}
	if !strings.Contains(SafeError("connect", context.DeadlineExceeded).Error(), "timed out") {
		t.Fatal("timeout classification lost")
	}
}
