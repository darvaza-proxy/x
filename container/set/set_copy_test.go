package set_test

import (
	"testing"

	"darvaza.org/core"
	"darvaza.org/x/container/set"
)

// Interface validations
var _ core.TestCase = copyFilterTestCase{}
var _ core.TestCase = copyToExistingTestCase{}
var _ core.TestCase = copyCollisionTestCase{}
var _ core.TestCase = copyEmptyTestCase{}

// copyFilterTestCase implements core.TestCase for copy filter tests
type copyFilterTestCase struct {
	filter       func(testItem) bool
	name         string
	sourceItems  []testItem
	expectIDs    []int // IDs expected to be in result
	notExpectIDs []int // IDs expected NOT to be in result
}

func (tc copyFilterTestCase) Name() string {
	return tc.name
}

func (tc copyFilterTestCase) Test(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	src, err := cfg.New(tc.sourceItems...)
	core.AssertMustNoError(t, err, "New source set")

	dst := src.Copy(nil, tc.filter)
	core.AssertNotNil(t, dst, "Copy result")

	// Check expected items
	for _, id := range tc.expectIDs {
		core.AssertTrue(t, dst.Contains(id), "should contain %d", id)
	}

	// Check not expected items
	for _, id := range tc.notExpectIDs {
		core.AssertFalse(t, dst.Contains(id), "should not contain %d", id)
	}

	// Verify count if all IDs are specified
	if tc.filter == nil && len(tc.expectIDs) > 0 {
		values := dst.Values()
		core.AssertEqual(t, len(tc.expectIDs), len(values), "item count")
	}
}

// newCopyFilterTestCase creates a new copyFilterTestCase
func newCopyFilterTestCase(name string, sourceItems []testItem, filter func(testItem) bool,
	expectIDs []int, notExpectIDs []int) copyFilterTestCase {
	return copyFilterTestCase{
		name:         name,
		sourceItems:  sourceItems,
		filter:       filter,
		expectIDs:    expectIDs,
		notExpectIDs: notExpectIDs,
	}
}

func makeCopyFilterTestCases() []core.TestCase {
	baseItems := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
		{ID: 4, Name: "four", Value: "fourth"},
		{ID: 5, Name: "five", Value: "fifth"},
	}

	return []core.TestCase{
		newCopyFilterTestCase(
			"filter ID greater than 2",
			baseItems,
			func(item testItem) bool { return item.ID > 2 },
			[]int{3, 4, 5},
			[]int{1, 2},
		),
		newCopyFilterTestCase(
			"nil condition copies all",
			baseItems,
			nil,
			[]int{1, 2, 3, 4, 5},
			nil,
		),
		newCopyFilterTestCase(
			"false condition copies none",
			baseItems,
			func(_ testItem) bool { return false },
			[]int{}, // Expect empty
			nil,
		),
	}
}

// TestSet_Copy_WithCondition tests the Copy method with a filtering condition.
func TestSet_Copy_WithCondition(t *testing.T) {
	core.RunTestCases(t, makeCopyFilterTestCases())
}

// copyToExistingTestCase implements core.TestCase for copy to existing set tests
type copyToExistingTestCase struct {
	filter            func(testItem) bool
	name              string
	sourceItems       []testItem
	destItems         []testItem
	expectIDs         []int
	notExpectIDs      []int
	uninitializedDest bool
}

func (tc copyToExistingTestCase) Name() string {
	return tc.name
}

func (tc copyToExistingTestCase) Test(t *testing.T) {
	t.Helper()
	cfg := testConfig()

	src, err := cfg.New(tc.sourceItems...)
	core.AssertMustNoError(t, err, "New source set")

	var dst *set.Set[int, int, testItem]
	if tc.uninitializedDest {
		dst = &set.Set[int, int, testItem]{}
	} else {
		dst, err = cfg.New(tc.destItems...)
		core.AssertMustNoError(t, err, "New destination set")
	}

	result := src.Copy(dst, tc.filter)
	core.AssertSame(t, dst, result, "should return same destination")

	// Check expected items
	for _, id := range tc.expectIDs {
		core.AssertTrue(t, dst.Contains(id), "should contain %d", id)
	}

	// Check not expected items
	for _, id := range tc.notExpectIDs {
		core.AssertFalse(t, dst.Contains(id), "should not contain %d", id)
	}
}

