package list_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/list"
)

func TestNewWithValues(t *testing.T) {
	values := core.S(1, 2, 3, 4, 5)
	l := list.New(values...)

	core.AssertSliceEqual(t, values, l.Values(), "values")
}

var _ core.TestCase = frontBackTestCase{}

// frontBackTestCase verifies Front and Back against a list built from input.
type frontBackTestCase struct {
	name      string
	wantFront string
	wantBack  string
	input     []string
	wantOK    bool
}

func newFrontBackTestCase(name string, input []string,
	wantFront, wantBack string, wantOK bool) frontBackTestCase {
	return frontBackTestCase{
		name:      name,
		wantFront: wantFront,
		wantBack:  wantBack,
		input:     input,
		wantOK:    wantOK,
	}
}

func (tc frontBackTestCase) Name() string { return tc.name }

func (tc frontBackTestCase) Test(t *testing.T) {
	t.Helper()

	l := list.New(tc.input...)
	front, frontOK := l.Front()
	back, backOK := l.Back()

	core.AssertEqual(t, tc.wantOK, frontOK, "front present")
	core.AssertEqual(t, tc.wantOK, backOK, "back present")
	core.AssertEqual(t, tc.wantFront, front, "front")
	core.AssertEqual(t, tc.wantBack, back, "back")
}

func frontBackTestCases() []frontBackTestCase {
	return []frontBackTestCase{
		newFrontBackTestCase("empty list", core.S[string](), "", "", false),
		newFrontBackTestCase("single element", core.S("only"), "only", "only",
			true),
		newFrontBackTestCase("multiple elements",
			core.S("first", "middle", "last"), "first", "last", true),
	}
}

func TestFrontBack(t *testing.T) {
	core.RunTestCases(t, frontBackTestCases())
}

func TestPushFrontBack(t *testing.T) {
	l := list.New[int]()

	// Build: [3, 1, 2, 4]
	l.PushBack(1)
	l.PushBack(2)
	l.PushFront(3)
	l.PushBack(4)

	core.AssertSliceEqual(t, core.S(3, 1, 2, 4), l.Values(), "values")
}

var _ core.TestCase = deleteMatchFnTestCase{}

// deleteMatchFnTestCase deletes even numbers from input and checks the remainder.
type deleteMatchFnTestCase struct {
	name  string
	input []int
	want  []int
}

func newDeleteMatchFnTestCase(name string, input,
	want []int) deleteMatchFnTestCase {
	return deleteMatchFnTestCase{
		name:  name,
		input: input,
		want:  want,
	}
}

func (tc deleteMatchFnTestCase) Name() string { return tc.name }

func (tc deleteMatchFnTestCase) Test(t *testing.T) {
	t.Helper()

	l := list.New(tc.input...)
	l.DeleteMatchFn(func(v int) bool { return v%2 == 0 })
	core.AssertSliceEqual(t, tc.want, l.Values(), "values")
}

func deleteMatchFnTestCases() []deleteMatchFnTestCase {
	return []deleteMatchFnTestCase{
		newDeleteMatchFnTestCase("delete none", core.S(1, 3, 5, 7),
			core.S(1, 3, 5, 7)),
		newDeleteMatchFnTestCase("delete some", core.S(1, 2, 3, 4, 5),
			core.S(1, 3, 5)),
		newDeleteMatchFnTestCase("delete all", core.S(2, 4, 6, 8),
			core.S[int]()),
	}
}

func TestDeleteMatchFn(t *testing.T) {
	core.RunTestCases(t, deleteMatchFnTestCases())
}

var _ core.TestCase = popFirstMatchFnTestCase{}

// popFirstMatchFnTestCase pops the first even number, checking value and remainder.
type popFirstMatchFnTestCase struct {
	name      string
	input     []int
	want      []int
	wantValue int
	wantOK    bool
}

func newPopFirstMatchFnTestCase(name string, input []int, wantValue int,
	wantOK bool, want []int) popFirstMatchFnTestCase {
	return popFirstMatchFnTestCase{
		name:      name,
		input:     input,
		want:      want,
		wantValue: wantValue,
		wantOK:    wantOK,
	}
}

func (tc popFirstMatchFnTestCase) Name() string { return tc.name }

func (tc popFirstMatchFnTestCase) Test(t *testing.T) {
	t.Helper()

	l := list.New(tc.input...)
	v, ok := l.PopFirstMatchFn(func(n int) bool { return n%2 == 0 })

	core.AssertEqual(t, tc.wantOK, ok, "popped")
	core.AssertEqual(t, tc.wantValue, v, "value")
	core.AssertSliceEqual(t, tc.want, l.Values(), "remaining")
}

func popFirstMatchFnTestCases() []popFirstMatchFnTestCase {
	return []popFirstMatchFnTestCase{
		newPopFirstMatchFnTestCase("found", core.S(1, 2, 3, 4, 5), 2, true,
			core.S(1, 3, 4, 5)),
		newPopFirstMatchFnTestCase("not found", core.S(1, 3, 5), 0, false,
			core.S(1, 3, 5)),
	}
}

