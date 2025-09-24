package tai

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaiFromString(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"@4000000000000000", false},
		{"@400000004B05F9FE", false},
		{"invalid", true},
		{"4000000000000000", true}, // missing @
		{"@invalid", true},
	}

	for _, tt := range tests {
		_, err := TaifromString(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("TaifromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
	}
}

func TestTaiComparison(t *testing.T) {
	t1 := TaifromGoTime(time.Unix(1000, 0))
	t2 := TaifromGoTime(time.Unix(2000, 0))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}
	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}
	if t1.Equal(t2) {
		t.Error("t1 should not equal t2")
	}

	t3 := TaifromGoTime(time.Unix(1000, 0))
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

func TestTaiArithmetic(t *testing.T) {
	base := TaifromGoTime(time.Unix(1000, 0))
	dur := 60 * time.Second

	later := base.Add(dur)
	diff := later.Sub(base)

	if diff != dur {
		t.Errorf("Add/Sub mismatch: got %v, want %v", diff, dur)
	}
}

func TestTaiMarshalJSON(t *testing.T) {
	tm := TaifromGoTime(time.Unix(1000, 0))
	data, err := json.Marshal(tm)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var tm2 Tai
	err = json.Unmarshal(data, &tm2)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("JSON round trip failed: %v != %v", tm, tm2)
	}
}

func TestTaiMarshalBinary(t *testing.T) {
	tm := TaifromGoTime(time.Unix(1000, 0))
	data := tm.MarshalBinary()

	var tm2 Tai
	err := tm2.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary error: %v", err)
	}

	if !tm.Equal(tm2) {
		t.Errorf("Binary round trip failed: %v != %v", tm, tm2)
	}
}

func TestConversion(t *testing.T) {
	// Test Tai to TaiN conversion
	tai := TaifromGoTime(time.Unix(1000, 0))
	tain := tai.TaiN()
	if tain.Nanosecond() != 0 {
		t.Error("Tai.TaiN() should have zero nanoseconds")
	}

	// Test TaiN to Tai conversion
	tain2 := TaiNfromGoTime(time.Unix(1000, 123456789))
	tai2 := tain2.Tai()
	if tai2.Unix() != tain2.Unix() {
		t.Error("TaiN.Tai() conversion failed")
	}
}
