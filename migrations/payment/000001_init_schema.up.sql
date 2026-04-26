CREATE TABLE payments
(
    id              UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    booking_id      UUID        NOT NULL,
    user_id         BIGINT      NOT NULL,
    amount_cents    BIGINT      NOT NULL CHECK (amount_cents > 0),
    currency        CHAR(3)     NOT NULL DEFAULT 'RUB',
    status          VARCHAR(20) NOT NULL, -- PENDING, SUCCESS, FAILED
    provider_txn_id VARCHAR(255),
    error_message   TEXT,
    created_at      TIMESTAMPTZ          DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX idx_payments_user_id ON payments (user_id);
CREATE UNIQUE INDEX idx_payments_booking_id ON payments (booking_id);
