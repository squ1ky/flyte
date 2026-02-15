CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255)        NOT NULL,
    phone_number  VARCHAR(20),
    role          VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE passenger_profiles
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT       NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    middle_name     VARCHAR(100),
    birth_date      DATE         NOT NULL,
    gender          VARCHAR(10) CHECK (gender IN ('male', 'female')),
    document_number VARCHAR(50)  NOT NULL,
    document_type   VARCHAR(20) DEFAULT 'passport',
    citizenship     CHAR(3)      NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profiles_user_id ON passenger_profiles (user_id);
CREATE UNIQUE INDEX idx_passenger_profiles_user_unique_doc ON passenger_profiles (user_id, document_number, document_type, citizenship);