package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ArdiSasongko/EwalletProjects-transaction/internal/external"
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

var flowStatus = map[string][]string{
	StatusPending: {StatusSuccess, StatusFailed},
	StatusSuccess: {StatusReversed},
	StatusFailed:  {StatusSuccess},
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
	w  external.Wallet
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

func (s *TransactionService) UpdateTransaction(ctx context.Context, payload *model.TransactionUpdatePayload) (model.TransactionResponse, error) {
	tsx, err := s.q.GetTransactionByReference(ctx, payload.Reference)
	if err != nil {
		return model.TransactionResponse{}, err
	}

	statusValid := false
	statusTransaction := flowStatus[string(tsx.TransactionStatus)]

	for i := range statusTransaction {
		if statusTransaction[i] == payload.TransactionStatus {
			statusValid = true
		}
	}

	if !statusValid {
		return model.TransactionResponse{}, fmt.Errorf("transaction status flow invalid, payload status - %s", payload.TransactionStatus)
	}

	currentAditionalInfo := map[string]interface{}{}
	if tsx.AdditionalInfo.Valid && tsx.AdditionalInfo.String != "" {
		if err := json.Unmarshal([]byte(tsx.AdditionalInfo.String), &currentAditionalInfo); err != nil {
			fmt.Printf("warning: failed to unmarshal current additional info: %v\n", err)
			currentAditionalInfo = map[string]interface{}{}
		}
	}

	if payload.AdditionalInfo != "" {
		newAdditionalInfo := map[string]interface{}{}
		if err := json.Unmarshal([]byte(payload.AdditionalInfo), &newAdditionalInfo); err != nil {
			return model.TransactionResponse{}, fmt.Errorf("failed to umarshal new additional info: %w", err)
		}

		for key, val := range newAdditionalInfo {
			currentAditionalInfo[key] = val
		}
	}

	var additionalInfo pgtype.Text
	if len(currentAditionalInfo) > 0 {
		byteAdditionalInfo, err := json.Marshal(currentAditionalInfo)
		if err != nil {
			return model.TransactionResponse{}, fmt.Errorf("failed to umarshal updated additional info: %w", err)
		}

		additionalInfo = pgtype.Text{
			String: string(byteAdditionalInfo),
			Valid:  true,
		}
	} else {
		additionalInfo = pgtype.Text{
			Valid: false,
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return model.TransactionResponse{}, fmt.Errorf("failed start database tx : %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)

	resp, err := qtx.UpdateTransactionStatusByReference(ctx, sqlc.UpdateTransactionStatusByReferenceParams{
		Reference:         payload.Reference,
		AdditionalInfo:    additionalInfo,
		TransactionStatus: sqlc.TransactionStatus(payload.TransactionStatus),
	})
	if err != nil {
		return model.TransactionResponse{}, err
	}

	amountFloat, _ := tsx.Amount.Float64Value()

	updatePayload := external.WalletRequest{
		Amount:    amountFloat.Float64,
		Reference: tsx.Reference,
	}

	var respTrans model.TransactionResponse
	switch tsx.TransactionType {
	case sqlc.TransactionTypePURCHASE:
		d, err := s.w.Debit(ctx, updatePayload, payload.Token)
		if err != nil {
			return model.TransactionResponse{}, fmt.Errorf("credit wallet error :%w", err)
		}

		respTrans = model.TransactionResponse{
			WalletID:  d.UserID,
			Reference: d.Reference,
			Amount:    d.Amount,
			CreatedAt: d.CreatedAt,
		}

	case sqlc.TransactionTypeTOPUP:
		d, err := s.w.Credit(ctx, updatePayload, payload.Token)
		if err != nil {
			return model.TransactionResponse{}, fmt.Errorf("credit wallet error :%w", err)
		}
		respTrans = model.TransactionResponse{
			WalletID:  d.UserID,
			Reference: d.Reference,
			Amount:    d.Amount,
			CreatedAt: d.CreatedAt,
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.TransactionResponse{}, err
	}

	respTrans.Status = string(resp)
	return respTrans, nil
}
