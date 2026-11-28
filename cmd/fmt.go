package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pascc/internal/diagnostics"
	"pascc/internal/formatter"
	"pascc/internal/i18n"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().StringVar(&DiagnosticFormatFlag, "diagnostic-format", "text", "")
	fmtCmd.Flags().BoolVarP(&FmtWrite, "write", "w", false, "")
	fmtCmd.Flags().BoolVarP(&FmtCheck, "check", "c", false, "")
	fmtCmd.Flags().StringVar(&FmtIndent, "indent", "tab", "")
}

var fmtCmd = &cobra.Command{
	Use:  "fmt",
	Args: atLeastOneFile,
	RunE: formatFiles,
}

var (
	FmtWrite  bool
	FmtCheck  bool
	FmtIndent string
)

func formatFiles(cmd *cobra.Command, args []string) error {
	if err := applyDiagnosticFormat(); err != nil {
		return err
	}
	indent, err := parseIndent(FmtIndent)
	if err != nil {
		return err
	}

	files, err := collectSources(args)
	if err != nil {
		return err
	}

	var failures []error
	unformatted := 0
	for _, file := range files {
		changed, err := formatFile(file, formatter.Options{Indent: indent}, cmd)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if changed && FmtCheck {
			fmt.Fprintln(cmd.OutOrStdout(), file)
			unformatted++
		}
	}

	if unformatted == 1 {
		failures = append(failures, errors.New(i18n.T(i18n.FmtNotFormatted, unformatted)))
	} else if unformatted > 1 {
		failures = append(failures, errors.New(i18n.T(i18n.FmtNotFormattedMany, unformatted)))
	}
	return errors.Join(failures...)
}

// formatFile formats one file according to the flags and reports whether
// its formatting differed from the canonical layout.
func formatFile(file string, opts formatter.Options, cmd *cobra.Command) (bool, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, fs.ErrNotExist) {
		return false, errors.New(i18n.T(i18n.FileNotFound, file))
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w", i18n.T(i18n.OpenFileFailed, file), err)
	}

	src, enc := toolchain.DecodeSource(data)
	out, err := formatter.Format(src, opts)
	var syntax *formatter.SyntaxError
	switch {
	case errors.As(err, &syntax):
		diagnostics.PrintDiagnostics(syntax.Diagnostics, file)
		return false, errors.New(i18n.T(i18n.FmtSyntaxError, file))
	case errors.Is(err, formatter.ErrUnstable):
		return false, errors.New(i18n.T(i18n.FmtUnstable, file))
	case err != nil:
		return false, err
	}

	changed := out != string(src)
	switch {
	case FmtCheck: // --check never modifies anything, even with --write
	case FmtWrite:
		if changed {
			if err := writeKeepingMode(file, toolchain.EncodeSource(out, enc)); err != nil {
				return false, fmt.Errorf("%s: %w", i18n.T(i18n.WriteFileFailed, file), err)
			}
		}
	default:
		fmt.Fprint(cmd.OutOrStdout(), out)
	}
	return changed, nil
}

func writeKeepingMode(file string, data []byte) error {
	info, err := os.Stat(file)
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, info.Mode().Perm())
}

// collectSources expands directories into the .psc files they contain.
// Files named explicitly are used as given, whatever their extension.
func collectSources(args []string) ([]string, error) {
	var files []string
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil || !info.IsDir() {
			files = append(files, arg)
			continue
		}
		err = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".psc") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func parseIndent(value string) (string, error) {
	if strings.EqualFold(value, "tab") {
		return "\t", nil
	}
	if n, err := strconv.Atoi(value); err == nil && n >= 1 && n <= 8 {
		return strings.Repeat(" ", n), nil
	}
	return "", errors.New(i18n.T(i18n.InvalidIndent, value))
}

func atLeastOneFile(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New(i18n.T(i18n.ArgsAtLeastOneFile))
	}
	return nil
}
