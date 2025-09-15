package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

// Interface validation
var _ core.TestCase = configValidationTestCase{}
var _ core.TestCase = configEqualTestCase{}

// Test types for external package tests
type testItem struct {
	Name  string
	Value string
	ID    int
}

func testConfig() set.Config[int, int, testItem] {
	return set.Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}
}

func makeTestItems() []testItem {
	return []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
	}
}

func verifySetContains(t *testing.T, s *set.Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func runTestConfigNew(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := makeTestItems()

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")
	core.AssertNotNil(t, s, "New result")

	verifySetContains(t, s, items)
}

func TestConfigNew(t *testing.T) {
	t.Run("basic new", runTestConfigNew)
}

func runTestConfigInit(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := makeTestItems()

	var s set.Set[int, int, testItem]
	err := cfg.Init(&s, items...)
	core.AssertMustNoError(t, err, "Init")

	verifySetContains(t, &s, items)
}

func runTestConfigInitTwice(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	var s set.Set[int, int, testItem]

	err := cfg.Init(&s)
	core.AssertMustNoError(t, err, "first Init")

	// Try to init again - should fail
	err = cfg.Init(&s)
	core.AssertError(t, err, "second Init should fail")
}

func TestConfigInit(t *testing.T) {
	t.Run("basic init", runTestConfigInit)
	t.Run("init twice", runTestConfigInitTwice)
}

func runTestConfigMustValid(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Should not panic
	s := cfg.Must(items...)
	core.AssertNotNil(t, s, "Must result")
}

func runTestConfigMustPanic(t *testing.T) {
	t.Helper()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Test panic with invalid config
	invalidCfg := set.Config[int, int, testItem]{}
	core.AssertPanic(t, func() {
		_ = invalidCfg.Must(items...)
	}, nil, "invalid config Must")
}

func TestConfigMust(t *testing.T) {
	t.Run("valid config", runTestConfigMustValid)
	t.Run("panic on invalid", runTestConfigMustPanic)
}

func verifyPushSuccess(t *testing.T, stored testItem, err error, expected testItem) {
	t.Helper()
	core.AssertMustNoError(t, err, "Push")
	core.AssertEqual(t, expected.ID, stored.ID, "Push returned item ID")
}

func verifyPushDuplicate(t *testing.T, err error) {
	t.Helper()
	core.AssertSame(t, set.ErrExist, err, "Push duplicate error")
}

//revive:disable-next-line:flag-parameter
func testPushItem(t *testing.T, s *set.Set[int, int, testItem], item testItem, shouldFail bool) {
	t.Helper()
	stored, err := s.Push(item)
	if shouldFail {
		verifyPushDuplicate(t, err)
	} else {
		verifyPushSuccess(t, stored, err, item)
	}
}

func runTestPushNew(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	testPushItem(t, s, testItem{ID: 1, Name: "one", Value: "first"}, false)
	testPushItem(t, s, testItem{ID: 2, Name: "two", Value: "second"}, false)
}

func runTestPushDuplicate(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one", Value: "first"})
	core.AssertMustNoError(t, err, "New")

	testPushItem(t, s, testItem{ID: 1, Name: "one-dup", Value: "duplicate"}, true)
}

func TestPush(t *testing.T) {
	t.Run("push new items", runTestPushNew)
	t.Run("push duplicate", runTestPushDuplicate)
}

func runTestGetExisting(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
	}

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func runTestGetWithCollisions(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 11, Name: "eleven", Value: "collision"}, // Same hash as 1
	}

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func runTestGetNonExistent(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	_, err = s.Get(999)
	core.AssertSame(t, set.ErrNotExist, err, "Get(999) error")
}

func TestGet(t *testing.T) {
	t.Run("get existing", runTestGetExisting)
	t.Run("get with collisions", runTestGetWithCollisions)
	t.Run("get non-existent", runTestGetNonExistent)
}

func runTestContainsExisting(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	core.AssertMustNoError(t, err, "New")

	core.AssertTrue(t, s.Contains(1), "Contains(1)")
}

func runTestContainsNonExistent(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	core.AssertMustNoError(t, err, "New")

	core.AssertFalse(t, s.Contains(2), "Contains(2)")
}

func TestContains(t *testing.T) {
	t.Run("contains existing", runTestContainsExisting)
	t.Run("contains non-existent", runTestContainsNonExistent)
}

