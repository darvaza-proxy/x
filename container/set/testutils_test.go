package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

// testItem is the element type shared across the set tests.
type testItem struct {
	ID    int
	Name  string
	Value string
}

// testConfig returns a valid Config keyed by ID with a modulo-10 hash, so
// IDs sharing a last digit collide into the same bucket.
func testConfig() set.Config[int, int, testItem] {
	return set.Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}
}

func makeTestItems() []testItem {
	return core.S(
		testItem{ID: 1, Name: "one", Value: "first"},
		testItem{ID: 2, Name: "two", Value: "second"},
	)
}

func makeTriple() []testItem {
	return core.S(
		testItem{ID: 1, Name: "one", Value: "first"},
		testItem{ID: 2, Name: "two", Value: "second"},
		testItem{ID: 3, Name: "three", Value: "third"},
	)
}

func makeHashCollisionItems() []testItem {
	// IDs 1, 11, 21 and 31 all hash to bucket 1.
	return core.S(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 11, Name: "eleven"},
		testItem{ID: 21, Name: "twenty-one"},
		testItem{ID: 31, Name: "thirty-one"},
	)
}

// assertGetAll verifies every item can be fetched back by key.
func assertGetAll(t *testing.T, s *set.Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertNoError(t, err, "get %d", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "id %d", item.ID)
	}
}

// assertHasIDs verifies the set holds exactly the given IDs, checking the
// count (so duplicates are caught) and membership of each.
func assertHasIDs(t *testing.T, s *set.Set[int, int, testItem], ids ...int) {
	t.Helper()
	values := s.Values()
	core.AssertEqual(t, len(ids), len(values), "count")

	got := make([]int, 0, len(values))
	for _, v := range values {
		got = append(got, v.ID)
	}
	for _, id := range ids {
		core.AssertTrue(t, core.SliceContains(got, id), "contains %d", id)
	}
}
