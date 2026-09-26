package versionsort_test

// cspell:words vcsa wlan FFFD

import (
	"slices"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/text/versionsort"
)

// Compile-time verification that test case types implement TestCase
var (
	_ core.TestCase = compareTestCase{}
	_ core.TestCase = sortTestCase{}
	_ core.TestCase = sortByTestCase{}
)

// compareTestCase pins where one string falls against another, and that
// the reverse comparison answers the opposite.
type compareTestCase struct {
	name string
	a    string
	b    string
	want int
}

func (tc compareTestCase) Name() string {
	return tc.name
}

func (tc compareTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.want, versionsort.Compare(tc.a, tc.b), "compare")
	core.AssertEqual(t, -tc.want, versionsort.Compare(tc.b, tc.a), "reverse")
}

func newCompareTestCase(name, a, b string, want int) compareTestCase {
	return compareTestCase{
		name: name,
		a:    a,
		b:    b,
		want: want,
	}
}

// newCompareTestCaseBefore is a row where a comes first.
func newCompareTestCaseBefore(name, a, b string) compareTestCase {
	return newCompareTestCase(name, a, b, -1)
}

func compareTestCases() []compareTestCase {
	return []compareTestCase{
		newCompareTestCase("same string", "eth0", "eth0", 0),
		newCompareTestCase("both empty", "", "", 0),
		newCompareTestCaseBefore("empty before anything", "", "a"),
		newCompareTestCaseBefore("empty before a tilde", "", "~"),
		newCompareTestCaseBefore("number, not text", "ttyS9", "ttyS10"),
		newCompareTestCaseBefore("a leading number decides", "2", "10"),
		newCompareTestCaseBefore("a leading number before the rest", "9.0", "10.0"),
		newCompareTestCaseBefore("a shared digit is part of the number", "a19", "a100"),
		newCompareTestCaseBefore("every number, not the last alone", "enp2s0", "enp10s0"),
		newCompareTestCaseBefore("numbers wider than any integer", "n99999999999999999999", "n100000000000000000000"),
		newCompareTestCaseBefore("port on a USB hub", "enp0s20f0u2", "enp0s20f0u10"),
		newCompareTestCaseBefore("port before the hub behind it", "enp0s20f0u1", "enp0s20f0u1u2"),
		newCompareTestCaseBefore("text before a later number", "sda10", "sdb"),
		newCompareTestCaseBefore("no number before one", "tty", "tty1"),
		newCompareTestCaseBefore("no number counts as zero", "a0~", "a"),
		newCompareTestCaseBefore("end before a letter", "vcs", "vcsa"),
		newCompareTestCaseBefore("end before other characters", "vcs", "vcs-"),
		newCompareTestCaseBefore("digit before a letter", "tty1", "ttyS0"),
		newCompareTestCaseBefore("digit before other characters", "loop0", "loop-control"),
		newCompareTestCaseBefore("letter before other characters", "dma_heap", "dm-0"),
		newCompareTestCaseBefore("upper case before lower", "enP1p1s0", "enp0s20f0u1"),
		newCompareTestCaseBefore("tilde before the end", "1.0~rc1", "1.0"),
		newCompareTestCaseBefore("tilde before a digit", "a~", "a1"),
		newCompareTestCaseBefore("zero before one", "ttyS0", "ttyS01"),
		newCompareTestCaseBefore("leading zeros fall back to bytes", "ttyS01", "ttyS1"),
		newCompareTestCaseBefore("a missing number ties with zero, then bytes", "tty", "tty0"),
		newCompareTestCaseBefore("later text before leading zeros", "a1b", "a01c"),
		newCompareTestCaseBefore("Unicode letter before punctuation", "é", "-"),
		newCompareTestCaseBefore("letters by code point, not by case", "e", "É"),
		newCompareTestCaseBefore("a shared first byte is not a shared character", "é", "×"),
		newCompareTestCaseBefore("a shared non-ASCII character", "éa", "éb"),
		newCompareTestCaseBefore("a shared invalid byte", "a\x80a", "a\x80b"),
		newCompareTestCaseBefore("no normalisation", "e\u0301", "\u00e9"),
		newCompareTestCaseBefore("only ASCII digits are numbers", "x10", "x٩"),
		newCompareTestCaseBefore("invalid UTF-8 after every character", "a\U0010FFFF", "a\x80"),
		newCompareTestCaseBefore("U+FFFD is a character, not an invalid byte", "a\uFFFD", "a\x80"),
		newCompareTestCaseBefore("invalid bytes by value", "a\x80", "a\xff"),
		newCompareTestCaseBefore("a leading dot is an ordinary character", "b", ".a"),
		newCompareTestCaseBefore("a file suffix is ordinary text", "a-1.tar.gz", "a.tar.gz"),
	}
}

