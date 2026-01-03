CREATE TABLE IF NOT EXISTS users
(
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    phone_number  VARCHAR(20),
    role          VARCHAR(20) DEFAULT 'user',
    created_at    TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS passengers
(
    id              SERIAL PRIMARY KEY,
    user_id         INT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    middle_name     VARCHAR(100),
    birth_date      DATE         NOT NULL,
    gender          VARCHAR(10) CHECK (gender IN ('male', 'female')),

    document_number VARCHAR(50)  NOT NULL,
    document_type   VARCHAR(20) DEFAULT 'passport',
    citizenship     VARCHAR(3),

    created_at      TIMESTAMP   DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_passengers_user_id ON passengers (user_id);
