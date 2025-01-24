package model

type TokenResponse struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
}
