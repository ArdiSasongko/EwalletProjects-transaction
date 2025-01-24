package service

import (
	"context"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	Transaction interface {
		Create(context.Context, *model.TransactionPayload) (sqlc.CreateTransactionRow, error)
		UpdateTransaction(context.Context, *model.TransactionUpdatePayload) (model.TransactionResponse, error)
		GetTransasction(context.Context, *model.GetTransaction) (sqlc.Transaction, error)
		GetTransactions(context.Context, *model.GetTransactions) ([]sqlc.GetTransactionsRow, error)
		CreateRefund(context.Context, *model.TransactionRefundPayload) (*model.RefundResponse, error)
	}
}

func NewService(q *sqlc.Queries, db *pgxpool.Pool) Service {
	external := external.NewExternal()
	return Service{
		Transaction: &TransactionService{
			q:        q,
			db:       db,
			external: external,
		},
	}
}
