package tai

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNowTaiNA(t *testing.T) {
	now := NowTaiNA()
	if now.IsZero() {
		t.Error("NowTaiNA() returned zero time")
	}
}

func TestTaiNAFromString(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"@40000000000000000000000000000000", false},
		{"@400000004B05F9FE0000000000000000", false},
		{"@400000004B05F9FE12345678000F4240", false}, // With attoseconds = 1000000 (1 million)
		{"invalid", true},
		{"40000000000000000000000000000000", true}, // missing @
		{"@invalid", true},
		{"@400000000000000000000000", true},          // too short (TAI64N format)
		{"@40000000000000000000000000000000X", true}, // too long
		{"@400000004B05F9FE12345678FFFFFFFF", true},  // attoseconds > 999999999
	}

	for _, tt := range tests {
		_, err := TaiNAfromString(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("TaiNAfromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func testNanosecondComparison(t *testing.T) {
	t1 := TaiNAfromGoTime(time.Unix(1000, 500000000))
	t2 := TaiNAfromGoTime(time.Unix(1000, 600000000))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}
	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}
	if t1.Equal(t2) {
		t.Error("t1 should not equal t2")
	}

	t3 := TaiNAfromGoTime(time.Unix(1000, 500000000))
	if !t1.Equal(t3) {
		t.Error("t1 should equal t3")
	}
}

func testAttosecondComparison(t *testing.T) {
	t4 := UnixTaiNA(1000, 500000000, 123456789)
	t5 := UnixTaiNA(1000, 500000000, 123456790)

	if !t4.Before(t5) {
		t.Error("t4 should be before t5 (attosecond precision)")
	}
}

func testCompareMethod(t *testing.T) {
	t1 := TaiNAfromGoTime(time.Unix(1000, 500000000))
	t2 := TaiNAfromGoTime(time.Unix(1000, 600000000))
	t3 := TaiNAfromGoTime(time.Unix(1000, 500000000))
	t4 := UnixTaiNA(1000, 500000000, 123456789)
	t5 := UnixTaiNA(1000, 500000000, 123456790)

	if t1.Compare(t2) != -1 {
		t.Error("t1.Compare(t2) should return -1")
	}
	if t2.Compare(t1) != 1 {
		t.Error("t2.Compare(t1) should return 1")
	}
	if t1.Compare(t3) != 0 {
		t.Error("t1.Compare(t3) should return 0")
	}
	if t4.Compare(t5) != -1 {
		t.Error("t4.Compare(t5) should return -1 (attosecond precision)")
	}
}

func TestTaiNAComparison(t *testing.T) {
	testNanosecondComparison(t)
	testAttosecondComparison(t)
	testCompareMethod(t)
}

func TestTaiNAArithmetic(t *testing.T) {
	base := TaiNAfromGoTime(time.Unix(1000, 500000000))
	dur := 60*time.Second + 250*time.Millisecond

	later, err := base.Add(dur)
	if err != nil {
		t.Fatalf("Add error: %v", err)
	}
	diff := later.Sub(base)

	if diff != dur {
		t.Errorf("Add/Sub mismatch: got %v, want %v", diff, dur)
	}
}

func TestTaiNAAttosecondArithmetic(t *testing.T) {
	base := UnixTaiNA(1000, 500000000, 123456789)

	// Test adding attoseconds
	result, err := base.AddAttoseconds(876543210)
	if err != nil {
		t.Fatalf("AddAttoseconds error: %v", err)
	}
	expected := UnixTaiNA(1000, 500000000, 999999999)
	if !result.Equal(expected) {
		t.Errorf("AddAttoseconds failed: got %v, want %v", result, expected)
	}

	// Test attosecond overflow to nanoseconds
	result2, err := base.AddAttoseconds(1000000000) // 1 billion attoseconds = 1 nanosecond
	if err != nil {
		t.Fatalf("AddAttoseconds overflow error: %v", err)
	}
	expected2 := UnixTaiNA(1000, 500000001, 123456789)
	if !result2.Equal(expected2) {
		t.Errorf("AddAttoseconds overflow failed: got %v, want %v", result2, expected2)
	}

	// Test negative attoseconds
	result3, err := base.AddAttoseconds(-123456789)
	if err != nil {
		t.Fatalf("AddAttoseconds negative error: %v", err)
	}
	expected3 := UnixTaiNA(1000, 500000000, 0)
	if !result3.Equal(expected3) {
		t.Errorf("AddAttoseconds negative failed: got %v, want %v", result3, expected3)
	}

	// Test SubAttoseconds
	t1 := UnixTaiNA(1000, 500000000, 999999999)
	t2 := UnixTaiNA(1000, 500000000, 123456789)
	diff := t1.SubAttoseconds(t2)
	expectedDiff := int64(876543210)
	if diff != expectedDiff {
		t.Errorf("SubAttoseconds failed: got %d, want %d", diff, expectedDiff)
	}
}

