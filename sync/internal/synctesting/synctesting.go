package synctesting

import "time"

// MetricReporter is the subset of *testing.B that ReportTryMetrics needs,
// accepted as an interface so tests can record the emitted metrics.
type MetricReporter interface {
	ReportMetric(n float64, unit string)
}

// ReportTryMetrics reports the attempts-per-acquisition, acquisitions-per-
// second and nanoseconds-per-attempt ratios for a TryLock-style benchmark.
// unit is the per-acquisition noun ("lock" or "rlock"). Guard clauses keep
// zero counts, zero attempts and zero elapsed time from emitting nonsense
// ratios.
func ReportTryMetrics(b MetricReporter, attempts, count int32,
	elapsed time.Duration, unit string) {
	if count <= 0 || attempts <= 0 {
		return
	}
	b.ReportMetric(float64(attempts)/float64(count), "attempts/"+unit)
	if elapsed <= 0 {
		return
	}
	b.ReportMetric(float64(count)/elapsed.Seconds(), unit+"s/sec")
	b.ReportMetric(float64(elapsed.Nanoseconds())/float64(attempts),
		"ns/attempt")
}
