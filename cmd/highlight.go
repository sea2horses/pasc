package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"pascc/internal/highlight"
	"pascc/internal/i18n"
	"pascc/internal/toolchain"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(highlightCmd)
	highlightCmd.Flags().StringVar(&HighlightFormat, "format", "ansi", "")
	highlightCmd.Flags().StringVar(&HighlightColor, "color", "auto", "")
}

var highlightCmd = &cobra.Command{
	Use:     "highlight",
	Aliases: []string{"hl"},
	Args:    atLeastOneFile,
	RunE:    highlightFiles,
}

var (
	HighlightFormat string
	HighlightColor  string
)

func highlightFiles(cmd *cobra.Command, args []string) error {
	format := strings.ToLower(HighlightFormat)
	if format != "ansi" && format != "html" {
		return errors.New(i18n.T(i18n.InvalidFormat, HighlightFormat))
	}
	colors, err := useColors(HighlightColor)
	if err != nil {
		return err
	}

	paths, err := collectSources(args)
	if err != nil {
		return err
	}

	var files []highlight.File
	var failures []error
	for _, path := range paths {
		src, err := readSource(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		files = append(files, highlight.File{Name: path, Segments: highlight.Segments(src)})
	}

	out := cmd.OutOrStdout()
	if out == os.Stdout {
		out = color.Output /* Translates escape sequences on older Windows consoles */
	}
	if format == "html" {
		if len(files) > 0 {
			io.WriteString(out, highlight.HTMLDocument(files))
		}
	} else {
		for _, file := range files {
			if colors {
				io.WriteString(out, highlight.ANSI(file.Segments))
			} else {
				io.WriteString(out, highlight.Text(file.Segments))
			}
		}
	}
	return errors.Join(failures...)
}

// useColors resolves --color. "auto" follows fatih/color, which disables
// colors when stdout is not a terminal or NO_COLOR is set.
func useColors(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "auto":
		return !color.NoColor, nil
	case "always":
		return true, nil
	case "never":
		return false, nil
	}
	return false, errors.New(i18n.T(i18n.InvalidColor, value))
}

func readSource(path string) ([]rune, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.New(i18n.T(i18n.FileNotFound, path))
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T(i18n.OpenFileFailed, path), err)
	}
	src, _ := toolchain.DecodeSource(data)
	return src, nil
}
