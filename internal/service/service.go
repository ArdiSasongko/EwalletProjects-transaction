package service

import (
	"context"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Transaction interface {
		Create(context.Context, *model.TransactionPayload) (sqlc.CreateTransactionRow, error)
	}
}

func NewService(q *sqlc.Queries, db *pgxpool.Pool) Service {
	return Service{
		Transaction: &TransactionService{
			q:  q,
			db: db,
		},
	}
}
