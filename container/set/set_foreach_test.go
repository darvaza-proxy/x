package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func testValuesNilReceiver(t *testing.T) {
	var s *set.Set[int, int, testItem]
	core.AssertNil(t, s.Values(), "values on nil receiver")
}

func testValuesUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	core.AssertNil(t, s.Values(), "values on uninitialised")
}

func TestValuesEdge(t *testing.T) {
	t.Run("nil receiver", testValuesNilReceiver)
	t.Run("uninitialised", testValuesUninitialised)
}

func testForEachNilReceiver(t *testing.T) {
	var s *set.Set[int, int, testItem]
	count := 0
	s.ForEach(func(testItem) bool { count++; return true })
	core.AssertEqual(t, 0, count, "no iteration on nil receiver")
}

func testForEachNilFunc(t *testing.T) {
	s := testConfig().Must(testItem{ID: 1, Name: "one"})
	core.AssertNoPanic(t, func() { s.ForEach(nil) }, "nil function is a no-op")
}

func testForEachUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	count := 0
	s.ForEach(func(testItem) bool { count++; return true })
	core.AssertEqual(t, 0, count, "no iteration on uninitialised")
}

func TestForEachEdge(t *testing.T) {
	t.Run("nil receiver", testForEachNilReceiver)
	t.Run("nil function", testForEachNilFunc)
	t.Run("uninitialised", testForEachUninitialised)
}