// newCopyToExistingTestCase creates a new copyToExistingTestCase
//
//revive:disable-next-line:argument-limit
func newCopyToExistingTestCase(name string, sourceItems []testItem, destItems []testItem,
	filter func(testItem) bool, expectIDs []int, notExpectIDs []int,
	uninitializedDest bool) copyToExistingTestCase {
	return copyToExistingTestCase{
		name:              name,
		sourceItems:       sourceItems,
		destItems:         destItems,
		filter:            filter,
		expectIDs:         expectIDs,
		notExpectIDs:      notExpectIDs,
		uninitializedDest: uninitializedDest,
	}
}

func makeCopyToExistingTestCases() []core.TestCase {
	return []core.TestCase{
		newCopyToExistingTestCase(
			"copy to uninitialised set",
			[]testItem{
				{ID: 1, Name: "one", Value: "first"},
				{ID: 2, Name: "two", Value: "second"},
			},
			nil,         // destItems
			nil,         // filter
			[]int{1, 2}, // expectIDs
			nil,         // notExpectIDs
			true,        // uninitializedDest
		),
		newCopyToExistingTestCase(
			"copy to set with same config",
			[]testItem{
				{ID: 1, Name: "one", Value: "first"},
				{ID: 2, Name: "two", Value: "second"},
			},
			[]testItem{
				{ID: 3, Name: "three", Value: "third"},
			},
			nil,            // filter
			[]int{1, 2, 3}, // expectIDs
			nil,            // notExpectIDs
			false,          // uninitializedDest
		),
		newCopyToExistingTestCase(
			"copy with filter to existing set",
			[]testItem{
				{ID: 1, Name: "one", Value: "first"},
				{ID: 2, Name: "two", Value: "second"},
				{ID: 3, Name: "three", Value: "third"},
			},
			[]testItem{
				{ID: 10, Name: "ten", Value: "tenth"},
			},
			func(item testItem) bool {
				return item.ID%2 == 0
			},
			[]int{2, 10}, // expectIDs
			[]int{1, 3},  // notExpectIDs
			false,        // uninitializedDest
		),
	}
}

// TestSet_Copy_ToExistingSet tests copying to an existing set.
func TestSet_Copy_ToExistingSet(t *testing.T) {
	core.RunTestCases(t, makeCopyToExistingTestCases())
}

func runTestCopyWithDifferentConfig(t *testing.T) {
	t.Helper()
	cfg1 := testConfig()
	// Different hash function.
	cfg2 := set.Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		Hash:      func(k int) (int, error) { return k % 5, nil }, // Different hash.
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}

	src, err := cfg1.New(
		testItem{ID: 1, Name: "one", Value: "first"},
		testItem{ID: 2, Name: "two", Value: "second"},
	)
	core.AssertMustNoError(t, err, "New source set")

	dst, err := cfg2.New()
	core.AssertMustNoError(t, err, "New destination set")

	result := src.Copy(dst, nil)
	core.AssertSame(t, dst, result, "should return same destination")

	// Should contain all items despite different config.
	core.AssertEqual(t, true, dst.Contains(1), "contains 1")
	core.AssertEqual(t, true, dst.Contains(2), "contains 2")
}

// TestSet_Copy_WithDifferentConfig tests copying between sets with different configurations.
func TestSet_Copy_WithDifferentConfig(t *testing.T) {
	t.Run("different hash function", runTestCopyWithDifferentConfig)
}

func runTestCopyNilSource(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	var src *set.Set[int, int, testItem]
	dst, err := cfg.New()
	core.AssertMustNoError(t, err, "New destination set")

	result := src.Copy(dst, nil)
	core.AssertSame(t, dst, result, "should return destination unchanged")
}

