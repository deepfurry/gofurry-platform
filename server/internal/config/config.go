// Package config reads process environment without loading developer files.
package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	Environment    string
	HTTPAddr       string
	DatabaseURL    string
	RedisURL       string
	RedisKeyPrefix string
	RiverSchema    string
	PublicOrigin   string
	CSRFSecret     string
	MailMode       string
	MailLocalDir   string
}

const DevelopmentCSRFSecret = "gofurry-development-only-csrf-secret"

func Load(service string) (Config, error) {
	return load(service, os.Getenv)
}

func load(service string, env func(string) string) (Config, error) {
	c := Config{Environment: env("APP_ENV"), DatabaseURL: env("DATABASE_URL")}
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return Config{}, errors.New("APP_ENV must be development, test or production")
	}
	if !validURL(c.DatabaseURL, "postgres", "postgresql") {
		return Config{}, errors.New("DATABASE_URL is required and must be a PostgreSQL URL")
	}
	switch service {
	case "api", "admin", "worker", "migrator":
	default:
		return Config{}, errors.New("unknown service")
	}
	if service == "api" || service == "admin" {
		c.HTTPAddr = env("HTTP_ADDR")
		_, port, err := net.SplitHostPort(c.HTTPAddr)
		n, parseErr := strconv.Atoi(port)
		if err != nil || parseErr != nil || n < 1 || n > 65535 {
			return Config{}, errors.New("HTTP_ADDR must be host:port with a valid port")
		}
	}
	if service == "api" {
		c.PublicOrigin = env("PUBLIC_ORIGIN")
		if c.PublicOrigin == "" && c.Environment == "development" {
			c.PublicOrigin = "http://localhost:4321"
		}
		u, err := url.Parse(c.PublicOrigin)
		if err != nil || u.Hostname() == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" ||
			(u.Scheme != "http" && u.Scheme != "https") || (c.Environment == "production" && u.Scheme != "https") {
			return Config{}, errors.New("PUBLIC_ORIGIN must be an exact origin (HTTPS required in production)")
		}
		c.CSRFSecret = env("CSRF_SECRET")
		if c.CSRFSecret == "" && c.Environment != "production" {
			c.CSRFSecret = DevelopmentCSRFSecret
		}
		if len(c.CSRFSecret) < 32 || (c.Environment == "production" && c.CSRFSecret == DevelopmentCSRFSecret) {
			return Config{}, errors.New("CSRF_SECRET must contain at least 32 bytes; production requires a private secret")
		}
		c.MailMode = env("MAIL_MODE")
		if c.MailMode == "" {
			c.MailMode = "local"
			if c.Environment == "production" {
				c.MailMode = "disabled"
			}
		}
		if c.MailMode != "local" && c.MailMode != "disabled" {
			return Config{}, errors.New("MAIL_MODE must be local or disabled")
		}
		if c.Environment == "production" && c.MailMode == "local" {
			return Config{}, errors.New("production cannot use local mail capture")
		}
		if c.MailMode == "local" {
			c.MailLocalDir = env("MAIL_LOCAL_DIR")
			if c.MailLocalDir == "" {
				c.MailLocalDir = filepath.Join("..", ".local", "mail")
			}
			root, _ := filepath.Abs(filepath.Join("..", ".local"))
			dir, err := filepath.Abs(c.MailLocalDir)
			rel, relErr := filepath.Rel(root, dir)
			if err != nil || relErr != nil || rel == "." || !filepath.IsLocal(rel) {
				return Config{}, errors.New("MAIL_LOCAL_DIR must be inside the repository private .local directory (launch from server)")
			}
			c.MailLocalDir = dir
		}
	}
	if service != "migrator" {
		c.RedisURL, c.RedisKeyPrefix = env("REDIS_URL"), env("REDIS_KEY_PREFIX")
		if !validURL(c.RedisURL, "redis", "rediss") {
			return Config{}, errors.New("REDIS_URL is required and must be a Redis URL")
		}
		if c.RedisKeyPrefix != "gfp:" {
			return Config{}, errors.New("REDIS_KEY_PREFIX must be gfp:")
		}
	}
	if service == "worker" || service == "migrator" {
		c.RiverSchema = env("RIVER_SCHEMA")
		if c.RiverSchema != "river" {
			return Config{}, errors.New("RIVER_SCHEMA must be river")
		}
	}
	return c, nil
}

func validURL(raw string, schemes ...string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.Fragment != "" {
		return false
	}
	for _, scheme := range schemes {
		if u.Scheme == scheme {
			return true
		}
	}
	return false
}