func TestPopFirstMatchFn(t *testing.T) {
	core.RunTestCases(t, popFirstMatchFnTestCases())
}

func TestMoveToBackFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first even number to back
	l.MoveToBackFirstMatchFn(func(n int) bool {
		return n%2 == 0
	})

	core.AssertSliceEqual(t, core.S(1, 3, 4, 5, 2), l.Values(), "values")
}

func TestMoveToFrontFirstMatchFn(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5)
	// Move first number > 3 to front
	l.MoveToFrontFirstMatchFn(func(n int) bool {
		return n > 3
	})

	core.AssertSliceEqual(t, core.S(4, 1, 2, 3, 5), l.Values(), "values")
}

var _ core.TestCase = firstMatchFnTestCase{}

// firstMatchFnTestCase finds the first string longer than five characters.
type firstMatchFnTestCase struct {
	name      string
	wantValue string
	input     []string
	wantOK    bool
}

func newFirstMatchFnTestCase(name string, input []string, wantValue string,
	wantOK bool) firstMatchFnTestCase {
	return firstMatchFnTestCase{
		name:      name,
		wantValue: wantValue,
		input:     input,
		wantOK:    wantOK,
	}
}

func (tc firstMatchFnTestCase) Name() string { return tc.name }

func (tc firstMatchFnTestCase) Test(t *testing.T) {
	t.Helper()

	l := list.New(tc.input...)
	v, ok := l.FirstMatchFn(func(s string) bool { return len(s) > 5 })

	core.AssertEqual(t, tc.wantOK, ok, "found")
	core.AssertEqual(t, tc.wantValue, v, "match")
}

func firstMatchFnTestCases() []firstMatchFnTestCase {
	return []firstMatchFnTestCase{
		newFirstMatchFnTestCase("found",
			core.S("apple", "banana", "cherry", "date"), "banana", true),
		newFirstMatchFnTestCase("not found", core.S("a", "b", "c"), "", false),
	}
}

func TestFirstMatchFn(t *testing.T) {
	core.RunTestCases(t, firstMatchFnTestCases())
}

func TestZero(t *testing.T) {
	l := list.New[int]()
	core.AssertEqual(t, 0, l.Zero(), "int zero")

	ls := list.New[string]()
	core.AssertEqual(t, "", ls.Zero(), "string zero")

	type custom struct {
		b string
		a int
	}
	lc := list.New[custom]()
	z := lc.Zero()
	core.AssertEqual(t, 0, z.a, "custom zero a")
	core.AssertEqual(t, "", z.b, "custom zero b")
}

func TestClone(t *testing.T) {
	original := list.New(1, 2, 3)
	cloned := original.Clone()

	// Verify they have same values
	core.AssertSliceEqual(t, original.Values(), cloned.Values(), "clone values")

	// Verify they are independent
	cloned.PushBack(4)
	core.AssertNotEqual(t, original.Len(), cloned.Len(), "independent length")
}

func testCopyFilterTransform(t *testing.T) {
	l := list.New(1, 2, 3, 4, 5, 6)
	// Copy only even numbers, tripled
	copied := l.Copy(func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 3, true
		}
		return 0, false
	})

	core.AssertSliceEqual(t, core.S(6, 12, 18), copied.Values(), "values")
}

func testCopyAll(t *testing.T) {
	l := list.New("a", "b", "c")
	copied := l.Copy(func(v string) (string, bool) {
		return v + v, true // double each string
	})

	core.AssertSliceEqual(t, core.S("aa", "bb", "cc"), copied.Values(), "values")
}

func TestCopy(t *testing.T) {
	t.Run("filter and transform", testCopyFilterTransform)
	t.Run("copy all", testCopyAll)
}

func TestPurge(t *testing.T) {
	// Purge is specifically for removing elements that don't match the type
	// This is more relevant when dealing with interface{} conversions
	// For a generic List[T], all elements should already be of type T
	l := list.New[int]()

	// Add some values
	l.PushBack(1)
	l.PushBack(2)
	l.PushBack(3)

	// Since all elements are already int, purge should remove nothing
	core.AssertEqual(t, 0, l.Purge(), "removed")
	core.AssertEqual(t, 3, l.Len(), "length")
}

func TestNilList(t *testing.T) {
	var l *list.List[int]

	// Test nil-safe methods
	core.AssertEqual(t, 0, l.Len(), "nil length")
	core.AssertNil(t, l.Sys(), "nil Sys")

	_, frontOK := l.Front()
	core.AssertFalse(t, frontOK, "nil Front")

	_, backOK := l.Back()
	core.AssertFalse(t, backOK, "nil Back")

	core.AssertEqual(t, 0, len(l.Values()), "nil Values")

	// These should not panic
	core.AssertNoPanic(t, func() {
		l.PushFront(1)
		l.PushBack(2)
		l.ForEach(func(int) bool { return true })
		l.DeleteMatchFn(func(int) bool { return true })
	}, "nil mutators")
}
