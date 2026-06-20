package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func testCopyIntoUninitialisedDst(t *testing.T) {
	cfg := testConfig()
	src := cfg.Must(testItem{ID: 1, Name: "one"}, testItem{ID: 2, Name: "two"})
	dst := &set.Set[int, int, testItem]{}

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2)

	// the copy must be independent: mutating src must not touch dst.
	_, _ = src.Push(testItem{ID: 3, Name: "three"})
	core.AssertFalse(t, dst.Contains(3), "dst independent of src")
}

func testCopyIntoCompatibleDst(t *testing.T) {
	// src and dst share the same Config, so Copy takes the append path.
	cfg := testConfig()
	src := cfg.Must(
		testItem{ID: 1, Name: "one"},     // hash 1, also present in dst
		testItem{ID: 11, Name: "eleven"}, // hash 1, new entry in existing bucket
		testItem{ID: 2, Name: "two"},     // hash 2, new bucket
	)
	dst := cfg.Must(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 3, Name: "three"},
	)

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	// union with ID 1 deduplicated, not doubled.
	assertHasIDs(t, dst, 1, 2, 3, 11)
}

func testCopyIntoCompatibleDstFiltered(t *testing.T) {
	cfg := testConfig()
	src := cfg.Must(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 2, Name: "two"},
		testItem{ID: 3, Name: "three"},
	)
	dst := cfg.Must(testItem{ID: 4, Name: "four"})

	got := src.Copy(dst, func(v testItem) bool { return v.ID != 2 })

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 3, 4)
}

func testCopyIntoEmptyCompatibleDst(t *testing.T) {
	// dst is initialised but empty, exercising the bucket preallocation.
	cfg := testConfig()
	src := cfg.Must(testItem{ID: 1, Name: "one"}, testItem{ID: 2, Name: "two"})
	dst := cfg.Must()

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2)
}

func testCopyIntoIncompatibleDst(t *testing.T) {
	// distinct Config instances are not Equal, so Copy falls back to Push.
	src := testConfig().Must(testItem{ID: 1, Name: "one"}, testItem{ID: 2, Name: "two"})
	dst := testConfig().Must(testItem{ID: 3, Name: "three"})

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2, 3)
}

func testCopyIntoIncompatibleDstFiltered(t *testing.T) {
	src := testConfig().Must(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 2, Name: "two"},
		testItem{ID: 3, Name: "three"},
	)
	dst := testConfig().Must()

	got := src.Copy(dst, func(v testItem) bool { return v.ID%2 == 0 })

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 2)
}

func TestCopyIntoDestination(t *testing.T) {
	t.Run("uninitialised dst", testCopyIntoUninitialisedDst)
	t.Run("compatible dst", testCopyIntoCompatibleDst)
	t.Run("compatible dst filtered", testCopyIntoCompatibleDstFiltered)
	t.Run("empty compatible dst", testCopyIntoEmptyCompatibleDst)
	t.Run("incompatible dst", testCopyIntoIncompatibleDst)
	t.Run("incompatible dst filtered", testCopyIntoIncompatibleDstFiltered)
}

func testCopySameSrcDst(t *testing.T) {
	cfg := testConfig()
	s := cfg.Must(testItem{ID: 1, Name: "one"}, testItem{ID: 2, Name: "two"})

	got := s.Copy(s, nil)

	core.AssertSame(t, s, got, "source returned unchanged")
	assertHasIDs(t, s, 1, 2)
}

func testCopyNilSource(t *testing.T) {
	var src *set.Set[int, int, testItem]
	dst := testConfig().Must(testItem{ID: 1, Name: "one"})

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned unchanged")
	assertHasIDs(t, dst, 1)
}

func testCopyUninitialisedSource(t *testing.T) {
	src := &set.Set[int, int, testItem]{}
	dst := testConfig().Must(testItem{ID: 1, Name: "one"})

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned unchanged")
	assertHasIDs(t, dst, 1)
}

func TestCopyNoOp(t *testing.T) {
	t.Run("source equals destination", testCopySameSrcDst)
	t.Run("nil source", testCopyNilSource)
	t.Run("uninitialised source", testCopyUninitialisedSource)
}

func testCloneNilReceiver(t *testing.T) {
	var s *set.Set[int, int, testItem]
	core.AssertNil(t, s.Clone(), "clone of nil receiver")
}

func testCloneUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	core.AssertNil(t, s.Clone(), "clone of uninitialised set")
}

func TestCloneInvalid(t *testing.T) {
	t.Run("nil receiver", testCloneNilReceiver)
	t.Run("uninitialised", testCloneUninitialised)
}
