package cmd

import (
	"pascc/internal/analyzer"
	"pascc/internal/diagnostics"
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

const defaultDir = "build"

func init() {
	rootCmd.AddCommand(buildCmd)

	// Flag help text is set per language by localize().
	buildCmd.Flags().BoolVar(&Debug, "debug", false, "")
	buildCmd.Flags().BoolVar(&Fast, "fast", false, "")
	buildCmd.Flags().BoolVar(&Temp, "temp", false, "")
	buildCmd.Flags().StringVar(&IndexBaseFlag, "index-base", "auto", "")
	buildCmd.Flags().StringVarP(&OutputDir, "output", "o", defaultDir, "")
	buildCmd.Flags().StringVarP(&TargetFlag, "target", "t", "native", "")
}

var buildCmd = &cobra.Command{
	Use:  "build",
	Args: exactlyOneFile,
	RunE: compile,
}

var (
	Debug     bool
	Temp      bool
	OutputDir string
	Fast      bool

	IndexBaseFlag string
	TargetFlag    string
)

func compile(cmd *cobra.Command, args []string) error {
	filename := args[0]

	/* Get the debug option on board */
	diagnostics.Debug = Debug

	indexBase, err := analyzer.ParseIndexBase(IndexBaseFlag)
	if err != nil {
		return err
	}

	target, err := toolchain.ParseBuildTarget(TargetFlag)
	if err != nil {
		return err
	}

	return toolchain.RunToolchain(
		&toolchain.ToolchainConfig{
			Filename:       filename,
			OutputDir:      OutputDir,
			TempBuild:      Temp,
			Runtime:        runtime.EmbeddedRuntime,
			ExtRuntime:     runtime.EmbeddedExtRuntime,
			ExtFastRuntime: runtime.EmbeddedExtFastRuntime,
			RuntimeMod:     runtime.EmbeddedRuntimeMod,
			Vendor:         runtime.EmbeddedVendor,
			Fast:           Fast,
			IndexBase:      indexBase,
			Target:         target,
		},
	)
}
