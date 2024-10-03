package x509utils

import (
	"testing"

	"darvaza.org/core"
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
