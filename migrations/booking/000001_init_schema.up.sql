CREATE TABLE bookings
(
    id                UUID PRIMARY KEY           DEFAULT gen_random_uuid(),
    user_id           BIGINT            NOT NULL,
    flight_id         BIGINT            NOT NULL,
    ref_code          VARCHAR(6) UNIQUE NOT NULL,
    status            VARCHAR(20)       NOT NULL DEFAULT 'PENDING', -- PENDING, PAID, CANCELLED
    total_price_cents BIGINT            NOT NULL,
    currency          CHAR(3)           NOT NULL DEFAULT 'RUB',
    contact_email     VARCHAR(255)      NOT NULL,
    expires_at        TIMESTAMPTZ       NOT NULL,
    created_at        TIMESTAMPTZ                DEFAULT NOW()
);

CREATE INDEX idx_bookings_user ON bookings (user_id);

CREATE INDEX idx_bookings_status_expires ON bookings (status, expires_at) WHERE
    status = 'PENDING';

CREATE TABLE tickets
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id           UUID         NOT NULL REFERENCES bookings (id) ON DELETE CASCADE,
    passenger_first_name VARCHAR(100) NOT NULL,
    passenger_last_name  VARCHAR(100) NOT NULL,
    passenger_doc_num    VARCHAR(50)  NOT NULL,
    seat_number          VARCHAR(5)   NOT NULL,
    price_cents          BIGINT       NOT NULL,
    ticket_number        VARCHAR(20),
    status               VARCHAR(20)      DEFAULT 'RESERVED' -- RESERVED, ISSUED
);

CREATE INDEX idx_tickets_booking_id ON tickets (booking_id);

CREATE TABLE booking_outbox
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    payload    JSONB       NOT NULL,
    status     VARCHAR(20)      DEFAULT 'PENDING',
    created_at TIMESTAMPTZ      DEFAULT NOW()
);

CREATE INDEX idx_outbox_pending ON booking_outbox (created_at) WHERE
    status = 'PENDING';
