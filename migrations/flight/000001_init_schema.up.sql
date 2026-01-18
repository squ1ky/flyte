CREATE TABLE IF NOT EXISTS airports
(
    code     CHAR(3) PRIMARY KEY,
    name     VARCHAR(100) NOT NULL,
    city     VARCHAR(100) NOT NULL,
    country  VARCHAR(100) NOT NULL,
    timezone VARCHAR(50)  NOT NULL
);

CREATE TABLE IF NOT EXISTS aircrafts
(
    id          SERIAL PRIMARY KEY,
    model       VARCHAR(100) NOT NULL, -- "Boeing 737-800"
    total_seats INT          NOT NULL
);

CREATE TABLE IF NOT EXISTS aircraft_seats
(
    id               SERIAL PRIMARY KEY,
    aircraft_id      INT         NOT NULL REFERENCES aircrafts (id) ON DELETE CASCADE,
    seat_number      VARCHAR(5)  NOT NULL,                   -- "1A", "1B"
    seat_class       VARCHAR(20) NOT NULL DEFAULT 'economy', -- 'economy', 'business', 'comfort'
    price_multiplier DECIMAL(3, 2)        DEFAULT 1.0,

    CONSTRAINT unique_aircraft_seat UNIQUE (aircraft_id, seat_number)
);

CREATE TABLE IF NOT EXISTS flights
(
    id                SERIAL PRIMARY KEY,
    flight_number     VARCHAR(10) NOT NULL,
    aircraft_id       INT         NOT NULL REFERENCES aircrafts (id),
    departure_airport CHAR(3)     NOT NULL REFERENCES airports (code),
    arrival_airport   CHAR(3)     NOT NULL REFERENCES airports (code),
    departure_time    TIMESTAMP   NOT NULL,
    arrival_time      TIMESTAMP   NOT NULL,
    base_price_cents  BIGINT      NOT NULL,
    status            VARCHAR(20) DEFAULT 'scheduled', -- 'scheduled', 'cancelled', 'arrived'
    created_at        TIMESTAMP   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flights_route ON flights (departure_airport, arrival_airport, departure_time);

CREATE TABLE IF NOT EXISTS seats
(
    id               SERIAL PRIMARY KEY,
    flight_id        INT           NOT NULL REFERENCES flights (id) ON DELETE CASCADE,
    seat_number      VARCHAR(5)    NOT NULL,
    seat_class       VARCHAR(20)   NOT NULL,
    price_multiplier DECIMAL(3, 2) NOT NULL,

    is_booked        BOOLEAN   DEFAULT FALSE,
    reserved_at      TIMESTAMP DEFAULT NULL,

    CONSTRAINT unique_flight_seat UNIQUE (flight_id, seat_number)
);

CREATE INDEX IF NOT EXISTS idx_seats_reservation_cleanup ON seats (reserved_at) WHERE is_booked = TRUE;

-- Outbox for async ElasticSearch sync
CREATE TABLE IF NOT EXISTS flight_outbox
(
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   VARCHAR(50) NOT NULL, -- 'FLIGHT_CREATED', 'SEAT_BOOKED', 'SEAT_RELEASED'
    payload      JSONB       NOT NULL,
    status       VARCHAR(20)      DEFAULT 'PENDING',
    created_at   TIMESTAMP        DEFAULT NOW(),
    processed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_created ON flight_outbox (status, created_at);
