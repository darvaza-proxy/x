package slices_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/slices"
)

var _ core.TestCase = orderedSetTestCase{}

// orderedSetTestCase checks NewOrderedSet deduplication and ordering.
type orderedSetTestCase struct {
	name  string
	input []string
	want  []string
}

func newOrderedSetTestCase(name string, input, want []string) orderedSetTestCase {
	return orderedSetTestCase{
		input: input,
		want:  want,
		name:  name,
	}
}

func (tc orderedSetTestCase) Name() string {
	return tc.name
}

func (tc orderedSetTestCase) Test(t *testing.T) {
	t.Helper()
	got := slices.NewOrderedSet(tc.input...)
	core.AssertSliceEqual(t, tc.want, got.Export(), "values")
}

func orderedSetTestCases() []orderedSetTestCase {
	return []orderedSetTestCase{
		newOrderedSetTestCase("empty", nil, core.S[string]()),
		newOrderedSetTestCase("no duplicates",
			core.S("a", "b", "c"), core.S("a", "b", "c")),
		newOrderedSetTestCase("with duplicates",
			core.S("a", "b", "a", "c", "b", "c"), core.S("a", "b", "c")),
	}
}

func TestSet(t *testing.T) {
	core.RunTestCases(t, orderedSetTestCases())
}

var _ core.TestCase = sortedSetAddTestCase{}

// sortedSetAddTestCase checks OrderedSet.Add count and resulting order.
type sortedSetAddTestCase struct {
	name    string
	initial []int
	add     []int
	result  []int
	want    int
}

func newSortedSetAddTestCase(name string, initial, add []int, want int,
	result []int) sortedSetAddTestCase {
	return sortedSetAddTestCase{
		initial: initial,
		add:     add,
		result:  result,
		name:    name,
		want:    want,
	}
}

func (tc sortedSetAddTestCase) Name() string {
	return tc.name
}

func (tc sortedSetAddTestCase) Test(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet[int]()
	set.Add(tc.initial...)
	core.AssertEqual(t, tc.want, set.Add(tc.add...), "Add count")
	core.AssertSliceEqual(t, tc.result, set.Export(), "result")
}

func sortedSetAddTestCases() []sortedSetAddTestCase {
	return []sortedSetAddTestCase{
		newSortedSetAddTestCase("add to empty set",
			nil, core.S(1, 2, 3), 3, core.S(1, 2, 3)),
		newSortedSetAddTestCase("add with duplicates",
			core.S(1, 2), core.S(2, 3, 3, 4), 2, core.S(1, 2, 3, 4)),
		newSortedSetAddTestCase("add nothing",
			core.S(1, 2, 3), core.S[int](), 0, core.S(1, 2, 3)),
		newSortedSetAddTestCase("add to nil set",
			nil, core.S(1, 2), 2, core.S(1, 2)),
		newSortedSetAddTestCase("add unsorted values",
			core.S(2, 4), core.S(5, 1, 3), 3, core.S(1, 2, 3, 4, 5)),
	}
}

func TestSortedSet_Add(t *testing.T) {
	core.RunTestCases(t, sortedSetAddTestCases())
}

var _ core.TestCase = sortedSetRemoveTestCase{}

// sortedSetRemoveTestCase checks OrderedSet.Remove count and remaining order.
type sortedSetRemoveTestCase struct {
	name    string
	initial []int
	remove  []int
	result  []int
	want    int
}

func newSortedSetRemoveTestCase(name string, initial, remove []int, want int,
	result []int) sortedSetRemoveTestCase {
	return sortedSetRemoveTestCase{
		initial: initial,
		remove:  remove,
		result:  result,
		name:    name,
		want:    want,
	}
}

func (tc sortedSetRemoveTestCase) Name() string {
	return tc.name
}

func (tc sortedSetRemoveTestCase) Test(t *testing.T) {
	t.Helper()
	set := slices.NewOrderedSet[int]()
	set.Add(tc.initial...)
	core.AssertEqual(t, tc.want, set.Remove(tc.remove...), "Remove count")
	core.AssertSliceEqual(t, tc.result, set.Export(), "result")
}

func sortedSetRemoveTestCases() []sortedSetRemoveTestCase {
	return []sortedSetRemoveTestCase{
		newSortedSetRemoveTestCase("remove from empty",
			core.S[int](), core.S(1, 2, 3), 0, core.S[int]()),
		newSortedSetRemoveTestCase("remove non-existing",
			core.S(1, 3, 5), core.S(2, 4), 0, core.S(1, 3, 5)),
		newSortedSetRemoveTestCase("remove duplicates",
			core.S(1, 2, 3, 4, 5), core.S(2, 2, 4, 4), 2, core.S(1, 3, 5)),
		newSortedSetRemoveTestCase("remove all",
			core.S(1, 2, 3), core.S(1, 2, 3), 3, core.S[int]()),
		newSortedSetRemoveTestCase("remove partial",
			core.S(1, 2, 3, 4, 5), core.S(2, 4), 2, core.S(1, 3, 5)),
		newSortedSetRemoveTestCase("remove more than present",
			core.S(1, 2, 3), core.S(1, 2, 3, 4), 3, core.S[int]()),
	}
}

func TestSortedSet_Remove(t *testing.T) {
	core.RunTestCases(t, sortedSetRemoveTestCases())
}

var _ core.TestCase = trimNTestCase{}

// trimNTestCase checks OrderedSet.TrimN against a pre-built set fixture.
type trimNTestCase struct {
	set         *slices.CustomSet[int]
	name        string
	minCapacity int
	want        bool
}

func newTrimNTestCase(name string, s *slices.CustomSet[int], minCapacity int,
	want bool) trimNTestCase {
	return trimNTestCase{
		set:         s,
		name:        name,
		minCapacity: minCapacity,
		want:        want,
	}
}

func (tc trimNTestCase) Name() string {
	return tc.name
}

func (tc trimNTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.set.TrimN(tc.minCapacity), "TrimN")
	if tc.set != nil {
		_, c := tc.set.Cap()
		core.AssertTrue(t, c >= tc.minCapacity, "capacity")
	}
}

// grownEmptySet returns an empty set carrying spare capacity, so that
// trimming it to zero exercises the capacity == 0 branch of doTrim.
func grownEmptySet() *slices.CustomSet[int] {
	s := slices.NewOrderedSet[int]()
	s.Grow(10)
	return s
}

func trimNTestCases() []trimNTestCase {
	return []trimNTestCase{
		newTrimNTestCase("nil set", nil, 0, false),
		newTrimNTestCase("empty set", slices.NewOrderedSet[int](), 0, false),
		newTrimNTestCase("empty set with spare capacity", grownEmptySet(), 0, true),
		newTrimNTestCase("set with elements above min capacity",
			slices.NewOrderedSet(1, 2, 3, 4, 5), 3, false),
		newTrimNTestCase("set with elements below min capacity",
			slices.NewOrderedSet(1, 2), 5, true),
		newTrimNTestCase("negative min capacity",
			slices.NewOrderedSet(1, 2, 3), -1, false),
	}
}

func TestSortedSet_TrimN(t *testing.T) {
	core.RunTestCases(t, trimNTestCases())
}
