package tai

import (
	"math"
	"testing"
	"time"

	"darvaza.org/core"
)

// Compile-time verification that test case types implement core.TestCase
var (
	_ core.TestCase = tainaAddTestCase{}
	_ core.TestCase = attosecondSplitTestCase{}
	_ core.TestCase = unixTAINATestCase{}
)

// tainaAddTestCase tests TAINA.Add overflow behaviour.
type tainaAddTestCase struct {
	expectErr error
	name      string
	tai       TAINA
	duration  time.Duration
}

func newTAINAAddTestCase(name string, tai TAINA, duration time.Duration,
	expectErr error) tainaAddTestCase {
	return tainaAddTestCase{
		name:      name,
		tai:       tai,
		duration:  duration,
		expectErr: expectErr,
	}
}

// Name returns the test case name for identification.
func (tc tainaAddTestCase) Name() string {
	return tc.name
}

// Test validates TAINA.Add overflow behaviour.
func (tc tainaAddTestCase) Test(t *testing.T) {
	t.Helper()

	_, err := tc.tai.Add(tc.duration)
	if tc.expectErr != nil {
		core.AssertErrorIs(t, err, tc.expectErr, "add error")
		return
	}
	core.AssertNoError(t, err, "add")
}

func TestTAINAArithmeticOverflow(t *testing.T) {
	testCases := []tainaAddTestCase{
		newTAINAAddTestCase("overflow near max uint64",
			TAINA{sec: math.MaxUint64 - 100}, 200*time.Second, ErrOverflow),
		newTAINAAddTestCase("underflow near zero",
			TAINA{sec: 100}, -200*time.Second, ErrUnderflow),
		newTAINAAddTestCase("normal values",
			TAINA{sec: 1000, nano: 500000000}, time.Hour, nil),
		newTAINAAddTestCase("nanosecond carry overflow at max",
			TAINA{sec: math.MaxUint64, nano: 999999999}, time.Nanosecond, ErrOverflow),
		newTAINAAddTestCase("nanosecond borrow underflow at zero",
			TAINA{}, -time.Nanosecond, ErrUnderflow),
	}

	core.RunTestCases(t, testCases)
}

// attosecondSplitTestCase tests TAINA.UnixAttosecondSplit.
type attosecondSplitTestCase struct {
	name              string
	tai               TAINA
	expectNanoseconds int64
	expectAttoseconds uint32
}

func newAttosecondSplitTestCase(name string, tai TAINA, expectNanoseconds int64,
	expectAttoseconds uint32) attosecondSplitTestCase {
	return attosecondSplitTestCase{
		name:              name,
		tai:               tai,
		expectNanoseconds: expectNanoseconds,
		expectAttoseconds: expectAttoseconds,
	}
}

// Name returns the test case name for identification.
func (tc attosecondSplitTestCase) Name() string {
	return tc.name
}

// Test validates the split timestamp representation.
func (tc attosecondSplitTestCase) Test(t *testing.T) {
	t.Helper()

	result := tc.tai.UnixAttosecondSplit()
	core.AssertEqual(t, tc.expectNanoseconds, result.UnixNanoseconds, "nanoseconds")
	core.AssertEqual(t, tc.expectAttoseconds, result.Attoseconds, "attoseconds")
}

func TestTAINAUnixAttosecondSplit(t *testing.T) {
	testCases := []attosecondSplitTestCase{
		newAttosecondSplitTestCase("large timestamp",
			TAINA{sec: TAICONST + 100000, nano: 999999999, atto: 123456789},
			100000*1e9+999999999, 123456789),
		newAttosecondSplitTestCase("normal timestamp",
			TAINA{sec: TAICONST + 1000, nano: 500000000, atto: 987654321},
			1000*1e9+500000000, 987654321),
		newAttosecondSplitTestCase("zero timestamp",
			TAINA{sec: TAICONST}, 0, 0),
		newAttosecondSplitTestCase("small timestamp",
			TAINA{sec: TAICONST + 1, nano: 123456789, atto: 555555555},
			1*1e9+123456789, 555555555),
	}

	core.RunTestCases(t, testCases)
}

func TestTAINAAddAttosecondsOverflow(t *testing.T) {
	t.Run("large positive attoseconds", runTestLargePositiveAttoseconds)
	t.Run("large negative attoseconds", runTestLargeNegativeAttoseconds)
	t.Run("normal attoseconds", runTestNormalAttoseconds)
}

