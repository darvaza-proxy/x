package set_test

import (
	"slices"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

// Names spelling the IDs the shared item fixtures use.
const (
	nameOne       = "one"
	nameTwo       = "two"
	nameThree     = "three"
	nameFour      = "four"
	nameEleven    = "eleven"
	nameTwentyOne = "twenty-one"
	nameThirtyOne = "thirty-one"
)

// testItem is the element type shared across the set tests.
type testItem struct {
	Name  string
	Value string
	ID    int
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

// testAltConfig returns a valid Config for the same item type with a
// modulo-7 hash. The differing hash is what keeps it from ever being
// Equal to testConfig: identical callback bodies may share a single
// function, so only a genuine difference in behaviour is decisive.
func testAltConfig() set.Config[int, int, testItem] {
	return set.Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		Hash:      func(k int) (int, error) { return k % 7, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}
}

func makeTestItems() []testItem {
	return core.S(
		testItem{ID: 1, Name: nameOne, Value: "first"},
		testItem{ID: 2, Name: nameTwo, Value: "second"},
	)
}

func makeTriple() []testItem {
	return core.S(
		testItem{ID: 1, Name: nameOne, Value: "first"},
		testItem{ID: 2, Name: nameTwo, Value: "second"},
		testItem{ID: 3, Name: nameThree, Value: "third"},
	)
}

func makeHashCollisionItems() []testItem {
	// IDs 1, 11, 21 and 31 all hash to bucket 1.
	return core.S(
		testItem{ID: 1, Name: nameOne},
		testItem{ID: 11, Name: nameEleven},
		testItem{ID: 21, Name: nameTwentyOne},
		testItem{ID: 31, Name: nameThirtyOne},
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
		core.AssertTrue(t, slices.Contains(got, id), "contains %d", id)
	}
}
