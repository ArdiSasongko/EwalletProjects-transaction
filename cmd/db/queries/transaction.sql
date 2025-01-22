-- name: CreateTransaction :one
INSERT INTO transaction (user_id, amount, transaction_type, transaction_status, reference, description, additional_info)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING reference, transaction_status;