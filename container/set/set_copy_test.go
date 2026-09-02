package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func testCopyIntoUninitialisedDst(t *testing.T) {
	cfg := testConfig()
	src := cfg.Must(testItem{ID: 1, Name: nameOne}, testItem{ID: 2, Name: nameTwo})
	dst := &set.Set[int, int, testItem]{}

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2)

	// the copy must be independent: mutating src must not touch dst.
	_, _ = src.Push(testItem{ID: 3, Name: nameThree})
	core.AssertFalse(t, dst.Contains(3), "dst independent of src")
}

func testCopyIntoCompatibleDst(t *testing.T) {
	// src and dst share the same Config, so Copy takes the append path.
	cfg := testConfig()
	src := cfg.Must(
		testItem{ID: 1, Name: nameOne},     // hash 1, also present in dst
		testItem{ID: 11, Name: nameEleven}, // hash 1, new entry in existing bucket
		testItem{ID: 2, Name: nameTwo},     // hash 2, new bucket
	)
	dst := cfg.Must(
		testItem{ID: 1, Name: nameOne},
		testItem{ID: 3, Name: nameThree},
	)

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	// union with ID 1 deduplicated, not doubled.
	assertHasIDs(t, dst, 1, 2, 3, 11)
}

func testCopyIntoCompatibleDstFiltered(t *testing.T) {
	cfg := testConfig()
	src := cfg.Must(
		testItem{ID: 1, Name: nameOne},
		testItem{ID: 2, Name: nameTwo},
		testItem{ID: 3, Name: nameThree},
	)
	dst := cfg.Must(testItem{ID: 4, Name: nameFour})

	got := src.Copy(dst, func(v testItem) bool { return v.ID != 2 })

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 3, 4)
}

func testCopyIntoEmptyCompatibleDst(t *testing.T) {
	// dst is initialised but empty, exercising the bucket preallocation.
	cfg := testConfig()
	src := cfg.Must(testItem{ID: 1, Name: nameOne}, testItem{ID: 2, Name: nameTwo})
	dst := cfg.Must()

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2)
}

func testCopyIntoIncompatibleDst(t *testing.T) {
	// dst hashes differently, so the Configs are not Equal and Copy
	// falls back to Push.
	src := testConfig().Must(testItem{ID: 1, Name: nameOne}, testItem{ID: 2, Name: nameTwo})
	dst := testAltConfig().Must(testItem{ID: 3, Name: nameThree})

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned")
	assertHasIDs(t, dst, 1, 2, 3)
}

func testCopyIntoIncompatibleDstFiltered(t *testing.T) {
	src := testConfig().Must(
		testItem{ID: 1, Name: nameOne},
		testItem{ID: 2, Name: nameTwo},
		testItem{ID: 3, Name: nameThree},
	)
	dst := testAltConfig().Must()

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
	s := cfg.Must(testItem{ID: 1, Name: nameOne}, testItem{ID: 2, Name: nameTwo})

	got := s.Copy(s, nil)

	core.AssertSame(t, s, got, "source returned unchanged")
	assertHasIDs(t, s, 1, 2)
}

func testCopyNilSource(t *testing.T) {
	var src *set.Set[int, int, testItem]
	dst := testConfig().Must(testItem{ID: 1, Name: nameOne})

	got := src.Copy(dst, nil)

	core.AssertSame(t, dst, got, "dst returned unchanged")
	assertHasIDs(t, dst, 1)
}

func testCopyUninitialisedSource(t *testing.T) {
	src := &set.Set[int, int, testItem]{}
	dst := testConfig().Must(testItem{ID: 1, Name: nameOne})

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
