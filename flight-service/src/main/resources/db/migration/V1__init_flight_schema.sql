CREATE TABLE aircrafts
(
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(50)  NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    total_seats INTEGER      NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE aircraft_seats
(
    id          BIGSERIAL PRIMARY KEY,
    aircraft_id BIGINT      NOT NULL REFERENCES aircrafts (id) ON DELETE CASCADE,
    row_number  INTEGER     NOT NULL,
    seat_column VARCHAR(2)  NOT NULL, -- A
    seat_number VARCHAR(10) NOT NULL, -- 12A
    cabin_class VARCHAR(50) NOT NULL, -- ECONOMY
    is_window   BOOLEAN     NOT NULL DEFAULT FALSE,
    is_aisle    BOOLEAN     NOT NULL DEFAULT FALSE,
    is_exit_row BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_aircraft_seat UNIQUE (aircraft_id, seat_number)
);

CREATE TABLE airports
(
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(10)  NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    city       VARCHAR(255) NOT NULL,
    country    VARCHAR(255) NOT NULL,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TABLE flights
(
    id                     BIGSERIAL PRIMARY KEY,
    flight_number          VARCHAR(20)    NOT NULL,
    origin_airport_id      BIGINT         NOT NULL REFERENCES airports (id),
    destination_airport_id BIGINT         NOT NULL REFERENCES airports (id),
    departure_time         TIMESTAMP      NOT NULL,
    arrival_time           TIMESTAMP      NOT NULL,
    base_price             NUMERIC(10, 2) NOT NULL,
    currency               VARCHAR(3)     NOT NULL,
    status                 VARCHAR(20)    NOT NULL DEFAULT 'SCHEDULED',
    aircraft_id            BIGINT         NOT NULL REFERENCES aircrafts (id),
    created_at             TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMP      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_flights_flight_number_departure
    ON flights (flight_number, departure_time);

CREATE INDEX idx_flights_origin_departure
    ON flights (origin_airport_id, departure_time);

CREATE INDEX idx_flights_destination_departure
    ON flights (destination_airport_id, departure_time);

CREATE INDEX idx_flights_aircraft
    ON flights (aircraft_id);
