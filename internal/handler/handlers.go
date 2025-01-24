package handler

import (
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/service"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	Health interface {
		CheckHealth(*fiber.Ctx) error
	}
	Middleware interface {
		AuthMiddleware() fiber.Handler
	}
	Transaction interface {
		Create(*fiber.Ctx) error
		Update(*fiber.Ctx) error
		GetTransaction(*fiber.Ctx) error
		GetTransactions(*fiber.Ctx) error
		Refund(*fiber.Ctx) error
	}
}

func NewHandler(q *sqlc.Queries, db *pgxpool.Pool) Handlers {
	service := service.NewService(q, db)
	external := external.NewExternal()
	return Handlers{
		Health: &HealthHandler{},
		Middleware: &MiddlewareHandler{
			external: external,
		},
		Transaction: &TransactionHandler{
			service: service,
		},
	}
}
