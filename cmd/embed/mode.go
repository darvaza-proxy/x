package main

import (
	"errors"

	"github.com/spf13/cobra"
)

const (
	modeFlag = "mode"

	copyMode   Mode = "copy"
	noCopyMode Mode = "nocopy"
	directMode Mode = "direct"
	wrapMode   Mode = "wrap"
)

// Mode indicates how are we going to embed the assets.
//
//   - `copy`:   indicates we will use a standard byte array.
//   - `nocopy`: indicates we will use `reflect` and `unsafe` to
//     use content directly for the .rodata section.
//   - `direct`: indicates we won't embed the file as use them
//     uncached from the filesystem.
//   - `wrap`: indicates we wrap data stored by the standard
//     `embed` system.
type Mode string

// String returns the value
func (m *Mode) String() string { return string(*m) }

// Type returns the type name
func (*Mode) Type() string { return "Mode" }

// Set changes the value
func (m *Mode) Set(s string) error {
	v := Mode(s)

	switch v {
	case copyMode, noCopyMode, directMode, wrapMode:
		*m = v
		return nil
	default:
		return errors.New(`accepted values are: "copy", "nocopy", "direct" and "wrap"`)
	}
}

func modeCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	s := []string{
		"copy\tuse standard byte array",
		"nocopy\tuse reflect and unsafe to avoid copying data",
		"direct\tpass-through to filesystem on runtime",
		"wrap\textend standard embed.FS functionality",
	}

	return s, cobra.ShellCompDirectiveDefault
}

func init() {
	rootCmd.Flags().Var(new(Mode), modeFlag,
		`embedding mode. One of "copy", "nocopy", "direct" or "wrap"`)
	_ = rootCmd.RegisterFlagCompletionFunc(modeFlag, modeCompletion)
}
