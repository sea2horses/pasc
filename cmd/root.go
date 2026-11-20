package cmd

import (
	"os"

	"pascc/internal/i18n"

	"github.com/spf13/cobra"
)

const version = "Indev (0.1.0)"

var (
	rootCmd = &cobra.Command{
		Use: "pascc",
		// Short, Long and flag help are filled in by localize().
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	i18n.Set(resolveLanguage(os.Args[1:]))
	localize()
	return rootCmd.Execute()
}

// TODO: add config files!
func init() {
}
