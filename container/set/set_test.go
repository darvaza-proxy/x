package set_test

import (
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/container/set"
)

func TestConfigNew(t *testing.T) {
	items := makeTestItems()
	s, err := testConfig().New(items...)
	core.AssertMustNoError(t, err, "new")
	core.AssertMustNotNil(t, s, "set")
	assertGetAll(t, s, items)
}

func testConfigInitBasic(t *testing.T) {
	var s set.Set[int, int, testItem]
	items := makeTestItems()
	core.AssertMustNoError(t, testConfig().Init(&s, items...), "init")
	assertGetAll(t, &s, items)
}

func testConfigInitTwice(t *testing.T) {
	var s set.Set[int, int, testItem]
	core.AssertMustNoError(t, testConfig().Init(&s), "first init")
	core.AssertErrorIs(t, testConfig().Init(&s), core.ErrInvalid, "second init")
}

func TestConfigInit(t *testing.T) {
	t.Run("basic", testConfigInitBasic)
	t.Run("twice", testConfigInitTwice)
}

func testConfigMustValid(t *testing.T) {
	s := testConfig().Must(testItem{ID: 1, Name: "one"})
	core.AssertNotNil(t, s, "set")
}

func testConfigMustPanic(t *testing.T) {
	var invalid set.Config[int, int, testItem]
	core.AssertPanic(t, func() { _ = invalid.Must(testItem{ID: 1}) }, nil, "invalid config")
}

func TestConfigMust(t *testing.T) {
	t.Run("valid", testConfigMustValid)
	t.Run("invalid panics", testConfigMustPanic)
}

func testPushNew(t *testing.T) {
	s := testConfig().Must()
	stored, err := s.Push(testItem{ID: 1, Name: "one"})
	core.AssertNoError(t, err, "push")
	core.AssertEqual(t, 1, stored.ID, "stored id")
}

func testPushDuplicate(t *testing.T) {
	s := testConfig().Must(testItem{ID: 1, Name: "one"})
	_, err := s.Push(testItem{ID: 1, Name: "duplicate"})
	core.AssertErrorIs(t, err, set.ErrExist, "duplicate")
}

func TestPush(t *testing.T) {
	t.Run("new", testPushNew)
	t.Run("duplicate", testPushDuplicate)
}

func testGetExisting(t *testing.T) {
	items := makeTestItems()
	assertGetAll(t, testConfig().Must(items...), items)
}

func testGetCollisions(t *testing.T) {
	items := makeHashCollisionItems()
	assertGetAll(t, testConfig().Must(items...), items)
}

func testGetNonExistent(t *testing.T) {
	_, err := testConfig().Must().Get(999)
	core.AssertErrorIs(t, err, set.ErrNotExist, "missing key")
}

func TestGet(t *testing.T) {
	t.Run("existing", testGetExisting)
	t.Run("collisions", testGetCollisions)
	t.Run("non-existent", testGetNonExistent)
}

var _ core.TestCase = containsTestCase{}

// containsTestCase checks Set.Contains against a shared populated set.
type containsTestCase struct {
	name string
	set  *set.Set[int, int, testItem]
	key  int
	want bool
}

func newContainsTestCase(name string, s *set.Set[int, int, testItem], key int,
	want bool) containsTestCase {
	return containsTestCase{
		name: name,
		set:  s,
		key:  key,
		want: want,
	}
}

func (tc containsTestCase) Name() string {
	return tc.name
}

func (tc containsTestCase) Test(t *testing.T) {
	t.Helper()
	core.AssertEqual(t, tc.want, tc.set.Contains(tc.key), "contains %d", tc.key)
}

func containsTestCases() []containsTestCase {
	// IDs 1 and 11 share bucket 1, so the absent sibling 21 probes a
	// populated bucket while 2 probes an empty one.
	s := testConfig().Must(testItem{ID: 1, Name: "one"}, testItem{ID: 11, Name: "eleven"})
	return []containsTestCase{
		newContainsTestCase("existing", s, 1, true),
		newContainsTestCase("collision sibling", s, 11, true),
		newContainsTestCase("absent in empty bucket", s, 2, false),
		newContainsTestCase("absent in populated bucket", s, 21, false),
	}
}

func TestContains(t *testing.T) {
	core.RunTestCases(t, containsTestCases())
}

func testPopExisting(t *testing.T) {
	s := testConfig().Must(testItem{ID: 1, Name: "one"})
	v, err := s.Pop(1)
	core.AssertNoError(t, err, "pop")
	core.AssertEqual(t, 1, v.ID, "popped id")
	core.AssertFalse(t, s.Contains(1), "removed")
}

func testPopNonExistent(t *testing.T) {
	_, err := testConfig().Must().Pop(1)
	core.AssertErrorIs(t, err, set.ErrNotExist, "missing key")
}

