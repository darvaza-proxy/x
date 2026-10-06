package errors_test

import (
	"testing"

	"darvaza.org/x/sync/errors"
)

// BenchmarkCompoundError measures recording an error in a new
// CompoundError, and reading one that holds two: its errors, its text
// and OK, the last alone and under RunParallel.
func BenchmarkCompoundError(b *testing.B) {
	b.Run("AppendError", runBenchmarkCompoundErrorAppendError)
	b.Run("Errors", runBenchmarkCompoundErrorErrors)
	b.Run("Error", runBenchmarkCompoundErrorError)
	b.Run("OK", runBenchmarkCompoundErrorOK)
	b.Run("OK_Parallel", runBenchmarkCompoundErrorOKParallel)
}

func runBenchmarkCompoundErrorAppendError(b *testing.B) {
	for b.Loop() {
		var ce errors.CompoundError
		if ce.AppendError(errOne) != &ce {
			b.Fatal("AppendError did not return its receiver")
		}
	}
}

func runBenchmarkCompoundErrorErrors(b *testing.B) {
	ce := newLoadedCompoundError(errOne, errTwo)

	for b.Loop() {
		ce.Errors()
	}
}

func runBenchmarkCompoundErrorError(b *testing.B) {
	ce := newLoadedCompoundError(errOne, errTwo)

	for b.Loop() {
		if ce.Error() == "" {
			b.Fatal("Error returned no text")
		}
	}
}

func runBenchmarkCompoundErrorOK(b *testing.B) {
	ce := newLoadedCompoundError(errOne, errTwo)

	for b.Loop() {
		ce.OK()
	}
}

func runBenchmarkCompoundErrorOKParallel(b *testing.B) {
	ce := newLoadedCompoundError(errOne, errTwo)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ce.OK()
		}
	})
}
