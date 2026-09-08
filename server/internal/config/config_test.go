package config

import (
	"strings"
	"testing"
)

func TestServiceEnvironmentContract(t *testing.T) {
	base := map[string]string{"APP_ENV": "test", "PUBLIC_ORIGIN": "http://localhost:4321", "DATABASE_URL": "postgres://local:sentinel@localhost/gfp_ci", "REDIS_URL": "redis://localhost:6379", "REDIS_KEY_PREFIX": "gfp:", "RIVER_SCHEMA": "river", "HTTP_ADDR": "127.0.0.1:8080"}
	for _, service := range []string{"api", "admin", "worker", "migrator"} {
		t.Run(service, func(t *testing.T) {
			if _, err := load(service, func(k string) string { return base[k] }); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, key := range []string{"APP_ENV", "DATABASE_URL", "REDIS_URL", "REDIS_KEY_PREFIX", "HTTP_ADDR", "PUBLIC_ORIGIN"} {
		t.Run("missing-"+key, func(t *testing.T) {
			_, err := load("api", func(k string) string {
				if k == key {
					return ""
				}
				return base[k]
			})
			if err == nil {
				t.Fatal("missing value accepted")
			}
		})
	}
	if _, err := load("worker", func(k string) string {
		if k == "RIVER_SCHEMA" {
			return "public"
		}
		return base[k]
	}); err == nil {
		t.Fatal("wrong River schema accepted")
	}
	_, err := load("api", func(k string) string {
		if k == "DATABASE_URL" {
			return "postgres://secret-value@%broken"
		}
		return base[k]
	})
	if err == nil || strings.Contains(err.Error(), "secret-value") {
		t.Fatal("invalid URL accepted or leaked")
	}
}

func TestPublicOriginEnvironmentRules(t *testing.T) {
	for _, test := range []struct {
		environment, origin string
		valid               bool
	}{
		{"development", "", true}, {"test", "", false}, {"test", "http://localhost:4321", true},
		{"production", "", false}, {"production", "http://example.com", false}, {"production", "https://example.com", true},
		{"production", "https://example.com/", false}, {"production", "https://example.com?x=1", false},
		{"production", "https://private-value@example.com", false}, {"production", "https://example.com#fragment", false},
		{"production", "null", false},
	} {
		env := map[string]string{"APP_ENV": test.environment, "PUBLIC_ORIGIN": test.origin, "DATABASE_URL": "postgres://localhost/gfp_ci", "REDIS_URL": "redis://localhost:6379", "REDIS_KEY_PREFIX": "gfp:", "HTTP_ADDR": "127.0.0.1:8080"}
		cfg, err := load("api", func(k string) string { return env[k] })
		if (err == nil) != test.valid {
			t.Error("PUBLIC_ORIGIN environment contract failed")
		}
		if err != nil && strings.Contains(err.Error(), "private-value") {
			t.Error("origin error leaked user info")
		}
		if test.environment == "development" && test.origin == "" && cfg.PublicOrigin != "http://localhost:4321" {
			t.Error("development origin does not match web")
		}
		if _, err := load("admin", func(k string) string { return env[k] }); err != nil {
			t.Error("Public Origin was required by Admin")
		}
	}
}
