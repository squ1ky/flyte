package experiment

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ReportRow is the subset of a results CSV row (see resultHeader) the report needs.
type ReportRow struct {
	ExperimentID string
	Scenario     string
	Run          int
	Phase        string
	Status       string
	Monitor      Monitor
	Expected     bool
	DeltaMs      int64 // only for status=detected
	// DeltaEffectMs is t1 - t0_effect (first fault effect, see RESULTS.md 3.1);
	// HasEffect is false for CSVs without the effect-based columns (then DeltaMs is the only latency).
	DeltaEffectMs int64
	HasEffect     bool
}

// ReadResults loads rows from one or more results CSV files written by ResultWriter.
func ReadResults(paths ...string) ([]ReportRow, error) {
	var rows []ReportRow
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		fileRows, err := parseResults(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		rows = append(rows, fileRows...)
	}
	return rows, nil
}

func parseResults(r io.Reader) ([]ReportRow, error) {
	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	col := map[string]int{}
	for i, name := range records[0] {
		col[name] = i
	}
	for _, name := range resultHeader {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("missing column %q, not a results file?", name)
		}
	}

	rows := make([]ReportRow, 0, len(records)-1)
	for line, rec := range records[1:] {
		run, err := strconv.Atoi(rec[col["run"]])
		if err != nil {
			return nil, fmt.Errorf("line %d: run: %w", line+2, err)
		}
		row := ReportRow{
			ExperimentID: rec[col["experiment_id"]],
			Scenario:     rec[col["scenario"]],
			Run:          run,
			Phase:        rec[col["phase"]],
			Status:       rec[col["status"]],
			Monitor: Monitor{
				MetricName: rec[col["metric"]],
				Algorithm:  rec[col["algorithm"]],
				Approach:   rec[col["approach"]],
			},
			Expected: rec[col["expected"]] == "true",
		}
		if row.Status == StatusDetected {
			if row.DeltaMs, err = strconv.ParseInt(rec[col["delta_ms"]], 10, 64); err != nil {
				return nil, fmt.Errorf("line %d: delta_ms: %w", line+2, err)
			}
			if rec[col["delta_effect_ms"]] != "" {
				if row.DeltaEffectMs, err = strconv.ParseInt(rec[col["delta_effect_ms"]], 10, 64); err != nil {
					return nil, fmt.Errorf("line %d: delta_effect_ms: %w", line+2, err)
				}
				row.HasEffect = true
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// LatencyStats summarises detection times of one detector in one scenario, over all fault phases.
// Median/Mean/Std/P25/P75/Min/Max use the effect-based delta (t1 - t0_effect) when the CSV
// has it, and fall back to the injection-inclusive delta (t1 - t0) for old CSVs.
type LatencyStats struct {
	Scenario string
	Monitor  Monitor
	Detected int
	Timeouts int
	// Seconds, over detected runs only. Zero when nothing was detected.
	Median, Mean, Std, P25, P75, Min, Max float64
	// MedianInj/MeanInj are the injection-inclusive counterparts (t1 - t0), seconds.
	// Zero when the CSV has no effect-based deltas.
	MedianInj, MeanInj float64
}

// OverheadStats summarises the injection overhead t0_effect - t0 (container restart and
// Kafka reconnection between "faulty container is up" and "first payment processed by it"),
// one value per run, per scenario.
type OverheadStats struct {
	Scenario string
	Runs     int
	// Seconds.
	Median, Mean, Min, Max float64
}

// FalsePositiveStats counts baseline-phase detections of one detector in one scenario.
type FalsePositiveStats struct {
	Scenario       string
	Monitor        Monitor
	BaselineRuns   int
	FalsePositives int
}

// StreamBatchPair compares median detection times of the same metric and algorithm.
type StreamBatchPair struct {
	Scenario, Metric, Algorithm string
	Stream, Batch               *LatencyStats
}

type Report struct {
	Latency        []LatencyStats
	FalsePositives []FalsePositiveStats
	StreamVsBatch  []StreamBatchPair
	Overhead       []OverheadStats
}

type scenarioMonitor struct {
	Scenario string
	Monitor
}

func BuildReport(rows []ReportRow) Report {
	type runKey struct {
		experimentID string
		run          int
	}
	runsByScenario := map[string]map[runKey]bool{}
	monitorsByScenario := map[string]map[Monitor]bool{}
	deltas := map[scenarioMonitor][]float64{}    // primary latency: effect-based when available
	injDeltas := map[scenarioMonitor][]float64{} // injection-inclusive counterpart
	timeouts := map[scenarioMonitor]int{}
	falsePositives := map[scenarioMonitor]int{}
	overheadByScenario := map[string]map[runKey]float64{}

	for _, r := range rows {
		if runsByScenario[r.Scenario] == nil {
			runsByScenario[r.Scenario] = map[runKey]bool{}
			monitorsByScenario[r.Scenario] = map[Monitor]bool{}
		}
		runsByScenario[r.Scenario][runKey{r.ExperimentID, r.Run}] = true
		monitorsByScenario[r.Scenario][r.Monitor] = true

		key := scenarioMonitor{r.Scenario, r.Monitor}
		switch {
		case r.Phase == PhaseBaseline && r.Status == StatusFalsePositive:
			falsePositives[key]++
		case r.Phase == PhaseFault && r.Expected && r.Status == StatusDetected:
			if r.HasEffect {
				deltas[key] = append(deltas[key], float64(r.DeltaEffectMs)/1000)
				injDeltas[key] = append(injDeltas[key], float64(r.DeltaMs)/1000)
				if overheadByScenario[r.Scenario] == nil {
					overheadByScenario[r.Scenario] = map[runKey]float64{}
				}
				// t0 and t0_effect are common to every monitor in the run; keep the first value seen.
				rk := runKey{r.ExperimentID, r.Run}
				if _, seen := overheadByScenario[r.Scenario][rk]; !seen {
					overheadByScenario[r.Scenario][rk] = float64(r.DeltaMs-r.DeltaEffectMs) / 1000
				}
			} else {
				deltas[key] = append(deltas[key], float64(r.DeltaMs)/1000)
			}
		case r.Phase == PhaseFault && r.Expected && r.Status == StatusTimeout:
			timeouts[key]++
		}
	}

	var report Report
	latencyByKey := map[scenarioMonitor]*LatencyStats{}
	awaited := map[scenarioMonitor]bool{}
	for key := range deltas {
		awaited[key] = true
	}
	for key := range timeouts {
		awaited[key] = true
	}
	for _, key := range sortedKeys(awaited) {
		s := latencyStats(deltas[key])
		if inj := injDeltas[key]; len(inj) == len(deltas[key]) && len(inj) > 0 {
			si := latencyStats(inj)
			s.MedianInj, s.MeanInj = si.Median, si.Mean
		}
		s.Scenario, s.Monitor, s.Timeouts = key.Scenario, key.Monitor, timeouts[key]
		report.Latency = append(report.Latency, s)
	}
	for i := range report.Latency {
		s := &report.Latency[i]
		latencyByKey[scenarioMonitor{s.Scenario, s.Monitor}] = s
	}

	for scenario, monitors := range monitorsByScenario {
		for m := range monitors {
			report.FalsePositives = append(report.FalsePositives, FalsePositiveStats{
				Scenario: scenario, Monitor: m,
				BaselineRuns:   len(runsByScenario[scenario]),
				FalsePositives: falsePositives[scenarioMonitor{scenario, m}],
			})
		}
	}
	sort.Slice(report.FalsePositives, func(i, j int) bool {
		return lessScenarioMonitor(
			scenarioMonitor{report.FalsePositives[i].Scenario, report.FalsePositives[i].Monitor},
			scenarioMonitor{report.FalsePositives[j].Scenario, report.FalsePositives[j].Monitor})
	})

	seen := map[[3]string]bool{}
	for _, s := range report.Latency {
		id := [3]string{s.Scenario, s.Monitor.MetricName, s.Monitor.Algorithm}
		if seen[id] {
			continue
		}
		seen[id] = true
		stream := latencyByKey[scenarioMonitor{s.Scenario, Monitor{s.Monitor.MetricName, s.Monitor.Algorithm, "stream"}}]
		batch := latencyByKey[scenarioMonitor{s.Scenario, Monitor{s.Monitor.MetricName, s.Monitor.Algorithm, "batch"}}]
		if stream != nil && batch != nil {
			report.StreamVsBatch = append(report.StreamVsBatch, StreamBatchPair{
				Scenario: s.Scenario, Metric: s.Monitor.MetricName, Algorithm: s.Monitor.Algorithm,
				Stream: stream, Batch: batch,
			})
		}
	}

	for scenario, byRun := range overheadByScenario {
		values := make([]float64, 0, len(byRun))
		for _, v := range byRun {
			values = append(values, v)
		}
		s := latencyStats(values)
		report.Overhead = append(report.Overhead, OverheadStats{
			Scenario: scenario, Runs: len(values),
			Median: s.Median, Mean: s.Mean, Min: s.Min, Max: s.Max,
		})
	}
	sort.Slice(report.Overhead, func(i, j int) bool { return report.Overhead[i].Scenario < report.Overhead[j].Scenario })
	return report
}

func latencyStats(values []float64) LatencyStats {
	s := LatencyStats{Detected: len(values)}
	if len(values) == 0 {
		return s
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	var sum float64
	for _, v := range sorted {
		sum += v
	}
	s.Mean = sum / float64(len(sorted))
	if len(sorted) > 1 {
		var sq float64
		for _, v := range sorted {
			sq += (v - s.Mean) * (v - s.Mean)
		}
		s.Std = math.Sqrt(sq / float64(len(sorted)-1)) // sample standard deviation
	}
	s.Median = percentile(sorted, 0.5)
	s.P25 = percentile(sorted, 0.25)
	s.P75 = percentile(sorted, 0.75)
	s.Min, s.Max = sorted[0], sorted[len(sorted)-1]
	return s
}

// percentile uses linear interpolation between closest ranks; sorted must be non-empty and ascending.
func percentile(sorted []float64, p float64) float64 {
	pos := p * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}

func sortedKeys(set map[scenarioMonitor]bool) []scenarioMonitor {
	keys := make([]scenarioMonitor, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return lessScenarioMonitor(keys[i], keys[j]) })
	return keys
}

func lessScenarioMonitor(a, b scenarioMonitor) bool {
	if a.Scenario != b.Scenario {
		return a.Scenario < b.Scenario
	}
	if a.MetricName != b.MetricName {
		return a.MetricName < b.MetricName
	}
	if a.Algorithm != b.Algorithm {
		return a.Algorithm < b.Algorithm
	}
	return a.Approach > b.Approach // stream before batch
}

// WriteMarkdown renders the report as markdown tables (seconds).
// Latency is t1 - t0_effect (first fault effect) where available; the last two columns of
// the latency table give the injection-inclusive counterpart (t1 - t0), and the overhead
// table shows how large the injection mechanics (t0_effect - t0) were per run.
func (r Report) WriteMarkdown(w io.Writer) error {
	var b strings.Builder

	b.WriteString("## Detection latency, s (fault phase, expected detectors)\n\n")
	b.WriteString("latency = t1 - t0_effect (first fault effect); inj columns = t1 - t0 (injection, includes service restart)\n\n")
	b.WriteString("| scenario | metric | algorithm | approach | detected | timeouts | median | mean | std | p25 | p75 | min | max | median_inj | mean_inj |\n")
	b.WriteString("|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|\n")
	for _, s := range r.Latency {
		medianInj, meanInj := "-", "-"
		if s.MedianInj > 0 || s.MeanInj > 0 {
			medianInj = fmt.Sprintf("%.1f", s.MedianInj)
			meanInj = fmt.Sprintf("%.1f", s.MeanInj)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d | %d | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			s.Scenario, s.Monitor.MetricName, s.Monitor.Algorithm, s.Monitor.Approach, s.Detected, s.Timeouts,
			sec(s, s.Median), sec(s, s.Mean), sec(s, s.Std), sec(s, s.P25), sec(s, s.P75), sec(s, s.Min), sec(s, s.Max),
			medianInj, meanInj)
	}

	if len(r.Overhead) > 0 {
		b.WriteString("\n## Injection overhead (t0_effect - t0), s per run\n\n")
		b.WriteString("| scenario | runs | median | mean | min | max |\n")
		b.WriteString("|---|---|---|---|---|---|\n")
		for _, o := range r.Overhead {
			fmt.Fprintf(&b, "| %s | %d | %.1f | %.1f | %.1f | %.1f |\n",
				o.Scenario, o.Runs, o.Median, o.Mean, o.Min, o.Max)
		}
	}

	b.WriteString("\n## Stream vs batch, median s\n\n")
	b.WriteString("| scenario | metric | algorithm | stream | batch | batch - stream | batch / stream |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, p := range r.StreamVsBatch {
		diff, ratio := "-", "-"
		if p.Stream.Detected > 0 && p.Batch.Detected > 0 {
			diff = fmt.Sprintf("%.1f", p.Batch.Median-p.Stream.Median)
			if p.Stream.Median > 0 {
				ratio = fmt.Sprintf("%.1fx", p.Batch.Median/p.Stream.Median)
			}
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			p.Scenario, p.Metric, p.Algorithm, sec(*p.Stream, p.Stream.Median), sec(*p.Batch, p.Batch.Median), diff, ratio)
	}

	b.WriteString("\n## False positives in baseline periods\n\n")
	b.WriteString("| scenario | metric | algorithm | approach | baseline periods | false positives | rate |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, f := range r.FalsePositives {
		rate := 0.0
		if f.BaselineRuns > 0 {
			rate = float64(f.FalsePositives) / float64(f.BaselineRuns)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %d | %d | %.0f%% |\n",
			f.Scenario, f.Monitor.MetricName, f.Monitor.Algorithm, f.Monitor.Approach,
			f.BaselineRuns, f.FalsePositives, rate*100)
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// WriteLatencyCSV writes the latency table in machine-readable form (for plots in the thesis).
func (r Report) WriteLatencyCSV(w io.Writer) error {
	out := csv.NewWriter(w)
	header := []string{"scenario", "metric", "algorithm", "approach", "detected", "timeouts",
		"median_s", "mean_s", "std_s", "p25_s", "p75_s", "min_s", "max_s", "median_inj_s", "mean_inj_s"}
	if err := out.Write(header); err != nil {
		return err
	}
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
	for _, s := range r.Latency {
		if err := out.Write([]string{
			s.Scenario, s.Monitor.MetricName, s.Monitor.Algorithm, s.Monitor.Approach,
			strconv.Itoa(s.Detected), strconv.Itoa(s.Timeouts),
			f(s.Median), f(s.Mean), f(s.Std), f(s.P25), f(s.P75), f(s.Min), f(s.Max),
			f(s.MedianInj), f(s.MeanInj),
		}); err != nil {
			return err
		}
	}
	out.Flush()
	return out.Error()
}

func sec(s LatencyStats, v float64) string {
	if s.Detected == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f", v)
}
