package public

import (
	"context"
	"github.com/deepfurry/gofurry-platform/server/internal/auth"
	"github.com/deepfurry/gofurry-platform/server/internal/identity"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/health"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/public/generated"
	"github.com/gofiber/fiber/v3"
	"time"
)

type Handler struct {
	health   *health.Checker
	auth     *auth.App
	identity *identity.App
	options  Options
}
type Options struct{ Environment, PublicOrigin string }

var _ generated.ServerInterface = (*Handler)(nil)

func Register(router fiber.Router, checker *health.Checker, authentication *auth.App, identities *identity.App, options Options) {
	h := &Handler{health: checker, auth: authentication, identity: identities, options: options}
	for _, path := range []string{"/auth", "/me", "/users"} {
		router.Use(path, func(c fiber.Ctx) error {
			ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
			defer cancel()
			c.SetContext(ctx)
			return c.Next()
		})
	}
	// Only the P0-1A auth/profile routes get this browser-origin boundary.
	for _, path := range []string{"/auth/register", "/auth/login", "/auth/logout", "/me/profile"} {
		router.Use(path, h.originGuard)
	}
	generated.RegisterHandlers(router, h)
}
func (*Handler) GetLive(c fiber.Ctx) error { return c.JSON(generated.Live{Status: "alive"}) }
func (h *Handler) GetReady(c fiber.Ctx) error {
	state := h.health.Ready(c.Context())
	code := fiber.StatusOK
	if state.Status == "unavailable" {
		code = fiber.StatusServiceUnavailable
	}
	return c.Status(code).JSON(generated.Ready{Status: generated.ReadyStatus(state.Status), Postgres: generated.ReadyPostgres(state.Postgres), Redis: generated.ReadyRedis(state.Redis)})
}
