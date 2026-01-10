CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS payments
(
    id            UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    booking_id    UUID UNIQUE    NOT NULL,
    user_id       UUID           NOT NULL,
    amount        DECIMAL(10, 2) NOT NULL,
    currency      VARCHAR(3)     NOT NULL  DEFAULT 'RUB',
    status        VARCHAR(20)    NOT NULL,
    error_message TEXT,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    processed_at  TIMESTAMP WITH TIME ZONE
);