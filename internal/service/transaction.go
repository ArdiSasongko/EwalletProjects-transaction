package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
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
	db       *pgxpool.Pool
	q        *sqlc.Queries
	external external.External
}

func roundToTwoDecimalPlaces(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func (s *TransactionService) Create(ctx context.Context, payload *model.TransactionPayload) (sqlc.CreateTransactionRow, error) {
	if !transType[payload.TransactionType] {
		return sqlc.CreateTransactionRow{}, fmt.Errorf("transaction type not allowed only 'TOPUP', 'PURCHASE', 'REFUND'")
	}

	jsonAditionalInfo := map[string]interface{}{}
	if payload.AdditionalInfo != "" {
		err := json.Unmarshal([]byte(payload.AdditionalInfo), &jsonAditionalInfo)
		if err != nil {
			return sqlc.CreateTransactionRow{}, fmt.Errorf("additional info invalid format")
		}
	}
	reference := generateReference(payload.TransactionType, payload.UserID)

	amountFloat := roundToTwoDecimalPlaces(payload.Amount)
	amountStr := fmt.Sprintf("%.2f", amountFloat)
	amountNumeric := pgtype.Numeric{}
	if err := amountNumeric.Scan(amountStr); err != nil {
		return sqlc.CreateTransactionRow{}, err
	}

	resp, err := s.q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		UserID:            payload.UserID,
		Amount:            amountNumeric,
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

	if payload.TransactionStatus == StatusFailed {
		switch {
		case strings.Contains(payload.Reference, string(sqlc.TransactionTypeTOPUP)):
			amount, _ := tsx.Amount.Float64Value()
			if err := s.external.Notif.SendNotification(ctx, external.NotifRequest{
				Recipient:    payload.Email,
				TemplateName: "topup_failed",
				Placeholder: map[string]string{
					"user_id":    string(tsx.UserID),
					"amount":     fmt.Sprintf("%.2f", amount.Float64),
					"reference":  payload.Reference,
					"created_at": tsx.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				},
			}); err != nil {
				return model.TransactionResponse{}, err
			}
		case strings.Contains(payload.Reference, string(sqlc.TransactionTypePURCHASE)):
			amount, _ := tsx.Amount.Float64Value()
			if err := s.external.Notif.SendNotification(ctx, external.NotifRequest{
				Recipient:    payload.Email,
				TemplateName: "purchase_failed",
				Placeholder: map[string]string{
					"user_id":    string(tsx.UserID),
					"amount":     fmt.Sprintf("%.2f", amount.Float64),
					"reference":  payload.Reference,
					"created_at": tsx.CreatedAt.Time.Format("2006-01-02 15:04:05"),
				},
			}); err != nil {
				return model.TransactionResponse{}, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return model.TransactionResponse{}, fmt.Errorf("failed to process transaction")
		}
		return model.TransactionResponse{
			Reference: payload.Reference,
			Status:    payload.TransactionStatus,
		}, nil
	}

	amountFloat, _ := tsx.Amount.Float64Value()

	updatePayload := external.WalletRequest{
		Amount:    amountFloat.Float64,
		Reference: tsx.Reference,
		Status:    string(resp),
	}

	var respTrans model.TransactionResponse
	log.Println(tsx.TransactionType)
	switch tsx.TransactionType {
	case sqlc.TransactionTypePURCHASE:
		d, err := s.external.Wallet.Debit(ctx, updatePayload, payload.Token)
		if err != nil {
			return model.TransactionResponse{}, fmt.Errorf("credit wallet error :%w", err)
		}
		log.Println()
		respTrans = model.TransactionResponse{
			WalletID:  d.UserID,
			Reference: d.Reference,
			Amount:    d.Amount,
			CreatedAt: d.CreatedAt,
		}

	case sqlc.TransactionTypeTOPUP:
		d, err := s.external.Wallet.Credit(ctx, updatePayload, payload.Token)
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

func (s *TransactionService) GetTransactions(ctx context.Context, payload *model.GetTransactions) ([]sqlc.GetTransactionsRow, error) {
	pageSize := payload.Limit
	pageNumber := payload.Offset

	limit := pageSize
	offset := (pageNumber - 1) * pageSize

	resp, err := s.q.GetTransactions(ctx, sqlc.GetTransactionsParams{
		UserID: payload.UserID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *TransactionService) GetTransasction(ctx context.Context, payload *model.GetTransaction) (sqlc.Transaction, error) {
	resp, err := s.q.GetTransactionByReferenceAndUserId(ctx, sqlc.GetTransactionByReferenceAndUserIdParams{
		UserID:    payload.UserID,
		Reference: payload.Reference,
	})

	if err != nil {
		return sqlc.Transaction{}, err
	}

	return resp, nil
}

func (s *TransactionService) CreateRefund(ctx context.Context, payload *model.TransactionRefundPayload) (*model.RefundResponse, error) {
	// using transaction for consistent
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.q.WithTx(tx)
	// get transaction
	tsx, err := qtx.GetTransactionByReference(ctx, payload.Reference)
	if err != nil {
		return nil, err
	}

	log.Println(tsx.TransactionType)
	log.Println(tsx.TransactionStatus)
	// check type and status
	if tsx.TransactionType != sqlc.TransactionTypePURCHASE || tsx.TransactionStatus != StatusSuccess {
		return nil, fmt.Errorf("only type 'PURCHASE' and status 'SUCCESS' can be refunded")
	}

	jsonAditionalInfo := map[string]interface{}{}
	if payload.AdditionalInfo != "" {
		err := json.Unmarshal([]byte(payload.AdditionalInfo), &jsonAditionalInfo)
		if err != nil {
			return nil, fmt.Errorf("additional info invalid format")
		}
	}
	// generate new reference
	ref := generateReference(string(sqlc.TransactionTypeREFUND), tsx.UserID)
	// create model for createtransaction
	tsxReq := sqlc.CreateTransactionParams{
		UserID:            tsx.UserID,
		Amount:            tsx.Amount,
		TransactionType:   sqlc.TransactionTypeREFUND,
		TransactionStatus: sqlc.TransactionStatusSUCCESS,
		Reference:         ref,
		Description: pgtype.Text{
			String: payload.Description,
			Valid:  true,
		},
		AdditionalInfo: pgtype.Text{
			String: payload.AdditionalInfo,
			Valid:  true,
		},
	}

	resp, err := qtx.CreateTransaction(ctx, tsxReq)
	if err != nil {
		return nil, err
	}

	// connect to wallet (credit)
	amount, _ := tsx.Amount.Float64Value()
	walletRequest := external.WalletRequest{
		Reference: resp.Reference,
		Amount:    amount.Float64,
	}

	walletResp, err := s.external.Wallet.Credit(ctx, walletRequest, payload.Token)
	if err != nil {
		return nil, err
	}

	response := model.RefundResponse{
		Reference:         walletRequest.Reference,
		TransactionStatus: string(resp.TransactionStatus),
		Amount:            walletRequest.Amount,
		CreatedAt:         walletResp.CreatedAt,
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &response, nil
}
