package handler

import (
	"fmt"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/config/logger"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/service"
	"github.com/gofiber/fiber/v2"
)

var log = logger.NewLogger()

type TransactionHandler struct {
	service service.Service
}

func (h *TransactionHandler) Create(ctx *fiber.Ctx) error {
	data := ctx.Locals("token").(model.TokenResponse)
	payload := new(model.TransactionPayload)

	payload.UserID = data.UserID

	if err := ctx.BodyParser(payload); err != nil {
		log.WithError(err).Errorf("bad request error, method: %v, path: %v", ctx.Method(), ctx.Path())
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := payload.Validate(); err != nil {
		errorValidate := fmt.Errorf("validate error")
		log.WithError(errorValidate).Errorf("bad request error, method: %v, path: %v", ctx.Method(), ctx.Path())
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	resp, err := h.service.Transaction.Create(ctx.Context(), payload)
	if err != nil {
		log.WithError(err).Errorf("internal server error, method: %v, path: %v", ctx.Method(), ctx.Path())
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "ok",
		"data":    resp,
	})
}
