package set_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
)

// Simple config for testing int sets where value == key
func simpleIntConfig() set.Config[int, int, int] {
	return set.Config[int, int, int]{
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemKey:   func(v int) (int, error) { return v, nil },
		ItemMatch: func(k int, v int) bool { return k == v },
	}
}

// Simple config for testing string sets where value == key
func simpleStringConfig() set.Config[string, string, string] {
	return set.Config[string, string, string]{
		Hash:      func(k string) (string, error) { return k, nil },
		ItemKey:   func(v string) (string, error) { return v, nil },
		ItemMatch: func(k string, v string) bool { return k == v },
	}
}

func TestSet_BasicOperations(t *testing.T) {
	t.Run("int set operations", runTestSetIntOperations)
	t.Run("string set operations", runTestSetStringOperations)
}

func runTestSetIntOperations(t *testing.T) {
	t.Helper()
	cfg := simpleIntConfig()
	s, err := cfg.New(1, 2, 3)
	core.AssertNoError(t, err, "new set")

	// Test Contains
	core.AssertTrue(t, s.Contains(1), "contains 1")
	core.AssertFalse(t, s.Contains(4), "not contains 4")

	// Test Get
	v, err := s.Get(2)
	core.AssertNoError(t, err, "get 2")
	core.AssertEqual(t, 2, v, "get value")

	// Test Push new
	prev, err := s.Push(4)
	core.AssertNoError(t, err, "push new")
	core.AssertEqual(t, 4, prev, "push returns new value")

	// Test Push existing
	prev, err = s.Push(2)
	core.AssertErrorIs(t, err, set.ErrExist, "push existing")
	core.AssertEqual(t, 2, prev, "push returns existing value")

	// Test Pop
	v, err = s.Pop(3)
	core.AssertNoError(t, err, "pop")
	core.AssertEqual(t, 3, v, "popped value")

	// Test Pop non-existent
	_, err = s.Pop(999)
	core.AssertErrorIs(t, err, set.ErrNotExist, "pop non-existent")
}

func runTestSetStringOperations(t *testing.T) {
	t.Helper()
	cfg := simpleStringConfig()
	s, err := cfg.New("apple", "banana", "cherry")
	core.AssertNoError(t, err, "new set")

	core.AssertTrue(t, s.Contains("banana"), "contains banana")
	core.AssertFalse(t, s.Contains("date"), "not contains date")

	v, err := s.Get("apple")
	core.AssertNoError(t, err, "get apple")
	core.AssertEqual(t, "apple", v, "get value")
}

func TestSet_NilReceiver(t *testing.T) {
	var s *set.Set[int, int, int]

	// Test nil-safe methods
	core.AssertFalse(t, s.Contains(1), "nil set contains")

	_, err := s.Get(1)
	core.AssertError(t, err, "nil set get")

	_, err = s.Push(1)
	core.AssertError(t, err, "nil set push")

	_, err = s.Pop(1)
	core.AssertError(t, err, "nil set pop")

	values := s.Values()
	core.AssertEqual(t, 0, len(values), "nil set values")

	s.ForEach(func(int) bool { return true })
	// Should not panic

	clone := s.Clone()
	core.AssertNil(t, clone, "nil set clone")
}

func TestSet_EmptySet(t *testing.T) {
	cfg := simpleIntConfig()
	s, err := cfg.New()
	core.AssertNoError(t, err, "new empty set")

	core.AssertFalse(t, s.Contains(1), "empty contains")

	_, err = s.Get(1)
	core.AssertErrorIs(t, err, set.ErrNotExist, "empty get")

	_, err = s.Pop(1)
	core.AssertErrorIs(t, err, set.ErrNotExist, "empty pop")

	values := s.Values()
	core.AssertEqual(t, 0, len(values), "empty values")

	count := 0
	s.ForEach(func(int) bool {
		count++
		return true
	})
	core.AssertEqual(t, 0, count, "empty foreach")
}

func TestSet_ForEach(t *testing.T) {
	cfg := simpleIntConfig()
	s, err := cfg.New(1, 2, 3, 4, 5)
	core.AssertNoError(t, err, "new set")

	// Test full iteration
	var sum int
	s.ForEach(func(v int) bool {
		sum += v
		return true
	})
	core.AssertEqual(t, 15, sum, "foreach sum")

	// Test early termination
	var count int
	s.ForEach(func(_ int) bool {
		count++
		return count < 3
	})
	core.AssertEqual(t, 3, count, "foreach early termination")

	// Test with nil function
	s.ForEach(nil)
	// Should not panic
}

