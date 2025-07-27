package set

import (
	"testing"

	"darvaza.org/core"
)

// Test types
type testItem struct {
	ID    int
	Name  string
	Value string
}

// testCase represents a single test case
type testCase struct {
	name string
	test func(t *testing.T)
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
	for _, item := range items {
		if v, err := s.Get(item.ID); err != nil {
			t.Errorf("Get(%d) failed: %v", item.ID, err)
		} else if v.ID != item.ID {
			t.Errorf("Get(%d) returned wrong item: %+v", item.ID, v)
		}
	}
}

func doTestConfigNew(t *testing.T) {
	cfg := testConfig()
	items := makeTestItems()

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if s == nil {
		t.Fatal("New returned nil")
	}

	verifySetContains(t, s, items)
}

func TestConfigNew(t *testing.T) {
	tests := []testCase{
		{"basic new", doTestConfigNew},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func doTestConfigInit(t *testing.T) {
	cfg := testConfig()
	items := makeTestItems()

	var s Set[int, int, testItem]
	if err := cfg.Init(&s, items...); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	verifySetContains(t, &s, items)
}

func doTestConfigInitTwice(t *testing.T) {
	cfg := testConfig()
	var s Set[int, int, testItem]

	if err := cfg.Init(&s); err != nil {
		t.Fatalf("First Init failed: %v", err)
	}

	// Try to init again - should fail
	if err := cfg.Init(&s); err == nil {
		t.Error("Second Init should fail")
	}
}

func TestConfigInit(t *testing.T) {
	tests := []testCase{
		{"basic init", doTestConfigInit},
		{"init twice", doTestConfigInitTwice},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testConfigMustValid(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Should not panic
	s := cfg.Must(items...)
	if s == nil {
		t.Fatal("Must returned nil")
	}
}

func testConfigMustPanic(t *testing.T) {
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
	}

	// Test panic with invalid config
	defer func() {
		if r := recover(); r == nil {
			t.Error("Must should panic with invalid config")
		}
	}()

	// This should panic
	invalidCfg := Config[int, int, testItem]{}
	_ = invalidCfg.Must(items...)
}

func TestConfigMust(t *testing.T) {
	tests := []testCase{
		{"valid config", testConfigMustValid},
		{"panic on invalid", testConfigMustPanic},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func verifyPushSuccess(t *testing.T, stored testItem, err error, expected testItem) {
	if err != nil {
		t.Errorf("Push failed: %v", err)
	} else if stored.ID != expected.ID {
		t.Errorf("Push returned wrong item: %+v", stored)
	}
}

func verifyPushDuplicate(t *testing.T, err error) {
	if err != ErrExist {
		t.Errorf("Push duplicate should return ErrExist, got: %v", err)
	}
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
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	testPushItem(t, s, testItem{ID: 1, Name: "one", Value: "first"}, false)
	testPushItem(t, s, testItem{ID: 2, Name: "two", Value: "second"}, false)
}

func testPushDuplicate(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one", Value: "first"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	testPushItem(t, s, testItem{ID: 1, Name: "one-dup", Value: "duplicate"}, true)
}

func TestPush(t *testing.T) {
	tests := []testCase{
		{"push new items", testPushNew},
		{"push duplicate", testPushDuplicate},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testGetExisting(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
	}

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for _, item := range items {
		if v, err := s.Get(item.ID); err != nil {
			t.Errorf("Get(%d) failed: %v", item.ID, err)
		} else if v.ID != item.ID {
			t.Errorf("Get(%d) returned wrong item: %+v", item.ID, v)
		}
	}
}

func testGetWithCollisions(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 11, Name: "eleven", Value: "collision"}, // Same hash as 1
	}

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for _, item := range items {
		if v, err := s.Get(item.ID); err != nil {
			t.Errorf("Get(%d) failed: %v", item.ID, err)
		} else if v.ID != item.ID {
			t.Errorf("Get(%d) returned wrong item: %+v", item.ID, v)
		}
	}
}

func testGetNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if _, err := s.Get(999); err != ErrNotExist {
		t.Errorf("Get(999) should return ErrNotExist, got: %v", err)
	}
}

func TestGet(t *testing.T) {
	tests := []testCase{
		{"get existing", testGetExisting},
		{"get with collisions", testGetWithCollisions},
		{"get non-existent", testGetNonExistent},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testContainsExisting(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if !s.Contains(1) {
		t.Error("Contains(1) should return true")
	}
}

func testContainsNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(testItem{ID: 1, Name: "one"})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if s.Contains(2) {
		t.Error("Contains(2) should return false")
	}
}

func TestContains(t *testing.T) {
	tests := []testCase{
		{"contains existing", testContainsExisting},
		{"contains non-existent", testContainsNonExistent},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testPopExisting(t *testing.T) {
	cfg := testConfig()
	item := testItem{ID: 1, Name: "one", Value: "first"}
	s, err := cfg.New(item)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if v, err := s.Pop(1); err != nil {
		t.Errorf("Pop failed: %v", err)
	} else if v.ID != item.ID {
		t.Errorf("Pop returned wrong item: %+v", v)
	}

	if s.Contains(1) {
		t.Error("Item should be removed after Pop")
	}
}

func testPopNonExistent(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if _, err := s.Pop(1); err != ErrNotExist {
		t.Errorf("Pop non-existent should return ErrNotExist, got: %v", err)
	}
}

func TestPop(t *testing.T) {
	tests := []testCase{
		{"pop existing", testPopExisting},
		{"pop non-existent", testPopNonExistent},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func doTestReset(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New(
		testItem{ID: 1, Name: "one"},
		testItem{ID: 2, Name: "two"},
		testItem{ID: 3, Name: "three"},
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Verify items exist
	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		t.Error("Items should exist before reset")
	}

	// Reset
	if err := s.Reset(); err != nil {
		t.Errorf("Reset failed: %v", err)
	}

	// Verify all items are removed
	if s.Contains(1) || s.Contains(2) || s.Contains(3) {
		t.Error("Items should not exist after reset")
	}
}

func TestReset(t *testing.T) {
	tests := []testCase{
		{"reset set", doTestReset},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func doTestClone(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
	}

	s1, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	s2 := s1.Clone()
	if s2 == nil {
		t.Fatal("Clone returned nil")
	}

	// Verify clone has same items
	for _, item := range items {
		if !s2.Contains(item.ID) {
			t.Errorf("Clone missing item %d", item.ID)
		}
	}

	// Modify original
	if _, err := s1.Push(testItem{ID: 3, Name: "three", Value: "third"}); err != nil {
		t.Errorf("Push failed: %v", err)
	}

	// Verify clone is independent
	if s2.Contains(3) {
		t.Error("Clone should not be affected by original modification")
	}
}

func TestClone(t *testing.T) {
	tests := []testCase{
		{"clone set", doTestClone},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testCopyFilter(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s1, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Copy only even IDs
	s2 := s1.Copy(nil, func(v testItem) bool {
		return v.ID%2 == 0
	})

	if s2.Contains(1) || s2.Contains(3) {
		t.Error("Copy should not contain odd IDs")
	}

	if !s2.Contains(2) {
		t.Error("Copy should contain even IDs")
	}
}

func testCopyAll(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s1, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Copy all items
	s2 := s1.Copy(nil, nil)

	for _, item := range items {
		if v, err := s2.Get(item.ID); err != nil {
			t.Errorf("Get(%d) failed: %v", item.ID, err)
		} else if v.Value != item.Value {
			t.Errorf("Value mismatch: got %s, expected %s", v.Value, item.Value)
		}
	}
}

func TestCopy(t *testing.T) {
	tests := []testCase{
		{"copy with filter", testCopyFilter},
		{"copy all", testCopyAll},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func doTestValues(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	values := s.Values()
	if len(values) != len(items) {
		t.Errorf("Values returned %d items, expected %d", len(values), len(items))
	}

	// Verify all items are present
	found := make(map[int]bool)
	for _, v := range values {
		found[v.ID] = true
	}

	for _, item := range items {
		if !found[item.ID] {
			t.Errorf("Values missing item %d", item.ID)
		}
	}
}

func TestValues(t *testing.T) {
	tests := []testCase{
		{"get values", doTestValues},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testForEachAll(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return true // continue
	})

	if count != len(items) {
		t.Errorf("ForEach visited %d items, expected %d", count, len(items))
	}
}

func testForEachEarlyTermination(t *testing.T) {
	cfg := testConfig()
	items := []testItem{
		{ID: 1, Name: "one", Value: "first"},
		{ID: 2, Name: "two", Value: "second"},
		{ID: 3, Name: "three", Value: "third"},
	}

	s, err := cfg.New(items...)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	count := 0
	s.ForEach(func(_ testItem) bool {
		count++
		return count < 2 // stop after 2 items
	})

	if count != 2 {
		t.Errorf("ForEach should stop after 2 items, visited %d", count)
	}
}

func TestForEach(t *testing.T) {
	tests := []testCase{
		{"all items", testForEachAll},
		{"early termination", testForEachEarlyTermination},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testConfigValidationMissingItemKey(t *testing.T) {
	cfg := Config[int, int, testItem]{
		Hash:      func(k int) (int, error) { return k, nil },
		ItemMatch: func(_ int, _ testItem) bool { return true },
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate should fail without ItemKey function")
	}
}

func testConfigValidationMissingHash(t *testing.T) {
	cfg := Config[int, int, testItem]{
		ItemKey:   func(v testItem) (int, error) { return v.ID, nil },
		ItemMatch: func(_ int, _ testItem) bool { return true },
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate should fail without Hash function")
	}
}

func testConfigValidationMissingItemMatch(t *testing.T) {
	cfg := Config[int, int, testItem]{
		ItemKey: func(v testItem) (int, error) { return v.ID, nil },
		Hash:    func(k int) (int, error) { return k, nil },
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate should fail without ItemMatch function")
	}
}

func testConfigValidationValid(t *testing.T) {
	cfg := testConfig()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate failed on valid config: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []testCase{
		{"missing ItemKey", testConfigValidationMissingItemKey},
		{"missing Hash", testConfigValidationMissingHash},
		{"missing ItemMatch", testConfigValidationMissingItemMatch},
		{"valid config", testConfigValidationValid},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testConfigEqualSame(t *testing.T) {
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	hashFn := func(k int) (int, error) { return k, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }

	cfg1 := Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}
	cfg2 := Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}

	if !cfg1.Equal(cfg2) {
		t.Error("Configs with same functions should be equal")
	}
}

func testConfigEqualDifferent(t *testing.T) {
	cfg1 := testConfig()
	cfg2 := testConfig() // Different function instances

	if cfg1.Equal(cfg2) {
		t.Error("Configs with different function instances should not be equal")
	}
}

func TestConfigEqual(t *testing.T) {
	tests := []testCase{
		{"same functions", testConfigEqualSame},
		{"different functions", testConfigEqualDifferent},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
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
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

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
	tests := []testCase{
		{"concurrent operations", doTestThreadSafety},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func testNilReceiverInit(t *testing.T) {
	var s *Set[int, int, testItem]
	cfg := testConfig()

	if err := cfg.Init(s); err == nil {
		t.Error("Init on nil receiver should return error")
	}
}

func testNilReceiverMethods(t *testing.T) {
	var s *Set[int, int, testItem]

	if s.Contains(1) {
		t.Error("nil set Contains should return false")
	}

	if _, err := s.Get(1); err == nil || !core.IsError(err, core.ErrNilReceiver) {
		t.Errorf("nil set Get should return ErrNilReceiver, got: %v", err)
	}

	if _, err := s.Push(testItem{ID: 1}); err == nil || !core.IsError(err, core.ErrNilReceiver) {
		t.Errorf("nil set Push should return ErrNilReceiver, got: %v", err)
	}
}

func TestNilReceiver(t *testing.T) {
	tests := []testCase{
		{"init nil receiver", testNilReceiverInit},
		{"methods on nil", testNilReceiverMethods},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}

func doTestEmptySet(t *testing.T) {
	cfg := testConfig()
	s, err := cfg.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Test operations on empty set
	values := s.Values()
	if len(values) != 0 {
		t.Errorf("Empty set Values() should return empty slice, got %d items", len(values))
	}

	count := 0
	s.ForEach(func(testItem) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Empty set ForEach should not iterate, got %d iterations", count)
	}

	if s.Contains(1) {
		t.Error("Empty set Contains should return false")
	}

	// Test Pop on empty set
	if _, err := s.Pop(1); err == nil {
		t.Error("Pop on empty set should return error")
	}

	// Test Get on empty set
	if _, err := s.Get(1); err == nil {
		t.Error("Get on empty set should return error")
	}
}

func TestEmptySet(t *testing.T) {
	tests := []testCase{
		{"empty set operations", doTestEmptySet},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
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
	for _, item := range items {
		if _, err := s.Push(item); err != nil {
			t.Errorf("Push(%d) failed: %v", item.ID, err)
		}
	}
}

func verifyGetItems(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	for _, item := range items {
		if v, err := s.Get(item.ID); err != nil {
			t.Errorf("Get(%d) failed: %v", item.ID, err)
		} else if v.ID != item.ID {
			t.Errorf("Get(%d) returned wrong item: %+v", item.ID, v)
		}
	}
}

func verifyContainsItems(t *testing.T, s *Set[int, int, testItem], items []testItem) {
	for _, item := range items {
		if !s.Contains(item.ID) {
			t.Errorf("Contains(%d) should return true", item.ID)
		}
	}
}

func doTestHashCollisions(t *testing.T) {
	cfg := testConfig() // Uses hash function that returns k % 10
	s, err := cfg.New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	items := makeHashCollisionItems()
	verifyPushItems(t, s, items)
	verifyGetItems(t, s, items)
	verifyContainsItems(t, s, items)
}

func TestHashCollisions(t *testing.T) {
	tests := []testCase{
		{"hash collisions", doTestHashCollisions},
	}

	for _, tc := range tests {
		t.Run(tc.name, tc.test)
	}
}
