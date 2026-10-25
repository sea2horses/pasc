package cmd

import (
	"pascc/internal/diagnostics"
	"pascc/internal/runtime"
	"pascc/internal/toolchain"

	"github.com/spf13/cobra"
)

const defaultDir = "build"

func init() {
	rootCmd.AddCommand(compileCmd)

	compileCmd.Flags().BoolVar(&Debug, "debug", false, "Prints debug information about the compiling process")
	compileCmd.Flags().StringVarP(&OutputDir, "output", "o", defaultDir, "Specifies the directory to build the executable onto")
}

var compileCmd = &cobra.Command{
	Use:   "compile [file: *.psc]",
	Short: "Compiles a .psc file into a native executable",
	Long:  "The compiler analyzes the program's code for compile-time errors and then uses Go's toolchain to convert it into a native executable",
	Args:  cobra.ExactArgs(1),
	RunE:  compile,
}

var (
	Debug     bool
	OutputDir string
)

func compile(cmd *cobra.Command, args []string) error {
	filename := args[0]

	/* Get the debug option on board */
	diagnostics.Debug = Debug

	return toolchain.RunToolchain(
		&toolchain.ToolchainConfig{
			Filename:  filename,
			OutputDir: OutputDir,
			Runtime:   runtime.EmbeddedRuntime,
		},
	)
}
