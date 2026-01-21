CREATE TABLE IF NOT EXISTS booking_outbox
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type    VARCHAR(50) NOT NULL,
    payload       JSONB       NOT NULL,
    status        VARCHAR(20)      DEFAULT 'PENDING',
    created_at    TIMESTAMP        DEFAULT NOW(),
    processed_at  TIMESTAMP,
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_outbox_status ON booking_outbox (status, created_at) WHERE status = 'PENDING';