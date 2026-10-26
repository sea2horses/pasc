package cmd

import (
	"pascc/internal/diagnostics"
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

const defaultDir = "build"

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolVar(&Debug, "debug", false, "Prints debug information about the compiling process")
	buildCmd.Flags().BoolVar(&Fast, "fast", false, "Uses an Enter prompt for Esperar Tecla and skips terminal dependencies")
	buildCmd.Flags().BoolVar(&Temp, "temp", false, "Builds the program temporarily and then discards the result (useful for testing)")
	buildCmd.Flags().StringVarP(&OutputDir, "output", "o", defaultDir, "Specifies the directory to build the executable onto")
}

var buildCmd = &cobra.Command{
	Use:   "build [file: *.psc]",
	Short: "Builds a .psc file into a native executable",
	Long:  "The compiler analyzes the program's code for compile-time errors and then uses Go's toolchain to convert it into a native executable",
	Args:  cobra.ExactArgs(1),
	RunE:  compile,
}

var (
	Debug     bool
	Temp      bool
	OutputDir string
	Fast      bool
)

func compile(cmd *cobra.Command, args []string) error {
	filename := args[0]

	/* Get the debug option on board */
	diagnostics.Debug = Debug

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
		},
	)
}
