package cmd

import "github.com/spf13/cobra"

const version = "Indev (0.1.0)"

var (
	rootCmd = &cobra.Command{
		Use:     "pascc",
		Short:   "Compiler & analyzer tool for the PSeInt language",
		Long:    "Pasc is a tool for both analyzing PSeInt code and compiling it to native executables bootstrapping from the go language.",
		Version: version,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

// TODO: add config files!
func init() {
}