func TestTaiNAAttosecondAccessors(t *testing.T) {
	tm := UnixTaiNA(1000, 123456789, 987654321)

	if tm.Nanosecond() != 123456789 {
		t.Errorf("Nanosecond() = %d, want 123456789", tm.Nanosecond())
	}

	if tm.Attosecond() != 987654321 {
		t.Errorf("Attosecond() = %d, want 987654321", tm.Attosecond())
	}

	// Test UnixAttosecondSplit calculation
	tm2 := UnixTaiNA(1, 123456789, 987654321)
	result := tm2.UnixAttosecondSplit()
	expectedNanos := tm2.UnixNano()
	expectedAttos := uint32(987654321)
	if result.UnixNanoseconds != expectedNanos {
		t.Errorf("UnixAttosecondSplit().UnixNanoseconds = %d, want %d", result.UnixNanoseconds, expectedNanos)
	}
	if result.Attoseconds != expectedAttos {
		t.Errorf("UnixAttosecondSplit().Attoseconds = %d, want %d", result.Attoseconds, expectedAttos)
	}
}

func TestTaiNATruncateRound(t *testing.T) {
	tm := UnixTaiNA(1000, 567890123, 456789012)

	// Test truncate to second
	truncated := tm.Truncate(time.Second)
	expected := UnixTaiNA(1000, 0, 0)
	if !truncated.Equal(expected) {
		t.Errorf("Truncate(Second) = %v, want %v", truncated, expected)
	}

	// Test round to second
	rounded := tm.Round(time.Second)
	expectedRound := UnixTaiNA(1001, 0, 0)
	if !rounded.Equal(expectedRound) {
		t.Errorf("Round(Second) = %v, want %v", rounded, expectedRound)
	}

	// Test truncate to nanosecond (should clear attoseconds)
	truncatedNano := tm.Truncate(time.Nanosecond)
	expectedNano := UnixTaiNA(1000, 567890123, 0)
	if !truncatedNano.Equal(expectedNano) {
		t.Errorf("Truncate(Nanosecond) = %v, want %v", truncatedNano, expectedNano)
	}
}

func TestTaiNAMarshalJSON(t *testing.T) {
	tm := UnixTaiNA(1000, 123456789, 987654321)
	data, err := json.Marshal(tm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var tm2 TaiNA
	err = json.Unmarshal(data, &tm2)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("JSON round trip failed: %v != %v", tm, tm2)
	}
}

func TestTaiNAMarshalBinary(t *testing.T) {
	tm := UnixTaiNA(1000, 123456789, 987654321)
	data := tm.MarshalBinary()

	if len(data) != TAINALength {
		t.Errorf("MarshalBinary length = %d, want %d", len(data), TAINALength)
	}

	var tm2 TaiNA
	err := tm2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("Binary round trip failed: %v != %v", tm, tm2)
	}
}

func TestTaiNAMarshalText(t *testing.T) {
	tm := UnixTaiNA(1000, 123456789, 987654321)
	data, err := tm.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}

	var tm2 TaiNA
	err = tm2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("Text round trip failed: %v != %v", tm, tm2)
	}
}

func TestTaiNAStringFormat(t *testing.T) {
	// Test known string format
	tm := UnixTaiNA(0, 0, 123456789)
	str := tm.String()

	// Should be 33 characters: @ + 32 hex digits
	if len(str) != 33 {
		t.Errorf("String length = %d, want 33", len(str))
	}

	if str[0] != '@' {
		t.Errorf("String should start with @, got %c", str[0])
	}

	// Verify round trip
	parsed, err := TaiNAfromString(str)
	if err != nil {
		t.Fatalf("TaiNAfromString error: %v", err)
	}

	if !tm.Equal(parsed) {
		t.Errorf("String round trip failed: %v != %v", tm, parsed)
	}
}

func TestTaiNAUnixMethods(t *testing.T) {
	goTime := time.Unix(1234567890, 123456789)
	tm := TaiNAfromGoTime(goTime)
	var err error
	tm, err = tm.AddAttoseconds(987654321) // Add attoseconds
	if err != nil {
		t.Fatalf("AddAttoseconds error: %v", err)
	}

	// The Unix methods should return the TAI time, not UTC time
	// TAI is ahead of UTC by the leap second offset
	expectedUnix := goTime.Unix() + int64(lsoffset(goTime))

	if tm.Unix() != expectedUnix {
		t.Errorf("Unix() = %d, want %d", tm.Unix(), expectedUnix)
	}

	expectedUnixNano := expectedUnix*1e9 + int64(tm.Nanosecond())
	if tm.UnixNano() != expectedUnixNano {
		t.Errorf("UnixNano() = %d, want %d", tm.UnixNano(), expectedUnixNano)
	}

	result := tm.UnixAttosecondSplit()
	if result.UnixNanoseconds != expectedUnixNano {
		t.Errorf("UnixAttosecondSplit().UnixNanoseconds = %d, want %d", result.UnixNanoseconds, expectedUnixNano)
	}
	if result.Attoseconds != tm.atto {
		t.Errorf("UnixAttosecondSplit().Attoseconds = %d, want %d", result.Attoseconds, tm.atto)
	}
}

