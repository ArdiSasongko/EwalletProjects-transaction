package handler

import "github.com/gofiber/fiber/v2"

type HealthHandler struct{}

func (h HealthHandler) CheckHealth(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"route":   ctx.Route().Path,
	})
}
