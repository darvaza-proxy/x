package tai

import (
	"errors"
	"math"
	"testing"
	"time"
)

func checkAddError(t *testing.T, result TaiNA, err, expectErr error) {
	if expectErr != nil {
		if err == nil {
			t.Errorf("Expected error %v, got result: %v", expectErr, result)
		}
		if !errors.Is(err, expectErr) {
			t.Errorf("Expected error %v, got: %v", expectErr, err)
		}
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTaiNAArithmeticOverflow(t *testing.T) {
	tests := []struct {
		name      string
		tai       TaiNA
		duration  time.Duration
		expectErr error
	}{
		{
			name:      "Add overflow - seconds near max uint64",
			tai:       TaiNA{sec: math.MaxUint64 - 100, nano: 0, atto: 0},
			duration:  200 * time.Second,
			expectErr: ErrTaiNAOverflow,
		},
		{
			name:      "Add underflow - seconds near zero",
			tai:       TaiNA{sec: 100, nano: 0, atto: 0},
			duration:  -200 * time.Second,
			expectErr: ErrTaiNAUnderflow,
		},
		{
			name:      "Add success - normal values",
			tai:       TaiNA{sec: 1000, nano: 500000000, atto: 0},
			duration:  time.Hour,
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.tai.Add(tt.duration)
			checkAddError(t, result, err, tt.expectErr)
		})
	}
}

//revive:disable:cognitive-complexity
func TestTaiNAUnixAttosecondSplit(t *testing.T) {
	tests := []struct {
		name              string
		tai               TaiNA
		expectNanoseconds int64
		expectAttoseconds uint32
	}{
		{
			name:              "Large timestamp works with split",
			tai:               TaiNA{sec: TAICONST + 100000, nano: 999999999, atto: 123456789},
			expectNanoseconds: 100000*1e9 + 999999999,
			expectAttoseconds: 123456789,
		},
		{
			name:              "Normal timestamp",
			tai:               TaiNA{sec: TAICONST + 1000, nano: 500000000, atto: 987654321},
			expectNanoseconds: 1000*1e9 + 500000000,
			expectAttoseconds: 987654321,
		},
		{
			name:              "Zero timestamp",
			tai:               TaiNA{sec: TAICONST, nano: 0, atto: 0},
			expectNanoseconds: 0,
			expectAttoseconds: 0,
		},
		{
			name:              "Small timestamp",
			tai:               TaiNA{sec: TAICONST + 1, nano: 123456789, atto: 555555555},
			expectNanoseconds: 1*1e9 + 123456789,
			expectAttoseconds: 555555555,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tai.UnixAttosecondSplit()

			if result.UnixNanoseconds != tt.expectNanoseconds {
				t.Errorf("UnixNanoseconds = %d, want %d", result.UnixNanoseconds, tt.expectNanoseconds)
			}
			if result.Attoseconds != tt.expectAttoseconds {
				t.Errorf("Attoseconds = %d, want %d", result.Attoseconds, tt.expectAttoseconds)
			}
		})
	}
}

//revive:enable:cognitive-complexity

func checkSecondsIncrease(t *testing.T, result, base TaiNA) {
	if result.sec <= base.sec {
		t.Errorf("Expected seconds to increase from %d to >%d, got %d", base.sec, base.sec, result.sec)
	}
}

func checkSecondsDecrease(t *testing.T, result, base TaiNA) {
	if result.sec >= base.sec {
		t.Errorf("Expected seconds to decrease from %d to <%d, got %d", base.sec, base.sec, result.sec)
	}
}

//revive:disable:cognitive-complexity
func TestTaiNAAddAttosecondsOverflow(t *testing.T) {
	base := TaiNA{sec: 1000, nano: 500000000, atto: 500000000}

	t.Run("Large positive attoseconds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddAttoseconds() panicked unexpectedly: %v", r)
			}
		}()
		result, err := base.AddAttoseconds(int64(math.MaxInt64 / 2))
		if err != nil {
			t.Fatalf("AddAttoseconds() failed: %v", err)
		}
		checkSecondsIncrease(t, result, base)
	})

	t.Run("Large negative attoseconds", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddAttoseconds() panicked unexpectedly: %v", r)
			}
		}()
		result, err := base.AddAttoseconds(int64(-math.MaxInt64 / 2))
		if err != nil {
			t.Fatalf("AddAttoseconds() failed: %v", err)
		}
		checkSecondsDecrease(t, result, base)
	})

	t.Run("Normal attoseconds addition", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddAttoseconds() panicked unexpectedly: %v", r)
			}
		}()
		_, err := base.AddAttoseconds(1000000000)
		if err != nil {
			t.Fatalf("AddAttoseconds() failed: %v", err)
		}
		// Normal addition should not significantly change seconds
	})
}

