package fs

import "testing"

func TestClean(t *testing.T) {
	tests := []struct {
		path string
		out  string
		ok   bool
	}{
		{"", ".", true},
		{".", ".", true},
		{"..", "..", false},
		{"../", "..", false},
		{"../.", "..", false},
		{"../a", "../a", false},
		{"a", "a", true},
		{"/", "/", false},
		{"/a", "/a", false},
		{"/a/..", "/", false},
		{"/../..//", "/../..", false},
		{"a/b", "a/b", true},
		{"a//b", "a/b", true},
		{"a/..", ".", true},
		{"a/../b", "b", true},
		{"a/b/..", "a", true},
		{"a/b/../foo", "a/foo", true},
	}

	for i, tc := range tests {
		s, ok := Clean(tc.path)

		if s == tc.out && ok == tc.ok {
			t.Logf("[%v/%v] Clean(%q) → %q %v",
				i, len(tests), tc.path,
				s, ok)
		} else {
			t.Errorf("[%v/%v] ERROR: Clean(%q) → %q %v (expected %q %v)",
				i, len(tests), tc.path,
				s, ok,
				tc.out, tc.ok)
		}
	}
}
