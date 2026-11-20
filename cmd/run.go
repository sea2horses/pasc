package cmd

import (
	"pascc/internal/analyzer"
	"pascc/internal/diagnostics"
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&Debug, "debug", false, "")
	runCmd.Flags().StringVar(&IndexBaseFlag, "index-base", "auto", "")
}

var runCmd = &cobra.Command{
	Use:  "run",
	Args: exactlyOneFile,
	RunE: run,
}

func run(cmd *cobra.Command, args []string) error {
	filename := args[0]
	diagnostics.Debug = Debug

	indexBase, err := analyzer.ParseIndexBase(IndexBaseFlag)
	if err != nil {
		return err
	}

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
			IndexBase:      indexBase,
		},
	)
}
