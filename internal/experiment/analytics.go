package experiment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// Monitor identifies one detector instance in analytics-service (one per metric x algorithm x approach).
type Monitor struct {
	MetricName string `json:"metricName"`
	Algorithm  string `json:"algorithm"`
	Approach   string `json:"approach"`
}

type Detection struct {
	Monitor
	WindowStart time.Time
	// DetectedAt is in analytics-db clock (anomalies.detected_at = now()).
	DetectedAt time.Time
	Score      float64
}

// AnalyticsClient controls detectors over REST and reads detections straight from analytics-db.
// Reading the table (not GET /api/anomalies) also lets the runner measure the db clock against
// the host clock, since t0 is taken on the host and t1 is stamped by Postgres.
type AnalyticsClient struct {
	baseURL string
	http    *http.Client
	db      *pgx.Conn
}

func NewAnalyticsClient(ctx context.Context, baseURL, dbURL string) (*AnalyticsClient, error) {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect analytics-db: %w", err)
	}
	return &AnalyticsClient{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}, db: conn}, nil
}

func (a *AnalyticsClient) Close(ctx context.Context) error {
	return a.db.Close(ctx)
}

func (a *AnalyticsClient) Monitors(ctx context.Context) ([]Monitor, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/api/anomalies/monitors", nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/anomalies/monitors: http %d", resp.StatusCode)
	}
	var out []Monitor
	return out, json.NewDecoder(resp.Body).Decode(&out)
}

// ResetDetectors zeroes every detector's state and clears its latch, so the next detection is recorded.
func (a *AnalyticsClient) ResetDetectors(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/anomalies/reset", nil)
	if err != nil {
		return err
	}
	resp, err := a.http.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("POST /api/anomalies/reset: http %d", resp.StatusCode)
	}
	return nil
}

// ClockSkew returns analytics-db clock minus host clock (round-trip midpoint estimate).
func (a *AnalyticsClient) ClockSkew(ctx context.Context) (time.Duration, error) {
	before := time.Now()
	var dbNow time.Time
	if err := a.db.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&dbNow); err != nil {
		return 0, err
	}
	after := time.Now()
	midpoint := before.Add(after.Sub(before) / 2)
	return dbNow.Sub(midpoint), nil
}

// DetectionsSince returns detections with detected_at >= since (db clock), oldest first.
func (a *AnalyticsClient) DetectionsSince(ctx context.Context, since time.Time) ([]Detection, error) {
	rows, err := a.db.Query(ctx, `
		SELECT metric_name, algorithm, approach, window_start, detected_at, score
		FROM anomalies
		WHERE detected_at >= $1 AND is_anomaly
		ORDER BY detected_at`, since)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Detection, error) {
		var d Detection
		err := row.Scan(&d.MetricName, &d.Algorithm, &d.Approach, &d.WindowStart, &d.DetectedAt, &d.Score)
		return d, err
	})
}