func TestTaiNAConstructors(t *testing.T) {
	// Test UnixTaiNA constructor
	taina := UnixTaiNA(1234567890, 123456789, 987654321)
	expected := TaiNAfromGoTime(time.Unix(1234567890, 123456789))
	var err error
	expected, err = expected.AddAttoseconds(987654321)
	if err != nil {
		t.Fatalf("AddAttoseconds error: %v", err)
	}
	if !taina.Equal(expected) {
		t.Error("UnixTaiNA() constructor failed")
	}

	// Test DateTaiNA constructor
	taina2 := DateTaiNA(DateConfig{
		Year:  2023,
		Month: time.May,
		Day:   15,
		Hour:  10,
		Min:   30,
		Sec:   45,
		Nsec:  123456789,
		Loc:   time.UTC,
	}, 987654321)
	expected2 := TaiNAfromGoTime(time.Date(2023, time.May, 15, 10, 30, 45, 123456789, time.UTC))
	expected2, err = expected2.AddAttoseconds(987654321)
	if err != nil {
		t.Fatalf("AddAttoseconds error: %v", err)
	}
	if !taina2.Equal(expected2) {
		t.Error("DateTaiNA() constructor failed")
	}
}

func TestTaiNAConversions(t *testing.T) {
	// Test TaiNA to TaiN conversion
	taina := UnixTaiNA(1000, 123456789, 987654321)
	tain := taina.TaiN()
	expectedTaiN := TaiNfromGoTime(time.Unix(1000, 123456789))
	if !tain.Equal(expectedTaiN) {
		t.Error("TaiNA.TaiN() conversion failed")
	}

	// Test TaiNA to Tai conversion
	tai := taina.Tai()
	expectedTai := TaifromGoTime(time.Unix(1000, 123456789))
	if !tai.Equal(expectedTai) {
		t.Error("TaiNA.Tai() conversion failed")
	}

	// Test TaiN to TaiNA conversion
	tain2 := TaiNfromGoTime(time.Unix(1000, 123456789))
	taina2 := TaiNAfromTaiN(tain2)
	expected := UnixTaiNA(1000, 123456789, 0) // Attoseconds should be 0
	if !taina2.Equal(expected) {
		t.Error("TaiNAfromTaiN() conversion failed")
	}
}

func TestTaiNAErrorConditions(t *testing.T) {
	// Test invalid binary data length
	var tm TaiNA
	err := tm.UnmarshalBinary([]byte{1, 2, 3}) // Too short
	if err == nil {
		t.Error("UnmarshalBinary should fail with invalid length")
	}

	// Test invalid attosecond range in binary data
	data := make([]byte, TAINALength)
	// Set attosecond field to invalid value (> 999999999)
	data[12] = 0xFF
	data[13] = 0xFF
	data[14] = 0xFF
	data[15] = 0xFF
	err = tm.UnmarshalBinary(data)
	if err == nil {
		t.Error("UnmarshalBinary should fail with invalid attosecond range")
	}

	// Test panic conditions in constructors
	defer func() {
		if r := recover(); r == nil {
			t.Error("UnixTaiNA should panic with invalid attoseconds")
		}
	}()
	UnixTaiNA(0, 0, 1000000000) // > 999999999
}

func TestTaiNAParsing(t *testing.T) {
	// Test ParseTaiNA function
	taina, err := ParseTaiNA("tai64na", "@400000004B05F9FE12345678000F4240")
	if err != nil {
		t.Errorf("ParseTaiNA error: %v", err)
	}
	if taina.IsZero() {
		t.Error("ParseTaiNA returned zero value")
	}

	// Test unsupported layout
	_, err = ParseTaiNA("invalid", "@400000004B05F9FE12345678000F4240")
	if err == nil {
		t.Error("ParseTaiNA should fail with unsupported layout")
	}
}

func TestTaiNAEdgeCases(t *testing.T) {
	// Test zero time
	zero := TaiNA{}
	if !zero.IsZero() {
		t.Error("Zero TaiNA should report IsZero() = true")
	}

	// Test maximum attosecond value
	maxAtto := UnixTaiNA(0, 0, 999999999)
	if maxAtto.Attosecond() != 999999999 {
		t.Error("Maximum attosecond value not preserved")
	}

	// Test attosecond underflow handling
	tm := UnixTaiNA(1000, 500000000, 100)
	result, err := tm.AddAttoseconds(-200)
	if err != nil {
		t.Fatalf("AddAttoseconds underflow error: %v", err)
	}
	// Should borrow from nanoseconds: 999999900 attoseconds, 499999999 nanoseconds
	if result.Attosecond() != 999999900 || result.Nanosecond() != 499999999 {
		t.Errorf("Attosecond underflow handling failed: atto=%d nano=%d",
			result.Attosecond(), result.Nanosecond())
	}
}
