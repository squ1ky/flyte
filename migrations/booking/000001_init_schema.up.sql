CREATE TABLE IF NOT EXISTS bookings
(
    id                 UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    user_id            BIGINT         NOT NULL,
    flight_id          BIGINT         NOT NULL,
    seat_number        VARCHAR(10)    NOT NULL,

    passenger_name     VARCHAR(255)   NOT NULL,
    passenger_passport VARCHAR(50)    NOT NULL,

    price              DECIMAL(10, 2) NOT NULL,
    currency           VARCHAR(3)     NOT NULL  DEFAULT 'RUB',

    status             VARCHAR(20)    NOT NULL  DEFAULT 'PENDING',

    created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user_id ON bookings (user_id);