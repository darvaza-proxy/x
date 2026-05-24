package cond

import (
	"testing"

	"darvaza.org/core"
)

// makeAnyMatchIntTestCase exercises makeAnyMatch[int] against a slice of
// inputs and verifies the boolean output for each one.
type makeAnyMatchIntTestCase struct {
	name     string
	funcs    []func(int) bool
	inputs   []int
	expected []bool
}

func newMakeAnyMatchIntTestCase(name string, funcs []func(int) bool,
	inputs []int, expected []bool) makeAnyMatchIntTestCase {
	return makeAnyMatchIntTestCase{
		name:     name,
		funcs:    funcs,
		inputs:   inputs,
		expected: expected,
	}
}

func (tc makeAnyMatchIntTestCase) Name() string { return tc.name }

func (tc makeAnyMatchIntTestCase) Test(t *testing.T) {
	t.Helper()
	fn := makeAnyMatch(tc.funcs)
	core.AssertMustNotNil(t, fn, "result")
	core.AssertMustEqual(t, len(tc.inputs), len(tc.expected), "table shape")

	for i, input := range tc.inputs {
		got := fn(input)
		core.AssertEqual(t, tc.expected[i], got, "input[%d]=%d", i, input)
	}
}

var _ core.TestCase = makeAnyMatchIntTestCase{}

func makeAnyMatchIntTestCases() []makeAnyMatchIntTestCase {
	gtZero := func(n int) bool { return n > 0 }
	ltZero := func(n int) bool { return n < 0 }
	gtTen := func(n int) bool { return n > 10 }

	return []makeAnyMatchIntTestCase{
		newMakeAnyMatchIntTestCase("empty slice always true",
			[]func(int) bool{},
			[]int{0, 1, -1, 42},
			[]bool{true, true, true, true}),
		newMakeAnyMatchIntTestCase("all-nil entries treated as empty",
			[]func(int) bool{nil, nil},
			[]int{0, 1, -1},
			[]bool{true, true, true}),
		newMakeAnyMatchIntTestCase("nil entries mixed with real predicate",
			[]func(int) bool{nil, gtZero, nil},
			[]int{-1, 0, 1},
			[]bool{false, false, true}),
		newMakeAnyMatchIntTestCase("single predicate",
			[]func(int) bool{gtZero},
			[]int{-1, 0, 1},
			[]bool{false, false, true}),
		newMakeAnyMatchIntTestCase("any-of two predicates",
			[]func(int) bool{ltZero, gtTen},
			[]int{-5, 0, 5, 15},
			[]bool{true, false, false, true}),
		newMakeAnyMatchIntTestCase("nils stripped between multiple real predicates",
			[]func(int) bool{nil, ltZero, nil, gtTen, nil},
			[]int{-5, 0, 15},
			[]bool{true, false, true}),
	}
}

// TestMakeAnyMatchInt verifies makeAnyMatch over int predicates.
func TestMakeAnyMatchInt(t *testing.T) {
	core.RunTestCases(t, makeAnyMatchIntTestCases())
}

// TestMakeAnyMatchShortCircuit confirms makeAnyMatch evaluates predicates
// in order and stops at the first match. Separated from the table-driven
// test because it asserts evaluation strategy, not the input/output spec.
func TestMakeAnyMatchShortCircuit(t *testing.T) {
	var trapHit bool
	matches := func(n int) bool { return n < 0 }
	trap := func(_ int) bool {
		trapHit = true
		return true
	}

	fn := makeAnyMatch([]func(int) bool{matches, trap})
	core.AssertEqual(t, true, fn(-1), "result")
	core.AssertEqual(t, false, trapHit, "trap not triggered")
}

// makeAnyMatchStringTestCase exercises makeAnyMatch over the string type to
// confirm the generic implementation works for non-numeric values.
type makeAnyMatchStringTestCase struct {
	name     string
	funcs    []func(string) bool
	inputs   []string
	expected []bool
}

func newMakeAnyMatchStringTestCase(name string, funcs []func(string) bool,
	inputs []string, expected []bool) makeAnyMatchStringTestCase {
	return makeAnyMatchStringTestCase{
		name:     name,
		funcs:    funcs,
		inputs:   inputs,
		expected: expected,
	}
}

