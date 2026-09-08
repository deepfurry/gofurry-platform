// auth-smoke exercises the real Public HTTP/application path with a temporary
// identity. Runtime DML uses gfp_api; the migrator removes only this run's fixture.
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
	"uuid"

	"github.com/deepfurry/gofurry-platform/server/internal/auth"
	"github.com/deepfurry/gofurry-platform/server/internal/config"
	"github.com/deepfurry/gofurry-platform/server/internal/database"
	"github.com/deepfurry/gofurry-platform/server/internal/identity"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/health"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/public"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (result error) {
	if os.Getenv("CI") != "" {
		return errors.New("real development auth smoke refuses CI")
	}
	cfg, err := config.Load("api")
	if err != nil {
		return err
	}
	if cfg.Environment != "development" {
		return errors.New("auth smoke requires development environment")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	api, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer api.Close()
	if _, err = database.Inspect(ctx, api, "gfp_api", "gfp_dev"); err != nil {
		return err
	}
	cleanupPool, err := database.Open(ctx, os.Getenv("AUTH_SMOKE_CLEANUP_URL"))
	if err != nil {
		return err
	}
	defer cleanupPool.Close()
	if _, err = database.Inspect(ctx, cleanupPool, "gfp_migrator", "gfp_dev"); err != nil {
		return err
	}
	authentication, err := auth.New(api)
	if err != nil {
		return err
	}
	app := fiber.New()
	public.Register(app, health.New(func(context.Context) error { return nil }, func(context.Context) error { return nil }), authentication, identity.New(api), public.Options{Environment: cfg.Environment, PublicOrigin: cfg.PublicOrigin})
	suffix := strings.ReplaceAll(uuid.NewV7().String(), "-", "")
	email := "p01a-smoke-" + suffix + "@example.invalid"
	var random [32]byte
	rand.Read(random[:])
	password := base64.RawURLEncoding.EncodeToString(random[:])
	defer func() {
		if err := cleanup(cleanupPool, email); err != nil {
			result = errors.Join(result, err)
		}
	}()
	request := func(method, path string, body any, cookie *http.Cookie, status int) (map[string]any, *http.Cookie, error) {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, nil, errors.New("auth smoke request encoding failed")
		}
		req := httptest.NewRequestWithContext(ctx, method, path, bytes.NewReader(encoded))
		req.Header.Set("Origin", cfg.PublicOrigin)
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}
		response, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			return nil, nil, errors.New("auth smoke HTTP request failed (details withheld)")
		}
		defer response.Body.Close()
		if response.StatusCode != status {
			return nil, nil, fmt.Errorf("auth smoke %s %s: expected %d, got %d (body withheld)", method, path, status, response.StatusCode)
		}
		var bodyResult map[string]any
		if status != 204 && json.NewDecoder(response.Body).Decode(&bodyResult) != nil {
			return nil, nil, errors.New("auth smoke invalid response")
		}
		var next *http.Cookie
		if cookies := response.Cookies(); len(cookies) > 0 {
			next = cookies[0]
		}
		return bodyResult, next, nil
	}
	credentials := map[string]string{"email": email, "password": password}
	_, registered, err := request("POST", "/auth/register", credentials, nil, 201)
	if err != nil {
		return err
	}
	if registered == nil || !registered.HttpOnly {
		return errors.New("auth smoke registration cookie missing")
	}
	if _, _, err = request("GET", "/me", nil, registered, 200); err != nil {
		return err
	}
	if _, _, err = request("POST", "/auth/logout", nil, registered, 204); err != nil {
		return err
	}
	_, loggedIn, err := request("POST", "/auth/login", credentials, nil, 200)
	if err != nil {
		return err
	}
	if loggedIn == nil || !loggedIn.HttpOnly {
		return errors.New("auth smoke login cookie missing")
	}
	handle := "smoke-" + suffix[:20]
	if _, _, err = request("PATCH", "/me/profile", map[string]any{"handle": handle, "display_name": "Temporary smoke profile", "bio": nil, "search_engine_indexing": false}, loggedIn, 200); err != nil {
		return err
	}
	profile, _, err := request("GET", "/users/"+handle, nil, nil, 200)
	if err != nil {
		return err
	}
	if len(profile) != 3 || profile["handle"] != handle {
		return errors.New("auth smoke public profile contract failed")
	}
	for _, key := range []string{"handle", "display_name", "bio"} {
		if _, exists := profile[key]; !exists {
			return errors.New("auth smoke public privacy check failed")
		}
	}
	if _, _, err = request("POST", "/auth/logout", nil, loggedIn, 204); err != nil {
		return err
	}
	if _, _, err = request("GET", "/me", nil, loggedIn, 401); err != nil {
		return err
	}
	fmt.Println("Real gfp_api register/login/me/profile/public-profile/logout checks passed (identity and credentials withheld)")
	return nil
}

func cleanup(pool *pgxpool.Pool, email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return database.SafeError("begin temporary identity cleanup", err)
	}
	defer tx.Rollback(ctx)
	// The unguessable address was generated in this process. It is never accepted
	// as CLI input; cleanup can target no user-selected or existing account.
	for _, table := range []string{"sessions", "password_credentials", "user_profiles"} {
		_, err = tx.Exec(ctx, "DELETE FROM app."+table+" WHERE user_id IN (SELECT user_id FROM app.auth_identities WHERE provider='email' AND provider_subject=$1)", email)
		if err != nil {
			return database.SafeError("temporary identity cleanup", err)
		}
	}
	var id string
	err = tx.QueryRow(ctx, "DELETE FROM app.auth_identities WHERE provider='email' AND provider_subject=$1 RETURNING user_id::text", email).Scan(&id)
	if err != nil {
		return database.SafeError("temporary identity lookup/cleanup", err)
	}
	if _, err = tx.Exec(ctx, "DELETE FROM app.users WHERE id=$1", id); err != nil {
		return database.SafeError("temporary account cleanup", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return database.SafeError("commit temporary identity cleanup", err)
	}
	fmt.Println("Temporary auth smoke identity and sessions removed using repository owner; no runtime privileges widened")
	return nil
}