func runTestCopyToSelf(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	src, err := cfg.New(
		testItem{ID: 1, Name: "one", Value: "first"},
	)
	core.AssertMustNoError(t, err, "New set")

	result := src.Copy(src, nil)
	core.AssertSame(t, src, result, "should return self")
	core.AssertEqual(t, true, src.Contains(1), "should still contain item")
}

func runTestCopyUninitialisedSource(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	var src set.Set[int, int, testItem]
	dst, err := cfg.New()
	core.AssertMustNoError(t, err, "New destination set")

	result := src.Copy(dst, nil)
	core.AssertSame(t, dst, result, "should return destination unchanged")
}

func runTestCopyNilDestinationCreatesNew(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	src, err := cfg.New(
		testItem{ID: 1, Name: "one", Value: "first"},
	)
	core.AssertMustNoError(t, err, "New source set")

	result := src.Copy(nil, nil)
	core.AssertNotNil(t, result, "should create new set")
	core.AssertNotSame(t, src, result, "should be different set")
	core.AssertEqual(t, true, result.Contains(1), "new set contains item")
}

// TestSet_Copy_EdgeCases tests edge cases for Copy method.
func TestSet_Copy_EdgeCases(t *testing.T) {
	t.Run("nil source", runTestCopyNilSource)
	t.Run("copy to self", runTestCopyToSelf)
	t.Run("uninitialised source", runTestCopyUninitialisedSource)
	t.Run("nil destination creates new", runTestCopyNilDestinationCreatesNew)
}

// copyCollisionTestCase implements core.TestCase for copy with hash collision tests
type copyCollisionTestCase struct {
	filter       func(testItem) bool
	name         string
	sourceItems  []testItem
	expectIDs    []int
	notExpectIDs []int
}

func (tc copyCollisionTestCase) Name() string {
	return tc.name
}

func (tc copyCollisionTestCase) Test(t *testing.T) {
	t.Helper()
	cfg := testConfig() // Hash function uses mod 10
	src, err := cfg.New(tc.sourceItems...)
	core.AssertMustNoError(t, err, "New source set")

	dst := src.Copy(nil, tc.filter)
	core.AssertNotNil(t, dst, "Copy result")

	// Check expected items
	for _, id := range tc.expectIDs {
		core.AssertTrue(t, dst.Contains(id), "should contain %d", id)
	}

	// Check not expected items
	for _, id := range tc.notExpectIDs {
		core.AssertFalse(t, dst.Contains(id), "should not contain %d", id)
	}
}

// newCopyCollisionTestCase creates a new copyCollisionTestCase
func newCopyCollisionTestCase(name string, sourceItems []testItem,
	filter func(testItem) bool, expectIDs []int, notExpectIDs []int) copyCollisionTestCase {
	return copyCollisionTestCase{
		name:         name,
		sourceItems:  sourceItems,
		filter:       filter,
		expectIDs:    expectIDs,
		notExpectIDs: notExpectIDs,
	}
}

func makeCopyCollisionTestCases() []core.TestCase {
	collisionItems := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 11, Name: "eleven", Value: "eleventh"}, // Collides with 1
		{ID: 21, Name: "twenty-one", Value: "21st"}, // Also collides with 1
		{ID: 2, Name: "two", Value: "second"},
		{ID: 12, Name: "twelve", Value: "twelfth"}, // Collides with 2
	}

	return []core.TestCase{
		newCopyCollisionTestCase(
			"copy all with collisions",
			collisionItems,
			nil, // Copy all items
			[]int{1, 11, 21, 2, 12},
			nil,
		),
		newCopyCollisionTestCase(
			"copy filtered with collisions",
			collisionItems,
			func(item testItem) bool { return item.ID < 20 },
			[]int{1, 11, 2, 12},
			[]int{21},
		),
	}
}

// TestSet_Copy_WithHashCollisions tests copying with hash collisions.
func TestSet_Copy_WithHashCollisions(t *testing.T) {
	core.RunTestCases(t, makeCopyCollisionTestCases())
}

