CREATE TABLE IF NOT EXISTS airports
(
    code     CHAR(3) PRIMARY KEY,
    name     VARCHAR(100) NOT NULL,
    city     VARCHAR(100) NOT NULL,
    country  VARCHAR(100) NOT NULL,
    timezone VARCHAR(50)  NOT NULL
);

CREATE TABLE IF NOT EXISTS flights
(
    id                SERIAL PRIMARY KEY,
    flight_number     VARCHAR(10)    NOT NULL,
    departure_airport CHAR(3)        NOT NULL REFERENCES airports (code),
    arrival_airport   CHAR(3)        NOT NULL REFERENCES airports (code),
    departure_time    TIMESTAMP      NOT NULL,
    arrival_time      TIMESTAMP      NOT NULL,
    price             DECIMAL(10, 2) NOT NULL,
    total_seats       INT            NOT NULL,
    status            VARCHAR(20) DEFAULT 'scheduled',
    created_at        TIMESTAMP   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flights_route ON flights (departure_airport, arrival_airport, departure_time);

CREATE TABLE IF NOT EXISTS seats
(
    id               SERIAL PRIMARY KEY,
    flight_id        INT        NOT NULL REFERENCES flights (id) ON DELETE CASCADE,
    seat_number      VARCHAR(5) NOT NULL,
    is_booked        BOOLEAN       DEFAULT FALSE,
    price_multiplier DECIMAL(3, 2) DEFAULT 1.0,
    CONSTRAINT unique_seat_per_flight UNIQUE (flight_id, seat_number)
);