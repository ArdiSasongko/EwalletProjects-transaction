-- name: CreateTransaction :one
INSERT INTO transaction (user_id, amount, transaction_type, transaction_status, reference, description, additional_info)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING reference, transaction_status;

-- name: GetTransactionByReference :one
SELECT id, user_id, amount, transaction_type, transaction_status, reference, description, additional_info, created_at, updated_at
FROM transaction WHERE reference = $1;

-- name: UpdateTransactionStatusByReference :one
UPDATE transaction SET transaction_status = $2, additional_info = $3, updated_at = CURRENT_TIMESTAMP
WHERE reference = $1
RETURNING transaction_status;

-- name: GetTransactions :many
SELECT reference, transaction_status, amount, transaction_type, created_at
FROM transaction WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetTransactionByReferenceAndUserId :one
SELECT id, user_id, amount, transaction_type, transaction_status, reference, description, additional_info, created_at, updated_at
FROM transaction 
WHERE reference = $1 AND user_id = $2;
