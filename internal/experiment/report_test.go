package experiment

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleResults = `experiment_id,scenario,run,phase,status,metric,approach,algorithm,expected,t0,t0_effect,t1,delta_ms,delta_effect_ms,window_start,score,clock_skew_ms
e1,payment_outage,1,fault,detected,conversion_rate,stream,cusum,true,t,te,t,22000,10000,w,5.1,0
e1,payment_outage,1,fault,detected,conversion_rate,batch,cusum,true,t,te,t,82000,70000,w,5.1,0
e1,payment_outage,2,baseline,false_positive,conversion_rate,stream,cusum,true,t,,t,500,,w,5.1,0
e1,payment_outage,2,fault,detected,conversion_rate,stream,cusum,true,t,te,t,32000,20000,w,5.1,0
e1,payment_outage,2,fault,timeout,conversion_rate,batch,cusum,true,t,te,,,,,,0
e1,payment_outage,3,fault,detected,conversion_rate,stream,cusum,true,t,te,t,42000,30000,w,5.1,0
e1,payment_outage,3,fault,detected,conversion_rate,batch,cusum,true,t,te,t,102000,90000,w,5.1,0
e1,payment_outage,3,fault,detected,abandonment_rate_expired,stream,cusum,false,t,te,t,13000,1000,w,3.1,0
`

func TestBuildReport(t *testing.T) {
	rows, err := parseResults(strings.NewReader(sampleResults))
	require.NoError(t, err)

	report := BuildReport(rows)

	// Unexpected detections (abandonment in payment_outage) do not count as latency.
	require.Len(t, report.Latency, 2)
	stream, batch := report.Latency[0], report.Latency[1]
	assert.Equal(t, "stream", stream.Monitor.Approach)
	assert.Equal(t, 3, stream.Detected)
	assert.Equal(t, 0, stream.Timeouts)
	assert.InDelta(t, 20.0, stream.Median, 1e-9)
	assert.InDelta(t, 10.0, stream.Std, 1e-9)
	assert.InDelta(t, 15.0, stream.P25, 1e-9)
	assert.InDelta(t, 32.0, stream.MedianInj, 1e-9)
	assert.InDelta(t, 32.0, stream.MeanInj, 1e-9)

	assert.Equal(t, "batch", batch.Monitor.Approach)
	assert.Equal(t, 2, batch.Detected)
	assert.Equal(t, 1, batch.Timeouts)
	assert.InDelta(t, 80.0, batch.Median, 1e-9)
	assert.InDelta(t, 92.0, batch.MedianInj, 1e-9)

	require.Len(t, report.Overhead, 1)
	assert.Equal(t, "payment_outage", report.Overhead[0].Scenario)
	assert.Equal(t, 3, report.Overhead[0].Runs)
	assert.InDelta(t, 12.0, report.Overhead[0].Median, 1e-9)
	assert.InDelta(t, 12.0, report.Overhead[0].Mean, 1e-9)

	require.Len(t, report.StreamVsBatch, 1)
	assert.Equal(t, "cusum", report.StreamVsBatch[0].Algorithm)

	fp := map[Monitor]FalsePositiveStats{}
	for _, f := range report.FalsePositives {
		fp[f.Monitor] = f
	}
	streamCusum := fp[Monitor{MetricName: "conversion_rate", Algorithm: "cusum", Approach: "stream"}]
	assert.Equal(t, 3, streamCusum.BaselineRuns)
	assert.Equal(t, 1, streamCusum.FalsePositives)

	var md strings.Builder
	require.NoError(t, report.WriteMarkdown(&md))
	assert.Contains(t, md.String(), "| payment_outage | conversion_rate | cusum | 20.0 | 80.0 | 60.0 | 4.0x |")
}

func TestParseResultsRejectsForeignCSV(t *testing.T) {
	_, err := parseResults(strings.NewReader("a,b\n1,2\n"))
	assert.Error(t, err)
}
