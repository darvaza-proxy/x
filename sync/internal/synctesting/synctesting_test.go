package synctesting_test

import (
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/x/sync/internal/synctesting"
)

// metricRecorder captures ReportMetric calls keyed by unit label.
type metricRecorder map[string]float64

func (r metricRecorder) ReportMetric(n float64, unit string) { r[unit] = n }

var _ synctesting.MetricReporter = metricRecorder{}

// reportTryMetricsTestCase exercises ReportTryMetrics across its guard
// clauses. The unit noun is fixed to "lock": it only feeds string
// concatenation, never a branch. want lists every metric the call must
// emit; absent keys must not be emitted at all.
type reportTryMetricsTestCase struct {
	want map[string]float64

	name string

	elapsedMS int

	attempts int32
	count    int32
}

func newReportTryMetricsTestCase(name string, attempts, count int32,
	elapsedMS int, want map[string]float64) reportTryMetricsTestCase {
	return reportTryMetricsTestCase{
		name:      name,
		attempts:  attempts,
		count:     count,
		elapsedMS: elapsedMS,
		want:      want,
	}
}

func (tc reportTryMetricsTestCase) Name() string { return tc.name }

func (tc reportTryMetricsTestCase) Test(t *testing.T) {
	t.Helper()
	got := metricRecorder{}
	elapsed := time.Duration(tc.elapsedMS) * time.Millisecond
	synctesting.ReportTryMetrics(got, tc.attempts, tc.count, elapsed, "lock")
	core.AssertEqual(t, len(tc.want), len(got), "metric count")
	for unit, want := range tc.want {
		core.AssertEqual(t, want, got[unit], unit)
	}
}

var _ core.TestCase = reportTryMetricsTestCase{}

func TestReportTryMetrics(t *testing.T) {
	core.RunTestCases(t, []reportTryMetricsTestCase{
		newReportTryMetricsTestCase("no acquisitions", 10, 0, 1000,
			map[string]float64{}),
		newReportTryMetricsTestCase("no attempts", 0, 5, 1000,
			map[string]float64{}),
		newReportTryMetricsTestCase("zero elapsed", 10, 5, 0,
			map[string]float64{"attempts/lock": 2}),
		newReportTryMetricsTestCase("happy path", 10, 5, 2000,
			map[string]float64{
				"attempts/lock": 2,
				"locks/sec":     2.5,
				"ns/attempt":    2e8,
			}),
	})
}
