CREATE TABLE airports
(
    code     CHAR(3) PRIMARY KEY,
    name     VARCHAR(100) NOT NULL,
    city     VARCHAR(100) NOT NULL,
    country  VARCHAR(100) NOT NULL,
    timezone VARCHAR(50)  NOT NULL
);

CREATE TABLE aircrafts
(
    id          SERIAL PRIMARY KEY,
    model       VARCHAR(100) NOT NULL,
    total_seats INT          NOT NULL,
    CONSTRAINT unique_aircraft_model UNIQUE (model)
);

CREATE TABLE aircraft_seats
(
    id          SERIAL PRIMARY KEY,
    aircraft_id INT         NOT NULL REFERENCES aircrafts (id) ON DELETE CASCADE,
    seat_number VARCHAR(5)  NOT NULL,
    seat_class  VARCHAR(20) NOT NULL CHECK (seat_class IN ('ECONOMY', 'BUSINESS', 'FIRST')),
    row_number  INT         NOT NULL,
    col_number  INT         NOT NULL,
    CONSTRAINT unique_aircraft_seat UNIQUE (aircraft_id, seat_number)
);

CREATE INDEX idx_aircraft_seats_aircraft ON aircraft_seats (aircraft_id);

CREATE TABLE airlines
(
    id        SERIAL PRIMARY KEY,
    iata_code CHAR(2) UNIQUE NOT NULL,
    name      VARCHAR(100)   NOT NULL,
    country   VARCHAR(100),
    logo_url  VARCHAR(255)
);

CREATE TABLE flights
(
    id                BIGSERIAL PRIMARY KEY,
    flight_number     VARCHAR(10) NOT NULL,
    airline_id        INT         NOT NULL REFERENCES airlines (id),
    aircraft_id       INT         NOT NULL REFERENCES aircrafts (id),
    departure_airport CHAR(3)     NOT NULL REFERENCES airports (code),
    arrival_airport   CHAR(3)     NOT NULL REFERENCES airports (code),
    departure_time    TIMESTAMPTZ NOT NULL,
    arrival_time      TIMESTAMPTZ NOT NULL,
    status            VARCHAR(20) DEFAULT 'SCHEDULED' CHECK (status IN ('SCHEDULED', 'DEPARTED', 'ARRIVED', 'CANCELLED')),
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_flights_search ON flights (departure_airport, arrival_airport, departure_time);

CREATE TABLE flight_fares
(
    id           BIGSERIAL PRIMARY KEY,
    flight_id    BIGINT      NOT NULL REFERENCES flights (id) ON DELETE CASCADE,
    seat_class   VARCHAR(20) NOT NULL CHECK (seat_class IN ('ECONOMY', 'BUSINESS', 'FIRST')),
    price_cents  BIGINT      NOT NULL,
    seats_total  INT         NOT NULL,
    seats_booked INT DEFAULT 0,
    CONSTRAINT unique_flight_fare UNIQUE (flight_id, seat_class),
    CONSTRAINT check_capacity CHECK (seats_booked <= seats_total)
);

CREATE TABLE seat_reservations
(
    id          BIGSERIAL PRIMARY KEY,
    flight_id   BIGINT      NOT NULL REFERENCES flights (id) ON DELETE CASCADE,
    seat_number VARCHAR(5)  NOT NULL,
    booking_id  UUID        NOT NULL,
    status      VARCHAR(20) NOT NULL CHECK (status IN ('LOCKED', 'CONFIRMED')),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at  TIMESTAMPTZ,

    CONSTRAINT unique_flight_seat UNIQUE (flight_id, seat_number)
);

CREATE INDEX idx_reservations_expiry ON seat_reservations (expires_at) WHERE status = 'LOCKED';

-- Outbox for ElasticSearch sync
CREATE TABLE flight_outbox
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    event_type  VARCHAR(50) NOT NULL,
    payload     JSONB       NOT NULL,
    status      VARCHAR(20)          DEFAULT 'PENDING',
    retry_count INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ          DEFAULT NOW()
);

CREATE INDEX idx_outbox_events_pending
    ON flight_outbox (created_at ASC) WHERE status = 'PENDING' AND retry_count < 5;