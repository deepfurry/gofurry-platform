package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deepfurry/gofurry-platform/server/internal/config"
	"github.com/deepfurry/gofurry-platform/server/internal/database"
	"github.com/deepfurry/gofurry-platform/server/internal/redisstore"
	platformruntime "github.com/deepfurry/gofurry-platform/server/internal/runtime"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/health"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/public"
	"github.com/gofiber/fiber/v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	c, err := config.Load("api")
	if err != nil {
		return err
	}
	logger := platformruntime.Logger("api", c.Environment)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	defer platformruntime.ShutdownDeadline(ctx, logger)()
	pool, err := database.Open(ctx, c.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	store, err := redisstore.Open(c.RedisURL, c.RedisKeyPrefix)
	if err != nil {
		return err
	}
	defer store.Close()
	checker := health.New(func(ctx context.Context) error { return database.Ready(ctx, pool) }, store.Ping)
	app := fiber.New(fiber.Config{ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second})
	public.Register(app, checker)
	return platformruntime.HTTP(ctx, app, c.HTTPAddr, checker, logger)
}