func TestCompare(t *testing.T) {
	core.RunTestCases(t, compareTestCases())
}

// TestCompareNamedType pins that a type over string compares without
// conversion.
func TestCompareNamedType(t *testing.T) {
	type name string

	core.AssertEqual(t, -1, versionsort.Compare[name]("eth2", "eth10"), "compare")
}

// deviceNames returns a mixed set of device and network interface names,
// out of order.
func deviceNames() []string {
	return core.S(
		"wlan0",
		"eth10",
		"enx00e04c680001",
		"eth2",
		"enp0s20f0u10",
		"enp0s20f0u2",
		"enp0s20f0u1u2",
		"enp0s20f0u1",
		"enp10s0",
		"enp2s0",
		"enP1p1s0",
		"lo",
		"usb0",
		"eth0",
		"ttyS9",
		"ttyS10",
		"ttyS01",
		"ttyS1",
		"ttyS0",
		"tty",
		"tty1",
		"tty~",
		"vcs",
		"vcsa",
		"loop0",
		"loop-control",
		"dma_heap",
		"dm-0",
		"sda10",
		"sdb",
	)
}

// sortedDeviceNames returns deviceNames in version order.
func sortedDeviceNames() []string {
	return core.S(
		"dma_heap",
		"dm-0",
		"enP1p1s0",
		"enp0s20f0u1",
		"enp0s20f0u1u2",
		"enp0s20f0u2",
		"enp0s20f0u10",
		"enp2s0",
		"enp10s0",
		"enx00e04c680001",
		"eth0",
		"eth2",
		"eth10",
		"lo",
		"loop0",
		"loop-control",
		"sda10",
		"sdb",
		"tty~",
		"tty",
		"tty1",
		"ttyS0",
		"ttyS01",
		"ttyS1",
		"ttyS9",
		"ttyS10",
		"usb0",
		"vcs",
		"vcsa",
		"wlan0",
	)
}

// sortTestCase pins the order Sort leaves a slice in, and that it gets
// there without panicking.
type sortTestCase struct {
	name  string
	input []string
	want  []string
}

func (tc sortTestCase) Name() string {
	return tc.name
}

func (tc sortTestCase) Test(t *testing.T) {
	t.Helper()

	got := slices.Clone(tc.input)
	core.AssertMustNoPanic(t, func() { versionsort.Sort(got) }, "sort")
	core.AssertSliceEqual(t, tc.want, got, "sorted")
}

func newSortTestCase(name string, input, want []string) sortTestCase {
	return sortTestCase{
		name:  name,
		input: input,
		want:  want,
	}
}

func sortTestCases() []sortTestCase {
	return []sortTestCase{
		newSortTestCase("nil slice", nil, nil),
		newSortTestCase("empty slice", core.S[string](), core.S[string]()),
		newSortTestCase("one element", core.S("eth0"), core.S("eth0")),
		newSortTestCase("already sorted", core.S("eth2", "eth10"), core.S("eth2", "eth10")),
		newSortTestCase("reversed", core.S("eth10", "eth2", "eth0"), core.S("eth0", "eth2", "eth10")),
		newSortTestCase("duplicates",
			core.S("eth10", "eth2", "eth10", "eth2"),
			core.S("eth2", "eth2", "eth10", "eth10")),
		newSortTestCase("all the same", core.S("eth0", "eth0", "eth0"), core.S("eth0", "eth0", "eth0")),
		newSortTestCase("empty strings first", core.S("b", "", "a", ""), core.S("", "", "a", "b")),
		newSortTestCase("leading zeros by bytes", core.S("ttyS1", "ttyS01"), core.S("ttyS01", "ttyS1")),
		newSortTestCase("device names", deviceNames(), sortedDeviceNames()),
	}
}

func TestSort(t *testing.T) {
	core.RunTestCases(t, sortTestCases())
}

// TestSortNamedType pins that a slice of a type over string sorts
// without conversion.
func TestSortNamedType(t *testing.T) {
	type name string

	names := core.S[name]("eth10", "eth2")
	versionsort.Sort(names)
	core.AssertSliceEqual(t, core.S[name]("eth2", "eth10"), names, "names")
}

// device is an element SortBy sorts by name, its id recording where it
// started.
type device struct {
	name string
	id   int
}

func newDevice(name string, id int) device {
	return device{name: name, id: id}
}

// deviceName is the key SortBy sorts devices by.
func deviceName(d device) string {
	return d.name
}

