package x509utils

import (
	"testing"
)

type nameAsTest struct {
	Name     string
	Expected string
	Ok       bool
}

func TestNameAsIP(t *testing.T) {
	var entries = []nameAsTest{
		{"a.b.c", "", false},
		{"0", "[0.0.0.0]", true},
		{"1.2.3.4", "[1.2.3.4]", true},
		{"1.2.3.400", "", false},
		{"::", "[::]", true},
		{"foo.example.org", "", false},
	}

	for _, entry := range entries {
		s, ok := NameAsIP(entry.Name)
		if s != entry.Expected || ok != entry.Ok {
			t.Errorf("NameAsIP(%q) -> %q, %v", entry.Name, s, ok)
		}
	}
}

func TestSanitizeName(t *testing.T) {
	var entries = []nameAsTest{
		// DNS names are case-insensitive (RFC 6125): an uppercase query
		// must fold to the lower-cased form Names stores under.
		{"WWW.EXAMPLE.COM", "www.example.com", true},
		{"www.example.com", "www.example.com", true},
		{"Mixed.Case.ORG:443", "mixed.case.org", true},
		{"1.2.3.4", "1.2.3.4", true},
		{"[2001:DB8::1]:443", "2001:db8::1", true},
		{"", "", false},
	}

	for _, entry := range entries {
		s, ok := SanitizeName(entry.Name)
		if s != entry.Expected || ok != entry.Ok {
			t.Errorf("SanitizeName(%q) -> %q, %v; want %q, %v",
				entry.Name, s, ok, entry.Expected, entry.Ok)
		}
	}
}

func TestNameAsSuffix(t *testing.T) {
	var entries = []nameAsTest{
		{"foo.example.com", ".example.com", true},
		{".example.com", "", false},
		{"a.b.c", ".b.c", true},
		{".b.c", "", false},
		{"b.c", ".c", true},
		{".c", "", false},
		{"c", "", false},
		{"", "", false},
	}

	for _, entry := range entries {
		s, ok := NameAsSuffix(entry.Name)
		if s != entry.Expected || ok != entry.Ok {
			t.Errorf("NameAsSuffix(%q) -> %q, %v", entry.Name, s, ok)
		}
	}
}