func runTestPopExisting(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	item := testItem{ID: 1, Name: "one", Value: "first"}
	s, err := cfg.New(item)
	core.AssertMustNoError(t, err, "New")

	v, err := s.Pop(1)
	core.AssertMustNoError(t, err, "Pop")
	core.AssertEqual(t, item.ID, v.ID, "Pop item ID")

	core.AssertFalse(t, s.Contains(1), "item removed after Pop")
}

func runTestPopNonExistent(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	_, err = s.Pop(1)
	core.AssertSame(t, set.ErrNotExist, err, "Pop non-existent error")
}

func TestPop(t *testing.T) {
	t.Run("pop existing", runTestPopExisting)
	t.Run("pop non-existent", runTestPopNonExistent)
}

func runTestReset(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 2, Name: "two"},
		testItem{ID: 3, Name: "three"},
	)
	core.AssertMustNoError(t, err, "New")

	// Verify items exist
	core.AssertTrue(t, s.Contains(1) && s.Contains(2) && s.Contains(3), "items exist before reset")

	// Reset
	err = s.Reset()
	core.AssertMustNoError(t, err, "Reset")

	// Verify all items are removed
	core.AssertFalse(t, s.Contains(1) || s.Contains(2) || s.Contains(3), "items removed after reset")
}

func TestReset(t *testing.T) {
	t.Run("reset set", runTestReset)
}

func runTestClone(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
	}

	s1, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	s2 := s1.Clone()
	core.AssertNotNil(t, s2, "Clone result")

	// Verify clone has same items
	for _, item := range items {
		core.AssertTrue(t, s2.Contains(item.ID), "Clone contains item %d", item.ID)
	}

	// Modify original
	_, err = s1.Push(testItem{ID: 3, Name: "three", Value: "third"})
	core.AssertMustNoError(t, err, "Push")

	// Verify clone is independent
	core.AssertFalse(t, s2.Contains(3), "clone independence")
}

func runTestCloneNilReceiver(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]
	clone := s.Clone()
	core.AssertNil(t, clone, "Clone on nil receiver")
}

func runTestCloneUninitialized(t *testing.T) {
	t.Helper()
	s := &set.Set[int, int, testItem]{}
	clone := s.Clone()
	core.AssertNil(t, clone, "Clone on uninitialised set")
}

func TestClone(t *testing.T) {
	t.Run("clone set", runTestClone)
	t.Run("nil receiver", runTestCloneNilReceiver)
	t.Run("uninitialized", runTestCloneUninitialized)
}

func runTestCopyFilter(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s1, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	// Copy only even IDs
	s2 := s1.Copy(nil, func(v testItem) bool {
		return v.ID%2 == 0
	})

	core.AssertFalse(t, s2.Contains(1) || s2.Contains(3), "copy should not contain odd IDs")
	core.AssertTrue(t, s2.Contains(2), "copy should contain even IDs")
}

func runTestCopyAll(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s1, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	// Copy all items
	s2 := s1.Copy(nil, nil)

	for _, item := range items {
		v, err := s2.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.Value, v.Value, "Get(%d) value", item.ID)
	}
}

func TestCopy(t *testing.T) {
	t.Run("copy with filter", runTestCopyFilter)
	t.Run("copy all", runTestCopyAll)
}

func runTestValues(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	values := s.Values()
	core.AssertEqual(t, len(items), len(values), "Values count")

	// Verify all items are present
	found := make(map[int]bool)
	for _, v := range values {
		found[v.ID] = true
	}

	for _, item := range items {
		core.AssertTrue(t, found[item.ID], "Values contains item %d", item.ID)
	}
}

func runTestValuesNilReceiver(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]
	values := s.Values()
	core.AssertNil(t, values, "Values on nil receiver")
}

func runTestValuesUninitialized(t *testing.T) {
	t.Helper()
	s := &set.Set[int, int, testItem]{}
	values := s.Values()
	core.AssertNil(t, values, "Values on uninitialised set")
}

func TestValues(t *testing.T) {
	t.Run("get values", runTestValues)
	t.Run("nil receiver", runTestValuesNilReceiver)
	t.Run("uninitialized", runTestValuesUninitialized)
}

func runTestForEachAll(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return true // continue
	})

	core.AssertEqual(t, len(items), count, "ForEach visited count")
}

func runTestForEachEarlyTermination(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")

	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return count < 2 // stop after 2 items
	})

	core.AssertEqual(t, 2, count, "ForEach early termination count")
}

