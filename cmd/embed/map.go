package main

import (
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/spf13/pflag"

	"darvaza.org/core"
)

// AssetsMap ...
type AssetsMap struct {
	fsys    fs.FS
	mode    Mode
	verbose bool

	patterns []string
	prefixes []string
}

// WriteTo ...
func (*AssetsMap) WriteTo(_ io.Writer) (int64, error) {
	return 0, core.ErrNotImplemented
}

// AddPatterns ...
func (m *AssetsMap) AddPatterns(patterns ...string) error {
	for _, s := range patterns {
		if !core.SliceContains(m.patterns, s) {
			if err := m.addPattern(s); err != nil {
				return err
			}

			m.patterns = append(m.patterns, s)
		}
	}
	return nil
}

func (m *AssetsMap) addPattern(s string) error {
	matches, err := fs.Glob(m.fsys, s)
	if err != nil {
		return core.Wrap(err, "%q: invalid pattern", s)
	}

	return m.AddEntries(matches...)
}

// AddEntries ...
func (*AssetsMap) AddEntries(...string) error {
	return core.ErrNotImplemented
}

// AddPrefixes ...
func (m *AssetsMap) AddPrefixes(prefixes ...string) error {
	for _, s := range prefixes {
		prefix := path.Clean(s)
		switch {
		case path.IsAbs(prefix), prefix == ".", strings.HasPrefix(prefix, "../"):
			return core.Wrap(core.ErrInvalid, "%q: invalid prefix", s)
		}

		m.prefixes = append(m.prefixes, prefix)
	}
	return nil
}

// New ...
func New(mode Mode, verbose bool) (*AssetsMap, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	m := &AssetsMap{
		fsys:    os.DirFS(cwd),
		mode:    mode,
		verbose: verbose,
	}

	return m, nil
}

// NewFromArgs ...
func NewFromArgs(flags *pflag.FlagSet, args []string) (*AssetsMap, error) {
	var mode = copyMode

	verbose, _ := flags.GetBool(verboseFlag)

	if p, ok := flags.Lookup(modeFlag).Value.(*Mode); ok && p != nil {
		mode = *p
	}

	m, err := New(mode, verbose)
	if err != nil {
		return nil, err
	}

	prefixes, _ := flags.GetStringArray(prefixFlag)
	if err := m.AddPrefixes(prefixes...); err != nil {
		return nil, err
	}

	if err := m.AddPatterns(args...); err != nil {
		return nil, err
	}

	return m, nil
}

const (
	prefixFlag       = "prefix"
	prefixShortFlag  = "p"
	verboseFlag      = "verbose"
	verboseShortFlag = "v"
	nameFlag         = "name"
	nameShortFlag    = "n"
)

func init() {
	flags := rootCmd.Flags()
	flags.BoolP(verboseFlag, verboseShortFlag, false, "enable verbose mode")
	flags.StringArrayP(prefixFlag, prefixShortFlag, []string{}, "prefixes to remove from embedded files")
	flags.StringP(nameFlag, nameShortFlag, "", "variable name where to store the files")
}
