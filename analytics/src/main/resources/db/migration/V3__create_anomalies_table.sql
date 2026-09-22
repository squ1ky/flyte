CREATE TABLE anomalies
(
    id           BIGSERIAL PRIMARY KEY,
    algorithm    VARCHAR(50)      NOT NULL,
    approach     VARCHAR(20)      NOT NULL,
    metric_name  VARCHAR(100)     NOT NULL,
    window_start TIMESTAMPTZ      NOT NULL,
    detected_at  TIMESTAMPTZ      NOT NULL DEFAULT now(),
    score        DOUBLE PRECISION NOT NULL,
    is_anomaly   BOOLEAN          NOT NULL DEFAULT true,

    CONSTRAINT chk_anomalies_approach CHECK (approach IN ('stream', 'batch'))
);

CREATE INDEX idx_anomalies_detected_at ON anomalies (detected_at);
CREATE INDEX idx_anomalies_metric_algorithm ON anomalies (metric_name, algorithm, approach);