// Package redisstore owns the gfp: key namespace, not a generic cache framework.
package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	client *redis.Client
	prefix string
}

type safeLogger struct{}

func (safeLogger) Printf(ctx context.Context, _ string, _ ...interface{}) {
	slog.WarnContext(ctx, "Redis driver event (details withheld)")
}
func init() { redis.SetLogger(safeLogger{}) }

func Open(rawURL, prefix string) (*Store, error) {
	if prefix != "gfp:" {
		return nil, errors.New("Redis key prefix must be gfp:")
	}
	options, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, errors.New("invalid Redis configuration (details withheld)")
	}
	options.DialTimeout, options.ReadTimeout, options.WriteTimeout = 3*time.Second, 2*time.Second, 2*time.Second
	options.ContextTimeoutEnabled = true
	options.MaxRetries = -1
	options.DisableIdentity = true
	return &Store{client: redis.NewClient(options), prefix: prefix}, nil
}

func (s *Store) Key(suffix string) string { return s.prefix + suffix }
func (s *Store) Close() error             { return s.client.Close() }
func (s *Store) Ping(ctx context.Context) error {
	if err := s.client.Ping(ctx).Err(); err != nil {
		return errors.New("Redis PING failed (details withheld)")
	}
	return nil
}

func (s *Store) Smoke(ctx context.Context) error {
	if err := s.Ping(ctx); err != nil {
		return err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return errors.New("Redis smoke nonce failed")
	}
	value := hex.EncodeToString(nonce[:])
	key := s.Key("smoke:" + value)
	if err := s.client.Set(ctx, key, value, time.Minute).Err(); err != nil {
		return errors.New("Redis smoke SET failed (details withheld)")
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = s.client.Del(cleanup, key).Err()
	}()
	got, err := s.client.Get(ctx, key).Result()
	if err != nil || got != value {
		return errors.New("Redis smoke GET failed (details withheld)")
	}
	if deleted, err := s.client.Del(ctx, key).Result(); err != nil || deleted != 1 {
		return errors.New("Redis smoke DEL failed (details withheld)")
	}
	return nil
}
