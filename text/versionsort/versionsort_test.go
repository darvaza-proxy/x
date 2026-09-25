package versionsort_test

// cspell:words vcsa wlan FFFD

import (
	"slices"
	"testing"

	"darvaza.org/core"

	"darvaza.org/x/text/versionsort"
)

// Compile-time verification that test case types implement TestCase
var _ core.TestCase = compareTestCase{}

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

// TestCompareSort pins the order of a mixed set of device and network
// interface names.
func TestCompareSort(t *testing.T) {
	names := core.S(
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
	want := core.S(
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

	slices.SortFunc(names, versionsort.Compare)
	core.AssertSliceEqual(t, want, names, "names")
}
