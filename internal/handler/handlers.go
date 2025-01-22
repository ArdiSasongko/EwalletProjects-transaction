package handler

import (
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external"
	//"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
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
}

func NewHandler( /*q *sqlc.Queries*/ db *pgxpool.Pool) Handlers {
	//service := service.NewService(q, db)
	userManagement := external.NewUserManagement()
	return Handlers{
		Health: &HealthHandler{},
		Middleware: &MiddlewareHandler{
			userManagement: userManagement,
		},
	}
}
