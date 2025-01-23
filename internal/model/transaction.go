package model

import (
	"time"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

type TransactionPayload struct {
	UserID          int32   `json:"user_id"`
	Amount          float64 `json:"amount" validate:"required,numeric,gte=1000"`
	TransactionType string  `json:"transaction_type" validate:"required"`
	Description     string  `json:"description" validate:"required,min=5,max=255"`
	AdditionalInfo  string  `json:"additional_info" validate:"omitempty"`
}

func (u *TransactionPayload) Validate() error {
	return Validate.Struct(u)
}

type TransactionUpdatePayload struct {
	Reference         string `json:"reference"`
	TransactionStatus string `json:"transaction_status" validate:"required"`
	AdditionalInfo    string `json:"additional_info"`
	Token             string
}

func (u *TransactionUpdatePayload) Validate() error {
	return Validate.Struct(u)
}

type TransactionResponse struct {
	WalletID  int32     `json:"wallet_id"`
	Reference string    `json:"reference"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type GetTransaction struct {
	UserID    int32
	Reference string
}

type GetTransactions struct {
	UserID int32
	Limit  int32
	Offset int32
}

type TransactionRefundPayload struct {
	Reference      string `json:"reference" validate:"required"`
	Description    string `json:"description"`
	AdditionalInfo string `json:"additional_info"`
	Token          string
}

func (u *TransactionRefundPayload) Validate() error {
	return Validate.Struct(u)
}

type RefundResponse struct {
	Reference         string    `json:"reference"`
	TransactionStatus string    `json:"transaction_status"`
	Amount            float64   `json:"amount"`
	CreatedAt         time.Time `json:"created_at"`
}