func TestPop(t *testing.T) {
	t.Run("existing", testPopExisting)
	t.Run("non-existent", testPopNonExistent)
}

func TestReset(t *testing.T) {
	items := makeTriple()
	s := testConfig().Must(items...)
	assertGetAll(t, s, items)

	core.AssertMustNoError(t, s.Reset(), "reset")
	for _, item := range items {
		core.AssertFalse(t, s.Contains(item.ID), "cleared %d", item.ID)
	}
}

func TestClone(t *testing.T) {
	items := makeTestItems()
	s1 := testConfig().Must(items...)

	s2 := s1.Clone()
	core.AssertMustNotNil(t, s2, "clone")
	assertGetAll(t, s2, items)

	_, _ = s1.Push(testItem{ID: 3, Name: "three"})
	core.AssertFalse(t, s2.Contains(3), "clone independent of original")
}

func testCopyFilter(t *testing.T) {
	s1 := testConfig().Must(makeTriple()...)
	s2 := s1.Copy(nil, func(v testItem) bool { return v.ID%2 == 0 })

	core.AssertFalse(t, s2.Contains(1), "odd excluded")
	core.AssertTrue(t, s2.Contains(2), "even included")
	core.AssertFalse(t, s2.Contains(3), "odd excluded")
}

func testCopyAll(t *testing.T) {
	items := makeTriple()
	s2 := testConfig().Must(items...).Copy(nil, nil)
	assertGetAll(t, s2, items)
}

func TestCopy(t *testing.T) {
	t.Run("filter", testCopyFilter)
	t.Run("all", testCopyAll)
}

func TestValues(t *testing.T) {
	s := testConfig().Must(makeTriple()...)
	assertHasIDs(t, s, 1, 2, 3)
}

func testForEachAll(t *testing.T) {
	s := testConfig().Must(makeTriple()...)
	count := 0
	s.ForEach(func(testItem) bool { count++; return true })
	core.AssertEqual(t, 3, count, "visited all")
}

func testForEachEarlyTermination(t *testing.T) {
	s := testConfig().Must(makeTriple()...)
	count := 0
	s.ForEach(func(testItem) bool { count++; return count < 2 })
	core.AssertEqual(t, 2, count, "stopped after two")
}

func TestForEach(t *testing.T) {
	t.Run("all items", testForEachAll)
	t.Run("early termination", testForEachEarlyTermination)
}

var _ core.TestCase = configValidationTestCase{}

// configValidationTestCase checks Config.Validate across missing callbacks.
type configValidationTestCase struct {
	cfg     set.Config[int, int, testItem]
	name    string
	wantErr bool
}

func newConfigValidationTestCase(name string, cfg set.Config[int, int, testItem],
	wantErr bool) configValidationTestCase {
	return configValidationTestCase{
		cfg:     cfg,
		name:    name,
		wantErr: wantErr,
	}
}

func (tc configValidationTestCase) Name() string {
	return tc.name
}

func (tc configValidationTestCase) Test(t *testing.T) {
	t.Helper()
	err := tc.cfg.Validate()
	if tc.wantErr {
		core.AssertError(t, err, "validate")
		return
	}
	core.AssertNoError(t, err, "validate")
}

func configValidationTestCases() []configValidationTestCase {
	full := testConfig()
	return []configValidationTestCase{
		newConfigValidationTestCase("missing ItemKey",
			set.Config[int, int, testItem]{Hash: full.Hash, ItemMatch: full.ItemMatch}, true),
		newConfigValidationTestCase("missing Hash",
			set.Config[int, int, testItem]{ItemKey: full.ItemKey, ItemMatch: full.ItemMatch}, true),
		newConfigValidationTestCase("missing ItemMatch",
			set.Config[int, int, testItem]{ItemKey: full.ItemKey, Hash: full.Hash}, true),
		newConfigValidationTestCase("valid", full, false),
	}
}

func TestConfigValidation(t *testing.T) {
	core.RunTestCases(t, configValidationTestCases())
}

func testConfigEqualSame(t *testing.T) {
	keyFn := func(v testItem) (int, error) { return v.ID, nil }
	hashFn := func(k int) (int, error) { return k, nil }
	matchFn := func(k int, v testItem) bool { return k == v.ID }
	cfg1 := set.Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}
	cfg2 := set.Config[int, int, testItem]{ItemKey: keyFn, Hash: hashFn, ItemMatch: matchFn}

	core.AssertTrue(t, cfg1.Equal(cfg2), "same functions")
}

func testConfigEqualDifferent(t *testing.T) {
	core.AssertFalse(t, testConfig().Equal(testConfig()), "different instances")
}

func TestConfigEqual(t *testing.T) {
	t.Run("same functions", testConfigEqualSame)
	t.Run("different functions", testConfigEqualDifferent)
}

