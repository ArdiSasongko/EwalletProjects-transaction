package model

import "github.com/go-playground/validator/v10"

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

type TransactionPayload struct {
	UserID          int32  `json:"user_id"`
	Amount          int32  `json:"amount" validate:"required,numeric,gte=1000"`
	TransactionType string `json:"transaction_type" validate:"required"`
	Description     string `json:"description" validate:"required,min=5,max=255"`
	AdditionalInfo  string `json:"additional_info" validate:"omitempty"`
}

func (u *TransactionPayload) Validate() error {
	return Validate.Struct(u)
}
