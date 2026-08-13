CREATE TABLE payment_error_profile
(
    window_start        TIMESTAMPTZ  NOT NULL,
    window_size_seconds INT          NOT NULL,
    error_message       VARCHAR(500) NOT NULL,
    error_count         INT          NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),

    PRIMARY KEY (window_start, window_size_seconds, error_message)
);

CREATE INDEX idx_payment_error_profile_updated_at ON payment_error_profile (updated_at);