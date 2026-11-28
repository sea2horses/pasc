package main

import (
	"os"

	"pascc/cmd"
	"pascc/internal/diagnostics"
)

func main() {
	if err := cmd.Execute(); err != nil {
		diagnostics.PrintError(err)
		os.Exit(1)
	}
}
