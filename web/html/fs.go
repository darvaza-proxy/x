package html

import (
	"html/template"
	"io/fs"

	"darvaza.org/core"
)

// ParseFS parses file on the given [fs.FS], optionally filtered
// by the given patterns.
// As opposed to the standard [template.ParseFS], templates will be
// names using the full path structure, not only the file name.
func ParseFS(fSys fs.FS, patterns ...string) (*template.Template, error) {
	files, err := GlobFS(fSys, patterns...)
	if err != nil {
		return nil, err
	}

	return parseFilesFS(fSys, files)
}

func parseFilesFS(fSys fs.FS, files []string) (*template.Template, error) {
	var t *template.Template

	for _, name := range files {
		b, err := fs.ReadFile(fSys, name)
		if err != nil {
			return nil, err
		}

		if t == nil {
			t = template.New(name)
		} else {
			t = t.New(name)
		}

		t, err = t.Parse(string(b))
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// GlobFS returns a list of files matching content. If no patterns
// are provided, all files will be listed.
// If multiple patterns match the same file, they'll only be listed
// once.
func GlobFS(fSys fs.FS, patterns ...string) ([]string, error) {
	var files []string

	if len(patterns) == 0 {
		return listAllGlobFS(fSys)
	}

	for _, pat := range patterns {
		s, err := fs.Glob(fSys, pat)
		if err != nil {
			return nil, err
		}
		files = append(files, s...)
	}

	return core.SliceUnique(files), nil
}

func listAllGlobFS(fSys fs.FS) ([]string, error) {
	var out []string

	err := fs.WalkDir(fSys, ".", func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			out = append(out, path)
		}
		return nil
	})

	return out, err
}