func (tc makeAnyMatchStringTestCase) Name() string { return tc.name }

func (tc makeAnyMatchStringTestCase) Test(t *testing.T) {
	t.Helper()
	fn := makeAnyMatch(tc.funcs)
	core.AssertMustNotNil(t, fn, "result")
	core.AssertMustEqual(t, len(tc.inputs), len(tc.expected), "table shape")

	for i, input := range tc.inputs {
		got := fn(input)
		core.AssertEqual(t, tc.expected[i], got, "input[%d]=%q", i, input)
	}
}

var _ core.TestCase = makeAnyMatchStringTestCase{}

func makeAnyMatchStringTestCases() []makeAnyMatchStringTestCase {
	longerThanFive := func(s string) bool { return len(s) > 5 }
	exactMatch := func(s string) bool { return s == "match" }

	return []makeAnyMatchStringTestCase{
		newMakeAnyMatchStringTestCase("either length or exact match",
			[]func(string) bool{longerThanFive, exactMatch},
			[]string{"short", "veryLongString", "match"},
			[]bool{false, true, true}),
	}
}

// TestMakeAnyMatchString verifies makeAnyMatch works for non-int generic
// types.
func TestMakeAnyMatchString(t *testing.T) {
	core.RunTestCases(t, makeAnyMatchStringTestCases())
}

// isCancelledTestCase exercises isCancelled against different channel states.
// The setup closure builds the channel inside Test to keep the table rows
// free of side-effects at declaration time.
type isCancelledTestCase struct {
	name  string
	setup func() <-chan struct{}
	want  bool
}

func newIsCancelledTestCase(name string, setup func() <-chan struct{},
	want bool) isCancelledTestCase {
	return isCancelledTestCase{name: name, setup: setup, want: want}
}

func (tc isCancelledTestCase) Name() string { return tc.name }

func (tc isCancelledTestCase) Test(t *testing.T) {
	t.Helper()
	ch := tc.setup()
	got := isCancelled(ch)
	core.AssertEqual(t, tc.want, got, "result")
}

var _ core.TestCase = isCancelledTestCase{}

func isCancelledTestCases() []isCancelledTestCase {
	return []isCancelledTestCase{
		newIsCancelledTestCase("nil channel reports not-cancelled",
			func() <-chan struct{} { return nil },
			false),
		newIsCancelledTestCase("open channel reports not-cancelled",
			func() <-chan struct{} { return make(chan struct{}) },
			false),
		newIsCancelledTestCase("closed channel reports cancelled",
			func() <-chan struct{} {
				ch := make(chan struct{})
				close(ch)
				return ch
			},
			true),
		newIsCancelledTestCase("buffered closed channel reports cancelled",
			func() <-chan struct{} {
				ch := make(chan struct{}, 1)
				close(ch)
				return ch
			},
			true),
		newIsCancelledTestCase("open buffered channel with queued value reports cancelled",
			func() <-chan struct{} {
				ch := make(chan struct{}, 1)
				ch <- struct{}{}
				return ch
			},
			true),
	}
}

// TestIsCancelled verifies the abort-channel reader behaves correctly for
// every receivable state: nil, open-empty, closed-unbuffered,
// closed-buffered, and open-buffered with a pending value. cond's
// production call sites only ever pass close-only signals, but the helper
// returns true for any "would receive without blocking" state.
func TestIsCancelled(t *testing.T) {
	core.RunTestCases(t, isCancelledTestCases())
}

// BenchmarkMakeAnyMatch measures the performance of makeAnyMatch with
// varying numbers of predicates.
func BenchmarkMakeAnyMatch(b *testing.B) {
	predicates := []func(int) bool{
		func(n int) bool { return n < 0 },
		func(n int) bool { return n > 100 },
		func(n int) bool { return n%2 == 0 },
		func(n int) bool { return n%3 == 0 },
		func(n int) bool { return n%5 == 0 },
	}

	benchCases := []struct {
		name string
		n    int
	}{
		{"Single", 1},
		{"Two", 2},
		{"Five", 5},
	}

	for _, bc := range benchCases {
		funcs := predicates[:bc.n]
		resultFunc := makeAnyMatch(funcs)

		b.Run(bc.name, func(b *testing.B) {
			for i := range b.N {
				_ = resultFunc(i)
			}
		})
	}
}