//revive:enable:cognitive-complexity

func testUnixOperation(t *testing.T, tai TaiNA, expectUnix int64) {
	result := tai.Unix()
	if result != expectUnix {
		t.Errorf("Unix() = %d, want %d", result, expectUnix)
	}
}

func testAttosecondSplitOperation(t *testing.T, tai TaiNA) {
	result := tai.UnixAttosecondSplit()
	expectedNanos := tai.UnixNano()
	expectedAttos := tai.atto
	if result.UnixNanoseconds != expectedNanos {
		t.Errorf("UnixAttosecondSplit().UnixNanoseconds = %d, want %d",
			result.UnixNanoseconds, expectedNanos)
	}
	if result.Attoseconds != expectedAttos {
		t.Errorf("UnixAttosecondSplit().Attoseconds = %d, want %d", result.Attoseconds, expectedAttos)
	}
}

func TestTaiNAExtremeValues(t *testing.T) {
	t.Run("Maximum values", func(_ *testing.T) {
		tai := TaiNA{sec: math.MaxUint64, nano: 999999999, atto: 999999999}
		_ = tai.Unix()
		_ = tai.UnixMilli()
		_ = tai.UnixMicro()
		_ = tai.UnixNano()
	})

	t.Run("Minimum values", func(t *testing.T) {
		tai := TaiNA{sec: 0, nano: 0, atto: 0}
		expectUnix := -int64(TAICONST)
		testUnixOperation(t, tai, expectUnix)
		testAttosecondSplitOperation(t, tai)
	})
}

func checkConstructorResult(t *testing.T, result TaiNA, sec, nsec int64, asec uint32) {
	if result.IsZero() && (sec != 0 || nsec != 0 || asec != 0) {
		t.Error("Constructor returned zero value for non-zero input")
	}
}

//revive:disable:cognitive-complexity
func TestTaiNAConstructorValidation(t *testing.T) {
	tests := []struct {
		name        string
		sec         int64
		nsec        int64
		asec        uint32
		expectPanic bool
	}{
		{
			name:        "Valid attoseconds",
			sec:         1000,
			nsec:        0,
			asec:        999999999,
			expectPanic: false,
		},
		{
			name:        "Invalid attoseconds - too large",
			sec:         0,
			nsec:        0,
			asec:        1000000000,
			expectPanic: true,
		},
		{
			name:        "Zero attoseconds",
			sec:         0,
			nsec:        0,
			asec:        0,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.expectPanic && r == nil {
					t.Error("Expected panic but function completed normally")
				}
				if !tt.expectPanic && r != nil {
					t.Errorf("Unexpected panic: %v", r)
				}
			}()

			result := UnixTaiNA(tt.sec, tt.nsec, tt.asec)
			if !tt.expectPanic {
				checkConstructorResult(t, result, tt.sec, tt.nsec, tt.asec)
			}
		})
	}
}

//revive:enable:cognitive-complexity

func testSubtraction(t *testing.T, tai1, tai2 TaiNA, expectDur time.Duration) {
	result := tai1.Sub(tai2)
	if result != expectDur {
		t.Errorf("Sub() = %v, want %v", result, expectDur)
	}
}

func testAddition(t *testing.T, tai TaiNA, duration time.Duration, expectSec uint64, expectNano uint32) {
	result, err := tai.Add(duration)
	if err != nil {
		t.Fatalf("Add() failed: %v", err)
	}
	if result.sec != expectSec || result.nano != expectNano {
		t.Errorf("Add() = sec:%d nano:%d, want sec:%d nano:%d",
			result.sec, result.nano, expectSec, expectNano)
	}
}

func TestTaiNAArithmeticEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		tai1       TaiNA
		tai2       TaiNA
		duration   time.Duration
		operation  string
		expectDur  time.Duration
		expectSec  uint64
		expectNano uint32
	}{
		{
			name:      "Subtraction at nanosecond boundary",
			tai1:      TaiNA{sec: 1000, nano: 999999999, atto: 999999999},
			tai2:      TaiNA{sec: 1000, nano: 0, atto: 0},
			operation: "Sub",
			expectDur: time.Duration(999999999) * time.Nanosecond,
		},
		{
			name:       "Addition crossing second boundary",
			tai1:       TaiNA{sec: 1000, nano: 0, atto: 0},
			operation:  "Add",
			duration:   time.Second + time.Nanosecond,
			expectSec:  1001,
			expectNano: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case "Sub":
				testSubtraction(t, tt.tai1, tt.tai2, tt.expectDur)
			case "Add":
				testAddition(t, tt.tai1, tt.duration, tt.expectSec, tt.expectNano)
			}
		})
	}
}