// interleavedDevices returns 32 devices cycling through four names out of
// order. slices.SortFunc insertion-sorts only short slices, and on one
// this long it does not keep equal names in their order.
func interleavedDevices() []device {
	const count = 32
	names := core.S("eth10", "eth2", "eth1", "eth0")

	devices := make([]device, 0, count)
	for id := range count {
		devices = append(devices, newDevice(names[id%len(names)], id))
	}
	return devices
}

// sortedInterleavedDevices returns interleavedDevices in version order,
// devices with the same name in their order.
func sortedInterleavedDevices() []device {
	devices := interleavedDevices()

	out := make([]device, 0, len(devices))
	for _, name := range core.S("eth0", "eth1", "eth2", "eth10") {
		for _, d := range devices {
			if d.name == name {
				out = append(out, d)
			}
		}
	}
	return out
}

// TestInterleavedDevicesUnstable pins that an unstable sort does not keep
// interleavedDevices' equal names in their order, so the row sorting them
// tells SortBy's stable sort from an unstable one.
func TestInterleavedDevicesUnstable(t *testing.T) {
	devices := interleavedDevices()
	slices.SortFunc(devices, func(a, b device) int {
		return versionsort.Compare(a.name, b.name)
	})

	core.AssertFalse(t, slices.Equal(sortedInterleavedDevices(), devices), "order kept")
}

// sortByTestCase pins the order SortBy leaves a slice in, and that it
// gets there without panicking.
type sortByTestCase struct {
	fn    func(device) string
	name  string
	input []device
	want  []device
}

func (tc sortByTestCase) Name() string {
	return tc.name
}

func (tc sortByTestCase) Test(t *testing.T) {
	t.Helper()

	got := slices.Clone(tc.input)
	core.AssertMustNoPanic(t, func() { versionsort.SortBy(got, tc.fn) }, "sort")
	core.AssertSliceEqual(t, tc.want, got, "sorted")
}

func newSortByTestCase(name string, input []device, fn func(device) string,
	want []device) sortByTestCase {
	return sortByTestCase{
		fn:    fn,
		name:  name,
		input: input,
		want:  want,
	}
}

func sortByTestCases() []sortByTestCase {
	return []sortByTestCase{
		newSortByTestCase("nil slice", nil, deviceName, nil),
		newSortByTestCase("empty slice", core.S[device](), deviceName, core.S[device]()),
		newSortByTestCase("one element",
			core.S(newDevice("eth0", 0)), deviceName,
			core.S(newDevice("eth0", 0))),
		newSortByTestCase("duplicates keep their order",
			core.S(newDevice("eth10", 0), newDevice("eth2", 1), newDevice("eth10", 2), newDevice("eth2", 3)),
			deviceName,
			core.S(newDevice("eth2", 1), newDevice("eth2", 3), newDevice("eth10", 0), newDevice("eth10", 2))),
		newSortByTestCase("all the same keep their order",
			core.S(newDevice("eth0", 0), newDevice("eth0", 1), newDevice("eth0", 2)),
			deviceName,
			core.S(newDevice("eth0", 0), newDevice("eth0", 1), newDevice("eth0", 2))),
		newSortByTestCase("empty strings first",
			core.S(newDevice("eth0", 0), newDevice("", 1)), deviceName,
			core.S(newDevice("", 1), newDevice("eth0", 0))),
		newSortByTestCase("stable beyond insertion sort",
			interleavedDevices(), deviceName, sortedInterleavedDevices()),
		newSortByTestCase("nil function, nil slice", nil, nil, nil),
		newSortByTestCase("nil function, empty slice", core.S[device](), nil, core.S[device]()),
		newSortByTestCase("nil function leaves the order",
			core.S(newDevice("eth10", 0), newDevice("eth2", 1)), nil,
			core.S(newDevice("eth10", 0), newDevice("eth2", 1))),
	}
}

func TestSortBy(t *testing.T) {
	core.RunTestCases(t, sortByTestCases())
}

// TestSortByCallsOnce pins that SortBy asks for each element's string
// once.
func TestSortByCallsOnce(t *testing.T) {
	devices := interleavedDevices()
	calls := make(map[int]int, len(devices))

	versionsort.SortBy(devices, func(d device) string {
		calls[d.id]++
		return d.name
	})

	core.AssertEqual(t, len(devices), len(calls), "elements asked")
	for id, n := range calls {
		core.AssertEqual(t, 1, n, "calls for device %d", id)
	}
}

// TestSortByNamedType pins that SortBy takes a function giving a type over
// string.
func TestSortByNamedType(t *testing.T) {
	type name string

	devices := core.S(newDevice("eth10", 0), newDevice("eth2", 1))
	versionsort.SortBy(devices, func(d device) name { return name(d.name) })
	core.AssertSliceEqual(t, core.S(newDevice("eth2", 1), newDevice("eth10", 0)), devices, "devices")
}
