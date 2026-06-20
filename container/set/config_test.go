package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func testConfigEqualDifferentItemKey(t *testing.T) {
	hashFn := func(k int) (int, error) { return k, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }
	cfg1 := set.Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) { return v.ID, nil }, Hash: hashFn, ItemMatch: matchFn,
	}
	cfg2 := set.Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) { return v.ID, nil }, Hash: hashFn, ItemMatch: matchFn,
	}

	core.AssertFalse(t, cfg1.Equal(cfg2), "differing ItemKey")
}

func testConfigEqualDifferentItemMatch(t *testing.T) {
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	hashFn := func(k int) (int, error) { return k, nil }
	cfg1 := set.Config[int, int, testItem]{
		ItemKey: keyFn, Hash: hashFn, ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}
	cfg2 := set.Config[int, int, testItem]{
		ItemKey: keyFn, Hash: hashFn, ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}

	core.AssertFalse(t, cfg1.Equal(cfg2), "differing ItemMatch")
}

func testConfigEqualNilCallbacks(t *testing.T) {
	full := testConfig()
	var empty set.Config[int, int, testItem]

	// two configs with nil callbacks compare equal (both nil).
	core.AssertTrue(t, empty.Equal(empty), "two empty configs")
	// a nil callback never matches a non-nil one.
	core.AssertFalse(t, empty.Equal(full), "empty versus full")
	core.AssertFalse(t, full.Equal(empty), "full versus empty")
}

func TestConfigEqualPartial(t *testing.T) {
	t.Run("different ItemKey", testConfigEqualDifferentItemKey)
	t.Run("different ItemMatch", testConfigEqualDifferentItemMatch)
	t.Run("nil callbacks", testConfigEqualNilCallbacks)
}
