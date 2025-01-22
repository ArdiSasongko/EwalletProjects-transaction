package handler

import (
	"strings"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external"
	"github.com/gofiber/fiber/v2"
)

type MiddlewareHandler struct {
	userManagement external.UserManagement
}

func (h *MiddlewareHandler) AuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authToken := ctx.Get("Authorization")
		if authToken == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization headers",
			})
		}

		rContext := ctx.Context()

		parts := strings.Split(authToken, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authorization headers are malformed",
			})
		}

		token := parts[1]

		userID, err := h.userManagement.ValidateToken(rContext, token)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		log.Println(userID)

		ctx.Locals("token", userID)
		return ctx.Next()
	}
}
