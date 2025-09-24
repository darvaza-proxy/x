package tai

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNowTaiN(t *testing.T) {
	now := NowTaiN()
	if now.IsZero() {
		t.Error("NowTaiN() returned zero time")
	}
}

func TestTaiNFromString(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"@400000000000000000000000", false},
		{"@400000004B05F9FE00000000", false},
		{"invalid", true},
		{"400000000000000000000000", true}, // missing @
		{"@invalid", true},
		{"@4000000000000000", true}, // too short
	}

	for _, tt := range tests {
		_, err := TaiNfromString(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("TaiNfromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestTaiNComparison(t *testing.T) {
	t1 := TaiNfromGoTime(time.Unix(1000, 500000000))
	t2 := TaiNfromGoTime(time.Unix(1000, 600000000))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}
	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}
	if t1.Equal(t2) {
		t.Error("t1 should not equal t2")
	}

	t3 := TaiNfromGoTime(time.Unix(1000, 500000000))
	if !t1.Equal(t3) {
		t.Error("t1 should equal t3")
	}

	if t1.Compare(t2) != -1 {
		t.Error("t1.Compare(t2) should return -1")
	}
	if t2.Compare(t1) != 1 {
		t.Error("t2.Compare(t1) should return 1")
	}
	if t1.Compare(t3) != 0 {
		t.Error("t1.Compare(t3) should return 0")
	}
}

func TestTaiNArithmetic(t *testing.T) {
	base := TaiNfromGoTime(time.Unix(1000, 500000000))
	dur := 60*time.Second + 250*time.Millisecond

	later := base.Add(dur)
	diff := later.Sub(base)

	if diff != dur {
		t.Errorf("Add/Sub mismatch: got %v, want %v", diff, dur)
	}
}

func TestTaiNNanoseconds(t *testing.T) {
	tm := TaiNfromGoTime(time.Unix(1000, 123456789))
	if tm.Nanosecond() != 123456789 {
		t.Errorf("Nanosecond() = %d, want 123456789", tm.Nanosecond())
	}
}

func TestTaiNTruncateRound(t *testing.T) {
	tm := TaiNfromGoTime(time.Unix(1000, 567890123))

	// Test truncate to second
	truncated := tm.Truncate(time.Second)
	expected := TaiNfromGoTime(time.Unix(1000, 0))
	if !truncated.Equal(expected) {
		t.Errorf("Truncate(Second) = %v, want %v", truncated, expected)
	}

	// Test round to second
	rounded := tm.Round(time.Second)
	expectedRound := TaiNfromGoTime(time.Unix(1001, 0))
	if !rounded.Equal(expectedRound) {
		t.Errorf("Round(Second) = %v, want %v", rounded, expectedRound)
	}
}

func TestTaiNMarshalJSON(t *testing.T) {
	tm := TaiNfromGoTime(time.Unix(1000, 123456789))
	data, err := json.Marshal(tm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var tm2 TaiN
	err = json.Unmarshal(data, &tm2)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("JSON round trip failed: %v != %v", tm, tm2)
	}
}

func TestTaiNMarshalBinary(t *testing.T) {
	tm := TaiNfromGoTime(time.Unix(1000, 123456789))
	data := tm.MarshalBinary()

	var tm2 TaiN
	err := tm2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("Binary round trip failed: %v != %v", tm, tm2)
	}
}

func TestUnixMethods(t *testing.T) {
	goTai := time.Unix(1234567890, 123456789)
	tm := TaiNfromGoTime(goTai)

	// The Unix methods should return the TAI time, not UTC time
	// TAI is ahead of UTC by the leap second offset
	expectedUnix := goTai.Unix() + int64(lsoffset(goTai))

	if tm.Unix() != expectedUnix {
		t.Errorf("Unix() = %d, want %d", tm.Unix(), expectedUnix)
	}

	expectedUnixNano := expectedUnix*1e9 + int64(tm.Nanosecond())
	if tm.UnixNano() != expectedUnixNano {
		t.Errorf("UnixNano() = %d, want %d", tm.UnixNano(), expectedUnixNano)
	}
}

func TestPackageConstructors(t *testing.T) {
	// Test Unix constructor
	tain := Unix(1234567890, 123456789)
	expected := TaiNfromGoTime(time.Unix(1234567890, 123456789))
	if !tain.Equal(expected) {
		t.Error("Unix() constructor failed")
	}

	// Test Date constructor
	tain2 := Date(DateConfig{
		Year:  2023,
		Month: time.May,
		Day:   15,
		Hour:  10,
		Min:   30,
		Sec:   45,
		Nsec:  123456789,
		Loc:   time.UTC,
	})
	expected2 := TaiNfromGoTime(time.Date(2023, time.May, 15, 10, 30, 45, 123456789, time.UTC))
	if !tain2.Equal(expected2) {
		t.Error("Date() constructor failed")
	}
}