func TestSet_Clone(t *testing.T) {
	cfg := simpleIntConfig()
	s, err := cfg.New(10, 20, 30)
	core.AssertNoError(t, err, "new set")

	clone := s.Clone()
	core.AssertNotNil(t, clone, "clone not nil")
	core.AssertEqual(t, len(s.Values()), len(clone.Values()), "clone length")

	// Verify same values
	for _, v := range []int{10, 20, 30} {
		core.AssertTrue(t, clone.Contains(v), "clone contains %d", v)
	}

	// Verify independence
	_, err = s.Push(40)
	core.AssertNoError(t, err, "push to original")
	core.AssertFalse(t, clone.Contains(40), "clone independence")
}

func TestSet_ConfigValidation(t *testing.T) {
	t.Run("valid config", runTestSetConfigValidationValid)
	t.Run("missing Hash", runTestSetConfigValidationMissingHash)
	t.Run("missing ItemKey", runTestSetConfigValidationMissingItemKey)
	t.Run("missing ItemMatch", runTestSetConfigValidationMissingItemMatch)
}

func runTestSetConfigValidationValid(t *testing.T) {
	t.Helper()
	cfg := simpleIntConfig()
	err := cfg.Validate()
	core.AssertNoError(t, err, "valid config")
}

func runTestSetConfigValidationMissingHash(t *testing.T) {
	t.Helper()
	cfg := simpleIntConfig()
	cfg.Hash = nil
	err := cfg.Validate()
	core.AssertError(t, err, "missing hash")
	core.AssertContains(t, err.Error(), "Hash", "error mentions Hash")
}

func runTestSetConfigValidationMissingItemKey(t *testing.T) {
	t.Helper()
	cfg := simpleIntConfig()
	cfg.ItemKey = nil
	err := cfg.Validate()
	core.AssertError(t, err, "missing ItemKey")
	core.AssertContains(t, err.Error(), "ItemKey", "error mentions ItemKey")
}

func runTestSetConfigValidationMissingItemMatch(t *testing.T) {
	t.Helper()
	cfg := simpleIntConfig()
	cfg.ItemMatch = nil
	err := cfg.Validate()
	core.AssertError(t, err, "missing ItemMatch")
	core.AssertContains(t, err.Error(), "ItemMatch", "error mentions ItemMatch")
}

func TestSet_HashCollisions(t *testing.T) {
	// Create config where all keys hash to the same bucket
	cfg := set.Config[int, int, int]{
		Hash:      func(_ int) (int, error) { return 0, nil }, // all hash to 0
		ItemKey:   func(v int) (int, error) { return v, nil },
		ItemMatch: func(k int, v int) bool { return k == v },
	}

	s, err := cfg.New(1, 2, 3, 4, 5)
	core.AssertNoError(t, err, "new set with collisions")

	// All should be findable despite hash collisions
	for _, v := range []int{1, 2, 3, 4, 5} {
		core.AssertTrue(t, s.Contains(v), "contains %d", v)
	}

	core.AssertEqual(t, 5, len(s.Values()), "length with collisions")
}

// Person type for testing key extraction (when key != value)
type Person struct {
	ID   int
	Name string
}

func personConfig() set.Config[int, int, Person] {
	return set.Config[int, int, Person]{
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemKey:   func(p Person) (int, error) { return p.ID, nil },
		ItemMatch: func(k int, p Person) bool { return k == p.ID },
	}
}

func TestSet_KeyExtraction(t *testing.T) {
	cfg := personConfig()
	people := []Person{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	s, err := cfg.New(people...)
	core.AssertNoError(t, err, "new set")

	// Test Get by ID
	p, err := s.Get(2)
	core.AssertNoError(t, err, "get person")
	core.AssertEqual(t, "Bob", p.Name, "person name")

	// Test Contains by ID
	core.AssertTrue(t, s.Contains(1), "contains Alice")
	core.AssertFalse(t, s.Contains(4), "not contains ID 4")

	// Test Pop by ID
	p, err = s.Pop(3)
	core.AssertNoError(t, err, "pop Charlie")
	core.AssertEqual(t, "Charlie", p.Name, "popped name")
	core.AssertEqual(t, 2, len(s.Values()), "length after pop")
}