func runTestLargePositiveAttoseconds(t *testing.T) {
	t.Helper()

	base := TAINA{sec: 1000, nano: 500000000, atto: 500000000}
	result, err := base.AddAttoseconds(math.MaxInt64 / 2)
	core.AssertMustNoError(t, err, "add")
	core.AssertTrue(t, result.sec > base.sec, "seconds increased")
}

func runTestLargeNegativeAttoseconds(t *testing.T) {
	t.Helper()

	base := TAINA{sec: 1000, nano: 500000000, atto: 500000000}
	result, err := base.AddAttoseconds(-math.MaxInt64 / 2)
	core.AssertMustNoError(t, err, "add")
	core.AssertTrue(t, result.sec < base.sec, "seconds decreased")
}

func runTestNormalAttoseconds(t *testing.T) {
	t.Helper()

	base := TAINA{sec: 1000, nano: 500000000, atto: 500000000}
	_, err := base.AddAttoseconds(1000000000)
	core.AssertNoError(t, err, "add")
}

func TestTAINAExtremeValues(t *testing.T) {
	t.Run("maximum values", runTestMaximumValues)
	t.Run("minimum values", runTestMinimumValues)
}

func runTestMaximumValues(t *testing.T) {
	t.Helper()

	tai := TAINA{sec: math.MaxUint64, nano: 999999999, atto: 999999999}
	core.AssertNoPanic(t, func() {
		_ = tai.Unix()
		_ = tai.UnixMilli()
		_ = tai.UnixMicro()
		_ = tai.UnixNano()
	}, "unix accessors")
}

func runTestMinimumValues(t *testing.T) {
	t.Helper()

	tai := TAINA{}
	core.AssertEqual(t, -int64(TAICONST), tai.Unix(), "unix seconds")

	split := tai.UnixAttosecondSplit()
	core.AssertEqual(t, tai.UnixNano(), split.UnixNanoseconds, "split nanoseconds")
	core.AssertEqual(t, tai.atto, split.Attoseconds, "split attoseconds")
}

// unixTAINATestCase tests UnixTAINA constructor validation.
type unixTAINATestCase struct {
	name        string
	sec         int64
	nsec        int64
	asec        uint32
	expectPanic bool
}

func newUnixTAINATestCase(name string, sec, nsec int64, asec uint32,
	expectPanic bool) unixTAINATestCase {
	return unixTAINATestCase{
		name:        name,
		sec:         sec,
		nsec:        nsec,
		asec:        asec,
		expectPanic: expectPanic,
	}
}

// Name returns the test case name for identification.
func (tc unixTAINATestCase) Name() string {
	return tc.name
}

// Test validates UnixTAINA attosecond range validation.
func (tc unixTAINATestCase) Test(t *testing.T) {
	t.Helper()

	if tc.expectPanic {
		core.AssertPanic(t, func() {
			UnixTAINA(tc.sec, tc.nsec, tc.asec)
		}, ErrInvalidAttosecondRange, "constructor")
		return
	}

	core.AssertNoPanic(t, func() {
		_ = UnixTAINA(tc.sec, tc.nsec, tc.asec)
	}, "constructor")
}

func TestTAINAConstructorValidation(t *testing.T) {
	testCases := []unixTAINATestCase{
		newUnixTAINATestCase("valid attoseconds", 1000, 0, 999999999, false),
		newUnixTAINATestCase("attoseconds too large", 0, 0, 1000000000, true),
		newUnixTAINATestCase("zero attoseconds", 0, 0, 0, false),
	}

	core.RunTestCases(t, testCases)
}

func TestTAINAArithmeticEdgeCases(t *testing.T) {
	t.Run("subtraction at nanosecond boundary", runTestSubNanosecondBoundary)
	t.Run("addition crossing second boundary", runTestAddSecondBoundary)
}

func runTestSubNanosecondBoundary(t *testing.T) {
	t.Helper()

	t1 := TAINA{sec: 1000, nano: 999999999, atto: 999999999}
	t2 := TAINA{sec: 1000}
	core.AssertEqual(t, 999999999*time.Nanosecond, t1.Sub(t2), "difference")
}

func runTestAddSecondBoundary(t *testing.T) {
	t.Helper()

	base := TAINA{sec: 1000}
	result, err := base.Add(time.Second + time.Nanosecond)
	core.AssertMustNoError(t, err, "add")
	core.AssertEqual(t, uint64(1001), result.sec, "seconds")
	core.AssertEqual(t, uint32(1), result.nano, "nanoseconds")
}
