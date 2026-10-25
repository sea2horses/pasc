package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const version = "Indev (0.1.0)"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Pasc",
	Long:  "All software has versions, this is Pasc's",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