func runTestCopyAppendToDuplicates(t *testing.T) {
	t.Helper()
	cfg := testConfig()

	src, err := cfg.New(
		testItem{ID: 1, Name: "one-src", Value: "first"},
		testItem{ID: 2, Name: "two-src", Value: "second"},
		testItem{ID: 3, Name: "three-src", Value: "third"},
	)
	core.AssertMustNoError(t, err, "New source set")

	dst, err := cfg.New(
		testItem{ID: 1, Name: "one-dst", Value: "first-dst"},   // Duplicate ID.
		testItem{ID: 4, Name: "four-dst", Value: "fourth-dst"}, // Unique.
	)
	core.AssertMustNoError(t, err, "New destination set")

	result := src.Copy(dst, nil)
	core.AssertSame(t, dst, result, "should return same destination")

	// Should not duplicate ID 1, should add 2 and 3, keep 4.
	values := dst.Values()
	core.AssertEqual(t, 4, len(values), "should have 4 unique items")

	// Verify ID 1 kept original (destination version).
	item1, err := dst.Get(1)
	core.AssertMustNoError(t, err, "Get(1)")
	core.AssertEqual(t, "one-dst", item1.Name, "ID 1 should be destination version")

	// Verify others exist.
	core.AssertEqual(t, true, dst.Contains(2), "contains 2")
	core.AssertEqual(t, true, dst.Contains(3), "contains 3")
	core.AssertEqual(t, true, dst.Contains(4), "contains 4")
}

// TestSet_Copy_AppendToDuplicates tests appending to a set with duplicates.
func TestSet_Copy_AppendToDuplicates(t *testing.T) {
	t.Run("append preserves existing", runTestCopyAppendToDuplicates)
}

// copyEmptyTestCase implements core.TestCase for copy with empty sets tests
type copyEmptyTestCase struct {
	name        string
	sourceItems []testItem
	destItems   []testItem
	expectIDs   []int
	expectCount int
}

func (tc copyEmptyTestCase) Name() string {
	return tc.name
}

func (tc copyEmptyTestCase) Test(t *testing.T) {
	t.Helper()
	cfg := testConfig()

	src, err := cfg.New(tc.sourceItems...)
	core.AssertMustNoError(t, err, "New source set")

	dst, err := cfg.New(tc.destItems...)
	core.AssertMustNoError(t, err, "New destination set")

	result := src.Copy(dst, nil)
	core.AssertSame(t, dst, result, "should return same destination")

	// Check expected items
	for _, id := range tc.expectIDs {
		core.AssertTrue(t, dst.Contains(id), "should contain %d", id)
	}

	// Check count
	values := dst.Values()
	core.AssertEqual(t, tc.expectCount, len(values), "item count")
}

// newCopyEmptyTestCase creates a new copyEmptyTestCase
func newCopyEmptyTestCase(name string, sourceItems []testItem, destItems []testItem,
	expectIDs []int, expectCount int) copyEmptyTestCase {
	return copyEmptyTestCase{
		name:        name,
		sourceItems: sourceItems,
		destItems:   destItems,
		expectIDs:   expectIDs,
		expectCount: expectCount,
	}
}

func makeCopyEmptyTestCases() []core.TestCase {
	return []core.TestCase{
		newCopyEmptyTestCase(
			"copy empty to empty",
			[]testItem{},
			[]testItem{},
			nil,
			0,
		),
		newCopyEmptyTestCase(
			"copy non-empty to empty",
			[]testItem{{ID: 1, Name: "one", Value: "first"}},
			[]testItem{},
			[]int{1},
			1,
		),
		newCopyEmptyTestCase(
			"copy empty to non-empty",
			[]testItem{},
			[]testItem{{ID: 1, Name: "one", Value: "first"}},
			[]int{1},
			1,
		),
	}
}

// TestSet_Copy_EmptySets tests copying with empty sets.
func TestSet_Copy_EmptySets(t *testing.T) {
	core.RunTestCases(t, makeCopyEmptyTestCases())
}
