package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/model"
	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusPending  = "PENDING"
	StatusSuccess  = "SUCCESS"
	StatusFailed   = "FAILED"
	StatusReversed = "REVERSED"
)

var transType = map[string]bool{
	"TOPUP":    true,
	"PURCHASE": true,
	"REFUND":   true,
}

func generateReference(typeTrans string, userID int32) string {
	now := time.Now()
	timeFormatted := now.Format("200601022150405")
	reference := fmt.Sprintf("%d%s%s", userID, typeTrans, timeFormatted)
	return reference
}

type TransactionService struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func (s *TransactionService) Create(ctx context.Context, payload *model.TransactionPayload) (sqlc.CreateTransactionRow, error) {
	if !transType[payload.TransactionType] {
		return sqlc.CreateTransactionRow{}, fmt.Errorf("transaction type not allowed only 'TOPUP', 'PURCHASE', 'REFUND'")
	}

	if payload.AdditionalInfo == "" {
		payload.AdditionalInfo = " "
	}

	reference := generateReference(payload.TransactionType, payload.UserID)

	amountBigInt := big.NewInt(int64(payload.Amount))
	resp, err := s.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID: payload.UserID,
		Amount: pgtype.Numeric{
			Int:   amountBigInt,
			Valid: true,
		},
		TransactionType:   sqlc.TransactionType(payload.TransactionType),
		TransactionStatus: StatusPending,
		Description: pgtype.Text{
			String: payload.Description,
			Valid:  true,
		},
		AdditionalInfo: pgtype.Text{
			String: payload.AdditionalInfo,
			Valid:  true,
		},
		Reference: reference,
	})

	if err != nil {
		return sqlc.CreateTransactionRow{}, err
	}

	return resp, nil
}
