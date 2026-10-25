package cmd

import "github.com/spf13/cobra"

const version = "Indev (0.1.0)"

var (
	rootCmd = &cobra.Command{
		Use:           "pascc",
		Short:         "Compiler & analyzer tool for the PSeInt language",
		Long:          "Pasc analyzes PSeInt programs and compiles them into native executables using the Go toolchain.",
		Version:       version,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

// TODO: add config files!
func init() {
}