func runTestForEachNilReceiver(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]
	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return true
	})
	core.AssertEqual(t, 0, count, "ForEach on nil receiver should not iterate")
}

func runTestForEachUninitialized(t *testing.T) {
	t.Helper()
	s := &set.Set[int, int, testItem]{}
	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return true
	})
	core.AssertEqual(t, 0, count, "ForEach on uninitialised set should not iterate")
}

func runTestForEachNilFunction(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	core.AssertMustNoError(t, err, "New")

	// Should not panic with nil function
	s.ForEach(nil)
	// Test passes if no panic
}

func TestForEach(t *testing.T) {
	t.Run("all items", runTestForEachAll)
	t.Run("early termination", runTestForEachEarlyTermination)
	t.Run("nil receiver", runTestForEachNilReceiver)
	t.Run("uninitialized", runTestForEachUninitialized)
	t.Run("nil function", runTestForEachNilFunction)
}

// configValidationTestCase implements core.TestCase for config validation tests
type configValidationTestCase struct {
	config      set.Config[int, int, testItem]
	name        string
	errorMsg    string
	expectError bool
}

func (tc configValidationTestCase) Name() string {
	return tc.name
}

func (tc configValidationTestCase) Test(t *testing.T) {
	t.Helper()
	err := tc.config.Validate()
	if tc.expectError {
		core.AssertError(t, err, tc.errorMsg)
	} else {
		core.AssertMustNoError(t, err, tc.errorMsg)
	}
}

// newConfigValidationTestCase creates a new configValidationTestCase
//
//revive:disable-next-line:argument-limit
func newConfigValidationTestCase(name string, config set.Config[int, int, testItem],
	expectError bool, errorMsg string) configValidationTestCase {
	return configValidationTestCase{
		name:        name,
		config:      config,
		expectError: expectError,
		errorMsg:    errorMsg,
	}
}

func makeConfigValidationTestCases() []core.TestCase {
	return []core.TestCase{
		newConfigValidationTestCase(
			"missing ItemKey",
			set.Config[int, int, testItem]{
				Hash:      func(k int) (int, error) { return k, nil },
				ItemMatch: func(_ int, _ testItem) bool { return true },
			},
			true,
			"Validate should fail without ItemKey function",
		),
		newConfigValidationTestCase(
			"missing Hash",
			set.Config[int, int, testItem]{
				ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
				ItemMatch: func(_ int, _ testItem) bool { return true },
			},
			true,
			"Validate should fail without Hash function",
		),
		newConfigValidationTestCase(
			"missing ItemMatch",
			set.Config[int, int, testItem]{
				ItemKey: func(v testItem) (int, error) { return v.ID, nil },
				Hash:    func(k int) (int, error) { return k, nil },
			},
			true,
			"Validate should fail without ItemMatch function",
		),
		newConfigValidationTestCase(
			"valid config",
			testConfig(),
			false,
			"Validate valid config",
		),
	}
}

func TestConfigValidation(t *testing.T) {
	core.RunTestCases(t, makeConfigValidationTestCases())
}

// configEqualTestCase implements core.TestCase for config equality tests
type configEqualTestCase struct {
	makeFunc func() (set.Config[int, int, testItem], set.Config[int, int, testItem])
	name     string
	errorMsg string
	expected bool
}

func (tc configEqualTestCase) Name() string {
	return tc.name
}

func (tc configEqualTestCase) Test(t *testing.T) {
	t.Helper()
	cfg1, cfg2 := tc.makeFunc()
	result := cfg1.Equal(cfg2)
	if tc.expected {
		core.AssertTrue(t, result, tc.errorMsg)
	} else {
		core.AssertFalse(t, result, tc.errorMsg)
	}
}

// newConfigEqualTestCase creates a new configEqualTestCase
//
//revive:disable-next-line:flag-parameter
func newConfigEqualTestCase(name string,
	makeFunc func() (set.Config[int, int, testItem], set.Config[int, int, testItem]),
	expected bool, errorMsg string) configEqualTestCase {
	return configEqualTestCase{
		name:     name,
		makeFunc: makeFunc,
		expected: expected,
		errorMsg: errorMsg,
	}
}

func makeConfigWithSameFunctions() (cfg1, cfg2 set.Config[int, int, testItem]) {
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	hashFn := func(k int) (int, error) { return k, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }

	cfg1 = set.Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}
	cfg2 = set.Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}
	return cfg1, cfg2
}

