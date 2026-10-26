package cmd

import (
	"pascc/internal/diagnostics"
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&Debug, "debug", false, "Prints debug information about the compiling process")
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
	diagnostics.Debug = Debug

	return toolchain.RunToolchain(
		&toolchain.ToolchainConfig{
			Filename:       filename,
			TempBuild:      true,
			Run:            true,
			Runtime:        runtime.EmbeddedRuntime,
			ExtRuntime:     runtime.EmbeddedExtRuntime,
			ExtFastRuntime: runtime.EmbeddedExtFastRuntime,
			RuntimeMod:     runtime.EmbeddedRuntimeMod,
			Vendor:         runtime.EmbeddedVendor,
		},
	)
}
