CREATE TABLE IF NOT EXISTS flight_outbox
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   VARCHAR(50) NOT NULL,
    payload      JSONB       NOT NULL,
    status       VARCHAR(20)      DEFAULT 'PENDING',
    created_at   TIMESTAMP        DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX idx_outbox_status_created ON flight_outbox (status, created_at);