func runConcurrentWriter(s *set.Set[int, int, testItem], done chan bool) {
	for i := range 100 {
		_, _ = s.Push(testItem{ID: i, Name: "item", Value: "value"})
	}
	done <- true
}

func runConcurrentReader(s *set.Set[int, int, testItem], done chan bool) {
	for i := range 100 {
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

func TestThreadSafety(t *testing.T) {
	s := testConfig().Must()
	done := make(chan bool)

	go runConcurrentWriter(s, done)
	go runConcurrentReader(s, done)
	go runConcurrentForEach(s, done)

	for range 3 {
		<-done
	}
	core.AssertNoPanic(t, func() { _ = s.Values() }, "consistent after concurrency")
}

func testNilReceiverInit(t *testing.T) {
	var s *set.Set[int, int, testItem]
	core.AssertErrorIs(t, testConfig().Init(s), core.ErrInvalid, "init on nil receiver")
}

func testNilReceiverMethods(t *testing.T) {
	var s *set.Set[int, int, testItem]
	core.AssertFalse(t, s.Contains(1), "contains on nil")

	_, err := s.Get(1)
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "get on nil")

	_, err = s.Push(testItem{ID: 1})
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "push on nil")
}

func TestNilReceiver(t *testing.T) {
	t.Run("init", testNilReceiverInit)
	t.Run("methods", testNilReceiverMethods)
}

func TestEmptySet(t *testing.T) {
	s := testConfig().Must()

	core.AssertEqual(t, 0, len(s.Values()), "empty values")
	core.AssertFalse(t, s.Contains(1), "empty contains")

	_, err := s.Pop(1)
	core.AssertErrorIs(t, err, set.ErrNotExist, "pop on empty")

	_, err = s.Get(1)
	core.AssertErrorIs(t, err, set.ErrNotExist, "get on empty")
}

func TestHashCollisions(t *testing.T) {
	items := makeHashCollisionItems()
	s := testConfig().Must(items...)
	assertGetAll(t, s, items)
	assertHasIDs(t, s, 1, 11, 21, 31)
}

// errKeyConfig is a valid Config whose ItemKey rejects negative IDs, used to
// drive the error paths of Push and the initial bulk insert.
func errKeyConfig() set.Config[int, int, testItem] {
	cfg := testConfig()
	cfg.ItemKey = func(v testItem) (int, error) {
		if v.ID < 0 {
			return 0, core.ErrInvalid
		}
		return v.ID, nil
	}
	return cfg
}

func TestInitItemKeyError(t *testing.T) {
	cfg := errKeyConfig()
	_, err := cfg.New(testItem{ID: -1, Name: "negative"})
	core.AssertErrorIs(t, err, core.ErrInvalid, "New with failing ItemKey")
}

func testResetNilReceiver(t *testing.T) {
	var s *set.Set[int, int, testItem]
	core.AssertErrorIs(t, s.Reset(), core.ErrNilReceiver, "reset nil receiver")
}

func testResetUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	core.AssertErrorIs(t, s.Reset(), core.ErrInvalid, "reset uninitialised")
}

func TestResetErrors(t *testing.T) {
	t.Run("nil receiver", testResetNilReceiver)
	t.Run("uninitialised", testResetUninitialised)
}

func testPopNilReceiver(t *testing.T) {
	var s *set.Set[int, int, testItem]
	_, err := s.Pop(1)
	core.AssertErrorIs(t, err, core.ErrNilReceiver, "pop nil receiver")
}

func testPopHashCollisionMiss(t *testing.T) {
	// ID 1 and 11 share hash 1; popping 11 finds the bucket but no match.
	s := testConfig().Must(testItem{ID: 1, Name: "one"})

	_, err := s.Pop(11)

	core.AssertErrorIs(t, err, set.ErrNotExist, "pop absent key in existing bucket")
	core.AssertTrue(t, s.Contains(1), "original entry retained")
}

func TestPopErrors(t *testing.T) {
	t.Run("nil receiver", testPopNilReceiver)
	t.Run("hash collision miss", testPopHashCollisionMiss)
}

func testGetUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	_, err := s.Get(1)
	core.AssertErrorIs(t, err, core.ErrNotImplemented, "get on uninitialised")
}

func testContainsUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	core.AssertFalse(t, s.Contains(1), "contains on uninitialised")
}

func testPushUninitialised(t *testing.T) {
	s := &set.Set[int, int, testItem]{}
	_, err := s.Push(testItem{ID: 1, Name: "one"})
	core.AssertErrorIs(t, err, core.ErrNotImplemented, "push on uninitialised")
}

func TestUninitialisedAccess(t *testing.T) {
	t.Run("get", testGetUninitialised)
	t.Run("contains", testContainsUninitialised)
	t.Run("push", testPushUninitialised)
}
