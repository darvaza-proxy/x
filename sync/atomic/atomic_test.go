package atomic_test

import (
	"sync"
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/sync/atomic"
)

// bitmaskOrTestCase exercises BitmaskOr against a fresh atomic.Uint32 seeded
// with start.
type bitmaskOrTestCase struct {
	name        string
	start       uint32
	mask        uint32
	wantResult  uint32
	wantChanged bool
}

func newBitmaskOrTestCase(name string, start, mask, wantResult uint32,
	wantChanged bool) bitmaskOrTestCase {
	return bitmaskOrTestCase{
		name:        name,
		start:       start,
		mask:        mask,
		wantResult:  wantResult,
		wantChanged: wantChanged,
	}
}

func (tc bitmaskOrTestCase) Name() string { return tc.name }

func (tc bitmaskOrTestCase) Test(t *testing.T) {
	t.Helper()
	var p atomic.Uint32
	p.Store(tc.start)

	got, changed := atomic.BitmaskOr(&p, tc.mask)
	core.AssertEqual(t, tc.wantResult, got, "result")
	core.AssertEqual(t, tc.wantChanged, changed, "changed")
	core.AssertEqual(t, tc.wantResult, p.Load(), "stored value")
}

// Compile-time verification that BitmaskOr test cases implement TestCase.
var _ core.TestCase = bitmaskOrTestCase{}

func bitmaskOrTestCases() []bitmaskOrTestCase {
	return []bitmaskOrTestCase{
		newBitmaskOrTestCase("set first bit", 0, 0x01, 0x01, true),
		newBitmaskOrTestCase("set same bit again", 0x01, 0x01, 0x01, false),
		newBitmaskOrTestCase("set additional bit", 0x01, 0x02, 0x03, true),
		newBitmaskOrTestCase("partial overlap still changes", 0x01, 0x03, 0x03, true),
		newBitmaskOrTestCase("mask already covered", 0x0F, 0x05, 0x0F, false),
		newBitmaskOrTestCase("zero mask never changes", 0xAA, 0x00, 0xAA, false),
		newBitmaskOrTestCase("zero mask on zero value", 0x00, 0x00, 0x00, false),
		newBitmaskOrTestCase("all bits set", 0x00, 0xFFFFFFFF, 0xFFFFFFFF, true),
	}
}

func TestBitmaskOr(t *testing.T) {
	core.RunTestCases(t, bitmaskOrTestCases())
}

// TestBitmaskOrConcurrent races N goroutines each setting one unique bit and
// verifies (a) all bits end up set and (b) exactly N goroutines observed
// changed == true.
func TestBitmaskOrConcurrent(t *testing.T) {
	// n must be in [1, 32] — each goroutine's mask must target a
	// distinct bit for the changedCount == n invariant to hold, and
	// Uint32 has 32 bits. The const _ below enforces this at compile
	// time: it overflows uint32 at n > 32 and produces a negative
	// shift count at n == 0.
	const n = 16
	const _ uint32 = 1 << (n - 1)

	var p atomic.Uint32
	var changedCount atomic.Int32
	var wg sync.WaitGroup

	wg.Add(n)
	for i := range n {
		go func(bit int) {
			defer wg.Done()
			if _, changed := atomic.BitmaskOr(&p, uint32(1)<<bit); changed {
				changedCount.Add(1)
			}
		}(i)
	}
	wg.Wait()

	core.AssertEqual(t, (uint32(1)<<n)-1, p.Load(), "all bits set")
	core.AssertEqual(t, int32(n), changedCount.Load(), "changed count")
}

// updateMaxTestCase exercises UpdateMax against a fresh atomic.Int32 seeded
// with start.
type updateMaxTestCase struct {
	name    string
	start   int32
	val     int32
	wantVal int32
}

func newUpdateMaxTestCase(name string, start, val, wantVal int32) updateMaxTestCase {
	return updateMaxTestCase{name: name, start: start, val: val, wantVal: wantVal}
}

func (tc updateMaxTestCase) Name() string { return tc.name }

func (tc updateMaxTestCase) Test(t *testing.T) {
	t.Helper()
	var p atomic.Int32
	p.Store(tc.start)

	got := atomic.UpdateMax(&p, tc.val)
	core.AssertEqual(t, tc.wantVal, got, "return")
	core.AssertEqual(t, tc.wantVal, p.Load(), "stored value")
}

// Compile-time verification that UpdateMax test cases implement TestCase.
var _ core.TestCase = updateMaxTestCase{}

func updateMaxTestCases() []updateMaxTestCase {
	return []updateMaxTestCase{
		newUpdateMaxTestCase("greater wins", 5, 10, 10),
		newUpdateMaxTestCase("equal is no-op", 5, 5, 5),
		newUpdateMaxTestCase("lesser is no-op", 5, 3, 5),
		newUpdateMaxTestCase("from zero", 0, 7, 7),
		newUpdateMaxTestCase("negative greater than negative", -10, -3, -3),
		newUpdateMaxTestCase("negative lesser than negative", -3, -10, -3),
	}
}

func TestUpdateMax(t *testing.T) {
	core.RunTestCases(t, updateMaxTestCases())
}

// TestUpdateMaxConcurrent races N goroutines, each writing a distinct value;
// the final state must equal the maximum of those values, and every per-call
// return value must lie in [caller's val, final *p].
func TestUpdateMaxConcurrent(t *testing.T) {
	const n = 32
	var p atomic.Int32
	var wg sync.WaitGroup

	results := make([]int32, n)
	wg.Add(n)
	for i := range n {
		go func(v int32) {
			defer wg.Done()
			results[v] = atomic.UpdateMax(&p, v)
		}(int32(i))
	}
	wg.Wait()

	final := p.Load()
	core.AssertEqual(t, int32(n-1), final, "max value")
	for i, got := range results {
		in := int32(i)
		core.AssertTrue(t, got >= in, "result>=val for i=%d (got=%d)", i, got)
		core.AssertTrue(t, got <= final, "result<=final for i=%d (got=%d)", i, got)
	}
}
