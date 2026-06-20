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

var _ core.TestCase = lenTestCase{}

// lenTestCase verifies Set.Len across receiver states and bucket layouts.
type lenTestCase struct {
	set  *set.Set[int, int, testItem]
	name string
	want int
}

func newLenTestCase(name string, s *set.Set[int, int, testItem], want int) lenTestCase {
	return lenTestCase{
		name: name,
		set:  s,
		want: want,
	}
}

func (tc lenTestCase) Name() string {
	return tc.name
}

func (tc lenTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.set.Len(), "len")
}

func lenTestCases() []lenTestCase {
	cfg := testConfig()
	return []lenTestCase{
		newLenTestCase("nil receiver", nil, 0),
		newLenTestCase("uninitialised", &set.Set[int, int, testItem]{}, 0),
		newLenTestCase("empty", cfg.Must(), 0),
		newLenTestCase("populated", cfg.Must(makeTestItems()...), 2),
		// IDs 1, 11, 21 and 31 all hash to bucket 1, so a single bucket
		// holds every entry; Len must sum the list, not count buckets.
		newLenTestCase("hash collisions", cfg.Must(makeHashCollisionItems()...), 4),
	}
}

func TestLen(t *testing.T) {
	core.RunTestCases(t, lenTestCases())
}
