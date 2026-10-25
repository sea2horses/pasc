package cmd

import (
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [file: *.psc]",
	Short: "Runs a .psc file, no fuss",
	Long:  "Unlike the build command, this just runs the program by building and then disposing the result",
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	filename := args[0]

	return toolchain.RunToolchain(
		&toolchain.ToolchainConfig{
			Filename:  filename,
			TempBuild: true,
			Run:       true,
			Runtime:   runtime.EmbeddedRuntime,
		},
	)
}
