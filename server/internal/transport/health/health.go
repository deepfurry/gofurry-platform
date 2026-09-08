// Package health implements the shared dependency policy without transport DTOs.
package health

import (
	"context"
	"sync/atomic"
	"time"
)

type Check func(context.Context) error
type State struct{ Status, Postgres, Redis string }
type Checker struct {
	postgres, redis Check
	stopping        atomic.Bool
}

func New(postgres, redis Check) *Checker { return &Checker{postgres: postgres, redis: redis} }
func (c *Checker) Stop()                 { c.stopping.Store(true) }

func (c *Checker) Ready(ctx context.Context) State {
	result := State{Status: "unavailable", Postgres: "down", Redis: "down"}
	if c.stopping.Load() {
		return result
	}
	pgCtx, cancelPG := context.WithTimeout(ctx, 2*time.Second)
	pgErr := c.postgres(pgCtx)
	cancelPG()
	redisCtx, cancelRedis := context.WithTimeout(ctx, time.Second)
	redisErr := c.redis(redisCtx)
	cancelRedis()
	if redisErr == nil {
		result.Redis = "up"
	}
	if pgErr == nil {
		result.Postgres, result.Status = "up", "ready"
		if redisErr != nil {
			result.Status = "degraded"
		}
	}
	if c.stopping.Load() {
		result.Status = "unavailable"
	}
	return result
}
