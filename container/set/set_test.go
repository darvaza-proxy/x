package set

import (
	"testing"

	"darvaza.org/core"
)

// Test types
type testItem struct {
	Name  string
	Value string
	ID    int
}

func testConfig() Config[int, int, testItem] {
	return Config[int, int, testItem]{
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

func verifySetContains(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func doTestConfigNew(t *testing.T) {
	cfg := testConfig()
	items := makeTestItems()

	s, err := cfg.New(items...)
	core.AssertMustNoError(t, err, "New")
	core.AssertNotNil(t, s, "New result")

	verifySetContains(t, s, items)
}

func TestConfigNew(t *testing.T) {
	t.Run("basic new", doTestConfigNew)
}

func doTestConfigInit(t *testing.T) {
	cfg := testConfig()
	items := makeTestItems()

	var s Set[int, int, testItem]
	err := cfg.Init(&s, items...)
	core.AssertMustNoError(t, err, "Init")

	verifySetContains(t, &s, items)
}

func doTestConfigInitTwice(t *testing.T) {
	cfg := testConfig()
	var s Set[int, int, testItem]

	err := cfg.Init(&s)
	core.AssertMustNoError(t, err, "first Init")

	// Try to init again - should fail
	err = cfg.Init(&s)
	core.AssertError(t, err, "second Init should fail")
}

func TestConfigInit(t *testing.T) {
	t.Run("basic init", doTestConfigInit)
	t.Run("init twice", doTestConfigInitTwice)
}

func testConfigMustValid(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Should not panic
	s := cfg.Must(items...)
	core.AssertNotNil(t, s, "Must result")
}

func testConfigMustPanic(t *testing.T) {
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Test panic with invalid config
	invalidCfg := Config[int, int, testItem]{}
	core.AssertPanic(t, func() {
		_ = invalidCfg.Must(items...)
	}, nil, "invalid config Must")
}

func TestConfigMust(t *testing.T) {
	t.Run("valid config", testConfigMustValid)
	t.Run("panic on invalid", testConfigMustPanic)
}

func verifyPushSuccess(t *testing.T, stored testItem, err error, expected testItem) {
	t.Helper()
	core.AssertMustNoError(t, err, "Push")
	core.AssertEqual(t, expected.ID, stored.ID, "Push returned item ID")
}

func verifyPushDuplicate(t *testing.T, err error) {
	t.Helper()
	core.AssertSame(t, ErrExist, err, "Push duplicate error")
}

//revive:disable-next-line:flag-parameter
func testPushItem(t *testing.T, s *Set[int, int, testItem], item testItem, shouldFail bool) {
	stored, err := s.Push(item)
	if shouldFail {
		verifyPushDuplicate(t, err)
	} else {
		verifyPushSuccess(t, stored, err, item)
	}
}

func testPushNew(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	testPushItem(t, s, testItem{ID: 1, Name: "one", Value: "first"}, false)
	testPushItem(t, s, testItem{ID: 2, Name: "two", Value: "second"}, false)
}

func testPushDuplicate(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one", Value: "first"})
	core.AssertMustNoError(t, err, "New")

	testPushItem(t, s, testItem{ID: 1, Name: "one-dup", Value: "duplicate"}, true)
}

func TestPush(t *testing.T) {
	t.Run("push new items", testPushNew)
	t.Run("push duplicate", testPushDuplicate)
}

func testGetExisting(t *testing.T) {
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

func testGetWithCollisions(t *testing.T) {
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

func testGetNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	_, err = s.Get(999)
	core.AssertSame(t, ErrNotExist, err, "Get(999) error")
}

func TestGet(t *testing.T) {
	t.Run("get existing", testGetExisting)
	t.Run("get with collisions", testGetWithCollisions)
	t.Run("get non-existent", testGetNonExistent)
}

func testContainsExisting(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	core.AssertMustNoError(t, err, "New")

	core.AssertTrue(t, s.Contains(1), "Contains(1)")
}

func testContainsNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	core.AssertMustNoError(t, err, "New")

	core.AssertFalse(t, s.Contains(2), "Contains(2)")
}

func TestContains(t *testing.T) {
	t.Run("contains existing", testContainsExisting)
	t.Run("contains non-existent", testContainsNonExistent)
}

func testPopExisting(t *testing.T) {
	cfg := testConfig()
	item := testItem{ID: 1, Name: "one", Value: "first"}
	s, err := cfg.New(item)
	core.AssertMustNoError(t, err, "New")

	v, err := s.Pop(1)
	core.AssertMustNoError(t, err, "Pop")
	core.AssertEqual(t, item.ID, v.ID, "Pop item ID")

	core.AssertFalse(t, s.Contains(1), "item removed after Pop")
}

func testPopNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	_, err = s.Pop(1)
	core.AssertSame(t, ErrNotExist, err, "Pop non-existent error")
}

func TestPop(t *testing.T) {
	t.Run("pop existing", testPopExisting)
	t.Run("pop non-existent", testPopNonExistent)
}

func doTestReset(t *testing.T) {
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
	t.Run("reset set", doTestReset)
}

func doTestClone(t *testing.T) {
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

func TestClone(t *testing.T) {
	t.Run("clone set", doTestClone)
}

func testCopyFilter(t *testing.T) {
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

func testCopyAll(t *testing.T) {
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
	t.Run("copy with filter", testCopyFilter)
	t.Run("copy all", testCopyAll)
}

func doTestValues(t *testing.T) {
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

func TestValues(t *testing.T) {
	t.Run("get values", doTestValues)
}

func testForEachAll(t *testing.T) {
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

func testForEachEarlyTermination(t *testing.T) {
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

func TestForEach(t *testing.T) {
	t.Run("all items", testForEachAll)
	t.Run("early termination", testForEachEarlyTermination)
}

func testConfigValidationMissingItemKey(t *testing.T) {
	cfg := Config[int, int, testItem]{
		Hash:      func(k int) (int, error) { return k, nil },
		ItemMatch: func(_ int, _ testItem) bool { return true },
	}

	err := cfg.Validate()
	core.AssertError(t, err, "Validate should fail without ItemKey function")
}

func testConfigValidationMissingHash(t *testing.T) {
	cfg := Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		ItemMatch: func(_ int, _ testItem) bool { return true },
	}

	err := cfg.Validate()
	core.AssertError(t, err, "Validate should fail without Hash function")
}

func testConfigValidationMissingItemMatch(t *testing.T) {
	cfg := Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) { return v.ID, nil },
		Hash:    func(k int) (int, error) { return k, nil },
	}

	err := cfg.Validate()
	core.AssertError(t, err, "Validate should fail without ItemMatch function")
}

func testConfigValidationValid(t *testing.T) {
	cfg := testConfig()
	err := cfg.Validate()
	core.AssertMustNoError(t, err, "Validate valid config")
}

func TestConfigValidation(t *testing.T) {
	t.Run("missing ItemKey", testConfigValidationMissingItemKey)
	t.Run("missing Hash", testConfigValidationMissingHash)
	t.Run("missing ItemMatch", testConfigValidationMissingItemMatch)
	t.Run("valid config", testConfigValidationValid)
}

func testConfigEqualSame(t *testing.T) {
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	hashFn := func(k int) (int, error) { return k, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }

	cfg1 := Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}
	cfg2 := Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}

	core.AssertTrue(t, cfg1.Equal(cfg2), "configs with same functions should be equal")
}

func testConfigEqualDifferent(t *testing.T) {
	cfg1 := testConfig()
	cfg2 := testConfig() // Different function instances

	core.AssertFalse(t, cfg1.Equal(cfg2), "configs with different function instances should not be equal")
}

func TestConfigEqual(t *testing.T) {
	t.Run("same functions", testConfigEqualSame)
	t.Run("different functions", testConfigEqualDifferent)
}

func runConcurrentWriter(s *Set[int, int, testItem], done chan bool) {
	for i := 0; i < 100; i++ {
		_, _ = s.Push(testItem{ID: i, Name: "item", Value: "value"})
	}
	done <- true
}

func runConcurrentReader(s *Set[int, int, testItem], done chan bool) {
	for i := 0; i < 100; i++ {
		s.Contains(i)
		_, _ = s.Get(i)
	}
	done <- true
}

func runConcurrentForEach(s *Set[int, int, testItem], done chan bool) {
	for i := 0; i < 10; i++ {
		s.ForEach(func(testItem) bool { return true })
	}
	done <- true
}

func doTestThreadSafety(t *testing.T) {
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
	t.Run("concurrent operations", doTestThreadSafety)
}

func testNilReceiverInit(t *testing.T) {
	var s *Set[int, int, testItem]
	cfg := testConfig()

	err := cfg.Init(s)
	core.AssertError(t, err, "Init on nil receiver should return error")
}

func testNilReceiverMethods(t *testing.T) {
	var s *Set[int, int, testItem]

	core.AssertFalse(t, s.Contains(1), "nil set Contains should return false")

	_, err := s.Get(1)
	core.AssertTrue(t, err != nil && core.IsError(err, core.ErrNilReceiver), "nil set Get should return ErrNilReceiver")

	_, err = s.Push(testItem{ID: 1})
	core.AssertTrue(t, err != nil && core.IsError(err, core.ErrNilReceiver), "nil set Push should return ErrNilReceiver")
}

func TestNilReceiver(t *testing.T) {
	t.Run("init nil receiver", testNilReceiverInit)
	t.Run("methods on nil", testNilReceiverMethods)
}

func doTestEmptySet(t *testing.T) {
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
	t.Run("empty set operations", doTestEmptySet)
}

func makeHashCollisionItems() []testItem {
	return []testItem{
		{ID: 1, Name: "one"},
		{ID: 11, Name: "eleven"},
		{ID: 21, Name: "twenty-one"},
		{ID: 31, Name: "thirty-one"},
	}
}

func verifyPushItems(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		_, err := s.Push(item)
		core.AssertMustNoError(t, err, "Push(%d)", item.ID)
	}
}

func verifyGetItems(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		v, err := s.Get(item.ID)
		core.AssertMustNoError(t, err, "Get(%d)", item.ID)
		core.AssertEqual(t, item.ID, v.ID, "Get(%d) item ID", item.ID)
	}
}

func verifyContainsItems(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	t.Helper()
	for _, item := range items {
		core.AssertTrue(t, s.Contains(item.ID), "Contains(%d)", item.ID)
	}
}

func doTestHashCollisions(t *testing.T) {
	cfg := testConfig() // Uses hash function that returns k % 10
	s, err := cfg.New()
	core.AssertMustNoError(t, err, "New")

	items := makeHashCollisionItems()
	verifyPushItems(t, s, items)
	verifyGetItems(t, s, items)
	verifyContainsItems(t, s, items)
}

func TestHashCollisions(t *testing.T) {
	t.Run("hash collisions", doTestHashCollisions)
}
