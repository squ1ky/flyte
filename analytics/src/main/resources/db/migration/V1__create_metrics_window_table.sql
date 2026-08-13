CREATE TABLE metrics_windows
(
    window_start                   TIMESTAMPTZ NOT NULL,
    window_size_seconds            INT         NOT NULL,

    created_count                  INT         NOT NULL DEFAULT 0,
    paid_count                     INT         NOT NULL DEFAULT 0,
    cancelled_expired_count        INT         NOT NULL DEFAULT 0,
    cancelled_payment_failed_count INT         NOT NULL DEFAULT 0,
    cancelled_user_cancelled_count INT         NOT NULL DEFAULT 0,

    updated_at                     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (window_start, window_size_seconds)
);

CREATE INDEX idx_metrics_windows_updated_at ON metrics_windows (updated_at);