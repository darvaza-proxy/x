// Package synctesting provides the benchmark reporting shared by the
// TryLock-style benchmarks in darvaza.org/x/sync.
//
// Channel and timing assertions come from darvaza.org/core, which names
// what a failing assertion saw. This package holds what has no home
// there.
//
// # Helpers
//
//   - ReportTryMetrics: report the attempts-per-acquisition,
//     acquisitions-per-second and nanoseconds-per-attempt ratios of a
//     TryLock benchmark, skipping the ratios a zero count, zero attempt
//     total or zero elapsed time would make meaningless.
//   - MetricReporter: the *testing.B subset ReportTryMetrics needs,
//     taken as an interface so a test can record what was emitted.
package synctesting
