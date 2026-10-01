package cond_test

import (
	"testing"

	"darvaza.org/x/sync/cond"
)

// BenchmarkTurnstile measures passing and holding a Turnstile, alone and
// under RunParallel, where every goroutine competes for the one token.
func BenchmarkTurnstile(b *testing.B) {
	b.Run("Pass", runBenchmarkTurnstilePass)
	b.Run("Pass_Parallel", runBenchmarkTurnstilePassParallel)
	b.Run("LockUnlock", runBenchmarkTurnstileLockUnlock)
	b.Run("LockUnlock_Parallel", runBenchmarkTurnstileLockUnlockParallel)
}

func runBenchmarkTurnstilePass(b *testing.B) {
	var ts cond.Turnstile

	for b.Loop() {
		ts.Pass()
	}
}

func runBenchmarkTurnstilePassParallel(b *testing.B) {
	var ts cond.Turnstile

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ts.Pass()
		}
	})
}

func runBenchmarkTurnstileLockUnlock(b *testing.B) {
	var ts cond.Turnstile
	var n int

	for b.Loop() {
		ts.Lock()
		n = 2*n + 1
		ts.Unlock()
	}
}

func runBenchmarkTurnstileLockUnlockParallel(b *testing.B) {
	var ts cond.Turnstile
	var n int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ts.Lock()
			n = 2*n + 1
			ts.Unlock()
		}
	})
}
