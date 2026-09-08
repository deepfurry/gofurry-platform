// Package config reads process environment without loading developer files.
package config

import (
	"errors"
	"net"
	"net/url"
	"os"
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
}

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
