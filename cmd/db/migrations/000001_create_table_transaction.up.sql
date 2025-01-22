CREATE TYPE transaction_type AS ENUM ('TOPUP', 'PURCHASE', 'REFUND');
CREATE TYPE transaction_status AS ENUM ('PENDING', 'SUCCESS', 'FAILED', 'REVERSED');

CREATE TABLE IF NOT EXISTS transaction (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    transaction_type transaction_type NOT NULL,
    transaction_status transaction_status NOT NULL,
    reference VARCHAR(255) NOT NULL,
    description text,
    additional_info VARCHAR(255),
    created_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP
);