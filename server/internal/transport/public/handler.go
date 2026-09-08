package public

import (
	"github.com/deepfurry/gofurry-platform/server/internal/transport/health"
	"github.com/deepfurry/gofurry-platform/server/internal/transport/public/generated"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{ health *health.Checker }

var _ generated.ServerInterface = (*Handler)(nil)

func Register(router fiber.Router, checker *health.Checker) {
	generated.RegisterHandlers(router, &Handler{health: checker})
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