func makeConfigWithMixedNil() (cfg1, cfg2 set.Config[int, int, testItem]) {
	cfg1 = set.Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
		// Hash is nil
	}
	cfg2 = set.Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) { return v.ID, nil },
		Hash:    func(k int) (int, error) { return k, nil },
		// ItemMatch is nil
	}
	return cfg1, cfg2
}

func makeConfigEqualTestCases() []core.TestCase {
	// Shared function instances
	hashFn := func(k int) (int, error) { return k, nil }
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }

	return []core.TestCase{
		newConfigEqualTestCase(
			"same functions",
			makeConfigWithSameFunctions,
			true,
			"configs with same functions should be equal",
		),
		newConfigEqualTestCase(
			"different function instances",
			func() (set.Config[int, int, testItem], set.Config[int, int, testItem]) {
				return testConfig(), testConfig()
			},
			false,
			"configs with different function instances should not be equal",
		),
		newConfigEqualTestCase(
			"both nil functions",
			func() (set.Config[int, int, testItem], set.Config[int, int, testItem]) {
				return set.Config[int, int, testItem]{}, set.Config[int, int, testItem]{}
			},
			true,
			"configs with all nil functions should be equal",
		),
		newConfigEqualTestCase(
			"one nil one not",
			func() (set.Config[int, int, testItem], set.Config[int, int, testItem]) {
				return set.Config[int, int, testItem]{}, testConfig()
			},
			false,
			"config with nil functions should not equal config with functions",
		),
		newConfigEqualTestCase(
			"mixed nil non-nil",
			makeConfigWithMixedNil,
			false,
			"configs with different nil fields should not be equal",
		),
		newConfigEqualTestCase(
			"different ItemKey",
			func() (set.Config[int, int, testItem], set.Config[int, int, testItem]) {
				cfg1 := set.Config[int, int, testItem]{
					Hash:      hashFn,
					ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
					ItemMatch: matchFn,
				}
				cfg2 := set.Config[int, int, testItem]{
					Hash:      hashFn,
					ItemKey:   func(v testItem) (int, error) { return v.ID + 1, nil },
					ItemMatch: matchFn,
				}
				return cfg1, cfg2
			},
			false,
			"configs with different ItemKey should not be equal",
		),
		newConfigEqualTestCase(
			"different ItemMatch",
			func() (set.Config[int, int, testItem], set.Config[int, int, testItem]) {
				cfg1 := set.Config[int, int, testItem]{
					Hash:      hashFn,
					ItemKey:   keyFn,
					ItemMatch: func(k int, v testItem) bool { return k == v.ID },
				}
				cfg2 := set.Config[int, int, testItem]{
					Hash:      hashFn,
					ItemKey:   keyFn,
					ItemMatch: func(_ int, _ testItem) bool { return false },
				}
				return cfg1, cfg2
			},
			false,
			"configs with different ItemMatch should not be equal",
		),
	}
}

func TestConfigEqual(t *testing.T) {
	core.RunTestCases(t, makeConfigEqualTestCases())
}

func runConcurrentWriter(s *set.Set[int, int, testItem], done chan bool) {
	for i := 0; i < 100; i++ {
		_, _ = s.Push(testItem{ID: i, Name: "item", Value: "value"})
	}
	done <- true
}

func runConcurrentReader(s *set.Set[int, int, testItem], done chan bool) {
	for i := 0; i < 100; i++ {
		s.Contains(i)
		_, _ = s.Get(i)
	}
	done <- true
}

func runConcurrentForEach(s *set.Set[int, int, testItem], done chan bool) {
	for range 10 {
		s.ForEach(func(testItem) bool { return true })
	}
	done <- true
}

func runTestThreadSafety(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	// Run concurrent operations
	done := make(chan bool)

	go runConcurrentWriter(s, done)
	go runConcurrentReader(s, done)
	go runConcurrentForEach(s, done)

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestThreadSafety(t *testing.T) {
	t.Run("concurrent operations", runTestThreadSafety)
}

func runTestNilReceiverInit(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]
	cfg := testConfig()

	err := cfg.Init(s)
	core.AssertError(t, err, "Init on nil receiver should return error")
}

