package config

import (
	"strings"
	"testing"
)

func TestServiceEnvironmentContract(t *testing.T) {
	base := map[string]string{"APP_ENV": "test", "DATABASE_URL": "postgres://local:sentinel@localhost/gfp_ci", "REDIS_URL": "redis://localhost:6379", "REDIS_KEY_PREFIX": "gfp:", "RIVER_SCHEMA": "river", "HTTP_ADDR": "127.0.0.1:8080"}
	for _, service := range []string{"api", "admin", "worker", "migrator"} {
		t.Run(service, func(t *testing.T) {
			if _, err := load(service, func(k string) string { return base[k] }); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, key := range []string{"APP_ENV", "DATABASE_URL", "REDIS_URL", "REDIS_KEY_PREFIX", "HTTP_ADDR"} {
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
