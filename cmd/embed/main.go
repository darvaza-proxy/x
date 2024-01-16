// Package main provides a tool to embed files in Go
// sources.
package main

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "embed [flags] <patterns...>",
	Short: "Embeds static assets in go files",
	Args:  cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		var out io.Writer

		flags := cmd.Flags()

		m, err := NewFromArgs(flags, args)
		if err != nil {
			return err
		}

		if s, _ := flags.GetString(outputFlag); s != "" {
			f, err := os.Create(s)
			if err != nil {
				return err
			}
			defer f.Close()

			out = f
		} else {
			out = os.Stdout
		}

		_, err = m.WriteTo(out)
		return err
	},
}

func main() {
	err := rootCmd.Execute()
	switch err {
	case nil:
		// success
	case pflag.ErrHelp:
		// help exit
		os.Exit(1)
	default:
		// other errors
		log.Fatal(err)
	}
}

const (
	outputFlag      = "output"
	outputShortFlag = "o"
)

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringP(outputFlag, outputShortFlag, "", "name of the output file")
}
