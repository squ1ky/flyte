package experiment

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleResults = `experiment_id,scenario,run,phase,status,metric,approach,algorithm,expected,t0,t1,delta_ms,window_start,score,clock_skew_ms
e1,payment_outage,1,fault,detected,conversion_rate,stream,cusum,true,t,t,10000,w,5.1,0
e1,payment_outage,1,fault,detected,conversion_rate,batch,cusum,true,t,t,70000,w,5.1,0
e1,payment_outage,2,baseline,false_positive,conversion_rate,stream,cusum,true,t,t,500,w,5.1,0
e1,payment_outage,2,fault,detected,conversion_rate,stream,cusum,true,t,t,20000,w,5.1,0
e1,payment_outage,2,fault,timeout,conversion_rate,batch,cusum,true,t,,,,,0
e1,payment_outage,3,fault,detected,conversion_rate,stream,cusum,true,t,t,30000,w,5.1,0
e1,payment_outage,3,fault,detected,conversion_rate,batch,cusum,true,t,t,90000,w,5.1,0
e1,payment_outage,3,fault,detected,abandonment_rate_expired,stream,cusum,false,t,t,1000,w,3.1,0
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

	assert.Equal(t, "batch", batch.Monitor.Approach)
	assert.Equal(t, 2, batch.Detected)
	assert.Equal(t, 1, batch.Timeouts)
	assert.InDelta(t, 80.0, batch.Median, 1e-9)

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
