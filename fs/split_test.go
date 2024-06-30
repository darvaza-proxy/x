package fs

import "testing"

func TestSplit(t *testing.T) {
	tests := []struct {
		path string
		dir  string
		file string
	}{
		{"a", ".", "a"},
		{"a/b", "a", "b"},
		{"a/b/c", "a/b", "c"},
		{"/a", "", "a"},
		{"/a//./b/c", "/a/b", "c"},
		{"/a/../b/c", "/b", "c"},
		{"aa/bb/cc", "aa/bb", "cc"},
	}

	for i, tc := range tests {
		dir, file := Split(tc.path)

		if dir == tc.dir && file == tc.file {
			t.Logf("[%v/%v] Split(%q) → %q %q",
				i, len(tests), tc.path,
				dir, file)
		} else {
			t.Errorf("[%v/%v] ERROR: Split(%q) → %q %q (expected %q %q)",
				i, len(tests), tc.path,
				dir, file,
				tc.dir, tc.file)
		}
	}
}
