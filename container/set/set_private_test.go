package set

import (
	"testing"

	"darvaza.org/core"
)

// Shared test types for internal package tests

type testItem struct {
	Name  string
	Value string
	ID    int
}

func testConfig() Config[int, int, testItem] {
	return Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}
}

// Tests that require private field access

func runTestPopErrorCases(t *testing.T) {
	t.Helper()
	// Test Pop with checkWithKey error
	cfg := testConfig()
	_, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	// Make set uninitialised by force.
	s2 := &Set[int, int, testItem]{}
	s2.cfg = cfg // Private field access
	// Try to Pop from uninitialised set.
	_, err = s2.Pop(1)
	core.AssertError(t, err, "Pop from uninitialised")
	core.AssertErrorIs(t, err, core.ErrNotImplemented, "error type")
}

func TestPopErrorCases(t *testing.T) {
	t.Run("error cases", runTestPopErrorCases)
}

// Tests that access private methods

func runTestInitNilReceiver(t *testing.T) {
	t.Helper()
	var s *Set[int, int, testItem]
	cfg := testConfig()
	err := s.init(cfg)
	core.AssertError(t, err, "init on nil receiver")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error type")
}

func runTestInitAlreadyInitialized(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	// Try to init again
	err = s.init(cfg)
	core.AssertError(t, err, "init on already initialised set")
	core.AssertContains(t, err.Error(), "already initialised", "error message")
}

func TestInit(t *testing.T) {
	t.Run("nil receiver", runTestInitNilReceiver)
	t.Run("already initialized", runTestInitAlreadyInitialized)
}

func runTestCheckWithValueNotReady(t *testing.T) {
	t.Helper()
	s := &Set[int, int, testItem]{}
	_, _, err := s.checkWithValue(testItem{ID: 1})
	core.AssertError(t, err, "checkWithValue on uninitialised set")
	core.AssertErrorIs(t, err, core.ErrNotImplemented, "error type")
}

func runTestCheckWithKeyNotReady(t *testing.T) {
	t.Helper()
	s := &Set[int, int, testItem]{}
	_, err := s.checkWithKey(1)
	core.AssertError(t, err, "checkWithKey on uninitialised set")
	core.AssertErrorIs(t, err, core.ErrNotImplemented, "error type")
}

func TestCheckMethods(t *testing.T) {
	t.Run("checkWithValue not ready", runTestCheckWithValueNotReady)
	t.Run("checkWithKey not ready", runTestCheckWithKeyNotReady)
}
