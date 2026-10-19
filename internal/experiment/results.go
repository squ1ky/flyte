package experiment

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	PhaseBaseline = "baseline"
	PhaseFault    = "fault"

	StatusDetected      = "detected"
	StatusTimeout       = "timeout"
	StatusFalsePositive = "false_positive"
)

// ResultRow is one line of the results CSV: one detector's outcome in one phase of one run.
type ResultRow struct {
	ExperimentID string
	Scenario     string
	Run          int
	Phase        string
	Status       string
	Monitor      Monitor
	Expected     bool
	// T0 is the injection moment (fault phase) or baseline start (baseline phase), host clock.
	T0 time.Time
	// T1 is detected_at converted to host clock; zero on timeout.
	T1          time.Time
	WindowStart time.Time
	Score       float64
	ClockSkew   time.Duration
}

var resultHeader = []string{
	"experiment_id", "scenario", "run", "phase", "status", "metric", "approach", "algorithm", "expected",
	"t0", "t1", "delta_ms", "window_start", "score", "clock_skew_ms",
}

// ResultWriter appends rows and flushes after each one, so a crashed or interrupted
// experiment keeps every run completed so far.
type ResultWriter struct {
	file *os.File
	csv  *csv.Writer
}

func NewResultWriter(path string) (*ResultWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := &ResultWriter{file: f, csv: csv.NewWriter(f)}
	if err := w.writeRecord(resultHeader); err != nil {
		f.Close()
		return nil, err
	}
	return w, nil
}

func (w *ResultWriter) Write(r ResultRow) error {
	t1, delta, windowStart, score := "", "", "", ""
	if !r.T1.IsZero() {
		t1 = r.T1.UTC().Format(time.RFC3339Nano)
		delta = strconv.FormatInt(r.T1.Sub(r.T0).Milliseconds(), 10)
		windowStart = r.WindowStart.UTC().Format(time.RFC3339)
		score = strconv.FormatFloat(r.Score, 'f', 4, 64)
	}
	return w.writeRecord([]string{
		r.ExperimentID, r.Scenario, strconv.Itoa(r.Run), r.Phase, r.Status,
		r.Monitor.MetricName, r.Monitor.Approach, r.Monitor.Algorithm, strconv.FormatBool(r.Expected),
		r.T0.UTC().Format(time.RFC3339Nano), t1, delta, windowStart, score,
		strconv.FormatInt(r.ClockSkew.Milliseconds(), 10),
	})
}

func (w *ResultWriter) writeRecord(record []string) error {
	if err := w.csv.Write(record); err != nil {
		return err
	}
	w.csv.Flush()
	if err := w.csv.Error(); err != nil {
		return fmt.Errorf("write results: %w", err)
	}
	return nil
}

func (w *ResultWriter) Close() error {
	w.csv.Flush()
	return w.file.Close()
}