func runTestNilReceiverMethods(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]

	core.AssertFalse(t, s.Contains(1), "nil set Contains should return false")

	_, err := s.Get(1)
	core.AssertError(t, err, "nil set Get should return error")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error type")

	_, err = s.Push(testItem{ID: 1})
	core.AssertError(t, err, "nil set Push should return error")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error type")
}

func TestNilReceiver(t *testing.T) {
	t.Run("init nil receiver", runTestNilReceiverInit)
	t.Run("methods on nil", runTestNilReceiverMethods)
}

func runTestEmptySet(t *testing.T) {
	t.Helper()
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	// Test operations on empty set
	values := s.Values()
	core.AssertEqual(t, 0, len(values), "empty set Values()")

	count := 0
	s.ForEach(func(testItem) bool {
		count++
		return true
	})
	core.AssertEqual(t, 0, count, "empty set ForEach iterations")

	core.AssertFalse(t, s.Contains(1), "empty set Contains")

	// Test Pop on empty set
	_, err = s.Pop(1)
	core.AssertError(t, err, "Pop on empty set should return error")

	// Test Get on empty set
	_, err = s.Get(1)
	core.AssertError(t, err, "Get on empty set should return error")
}

func TestEmptySet(t *testing.T) {
	t.Run("empty set operations", runTestEmptySet)
}

func makeHashCollisionItems() []testItem {
	return []testItem{
		{ID: 1, Name: "one"},
		{ID: 11, Name: "eleven"},
		{ID: 21, Name: "twenty-one"},
		{ID: 31, Name: "thirty-one"},
	}
}

func verifyPushItems(t *testing.T, s *set.Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		_, err := s.Push(item)
		core.AssertMustNoError(t, err, "Push(%d)", item.ID)
	}
}

func verifyGetItems(t *testing.T, s *set.Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func verifyContainsItems(t *testing.T, s *set.Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		core.AssertTrue(t, s.Contains(item.ID), "Contains(%d)", item.ID)
	}
}

func runTestHashCollisions(t *testing.T) {
	t.Helper()
	cfg := testConfig() // Uses hash function that returns k % 10
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	items := makeHashCollisionItems()
	verifyPushItems(t, s, items)
	verifyGetItems(t, s, items)
	verifyContainsItems(t, s, items)
}

func TestHashCollisions(t *testing.T) {
	t.Run("hash collisions", runTestHashCollisions)
}

// Tests for achieving 100% coverage

func runTestResetNilReceiver(t *testing.T) {
	t.Helper()
	var s *set.Set[int, int, testItem]
	err := s.Reset()
	core.AssertError(t, err, "Reset on nil receiver")
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "error type")
}

func runTestResetUninitialized(t *testing.T) {
	t.Helper()
	s := &set.Set[int, int, testItem]{}
	err := s.Reset()
	core.AssertError(t, err, "Reset on uninitialised set")
	core.AssertErrorIs(t, err, core.ErrInvalid, "error type")
}

func TestResetEdgeCases(t *testing.T) {
	t.Run("nil receiver", runTestResetNilReceiver)
	t.Run("uninitialized", runTestResetUninitialized)
}

func runTestUnsafeInitPushWithError(t *testing.T) {
	t.Helper()
	// Create a config that will cause errors
	errorConfig := set.Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) {
			if v.ID == 999 {
				return 0, core.ErrInvalid // This will cause Push to fail
			}
			return v.ID, nil
		},
		Hash:      func(k int) (int, error) { return k % 10, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}

	s := &set.Set[int, int, testItem]{}
	err := errorConfig.Init(s, testItem{ID: 1}, testItem{ID: 999})
	core.AssertError(t, err, "unsafeInitPush with error item")
}

func TestUnsafeInitPush(t *testing.T) {
	t.Run("with error", runTestUnsafeInitPushWithError)
}

func runTestCheckWithValueItemKeyError(t *testing.T) {
	t.Helper()
	errorConfig := set.Config[int, int, testItem]{
		ItemKey: func(_ testItem) (int, error) {
			return 0, core.ErrInvalid
		},
		Hash:      func(k int) (int, error) { return k, nil },
		ItemMatch: func(k int, v testItem) bool { return k == v.ID },
	}

	s, err := errorConfig.New()
	core.AssertMustNoError(t, err, "New")

	_, err = s.Push(testItem{ID: 1})
	core.AssertError(t, err, "Push with ItemKey error")
}

func TestCheckWithValueItemKeyError(t *testing.T) {
	t.Run("ItemKey error", runTestCheckWithValueItemKeyError)
}
