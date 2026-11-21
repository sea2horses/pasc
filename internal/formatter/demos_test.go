package formatter

import (
	"os"
	"path/filepath"
	"testing"

	"pascc/internal/toolchain"
)

// Every demo program must format without error, and formatting must be
// idempotent: formatting formatted code changes nothing.
func TestFormatDemos(t *testing.T) {
	files, err := filepath.Glob("../../demos/*.psc")
	if err != nil || len(files) == 0 {
		t.Fatalf("no demos found: %v", err)
	}
	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			src, _ := toolchain.DecodeSource(data)
			once, err := Format(src, Options{})
			if err != nil {
				t.Fatalf("format: %v", err)
			}
			twice, err := Format([]rune(once), Options{})
			if err != nil {
				t.Fatalf("second format: %v", err)
			}
			if once != twice {
				t.Fatalf("formatting is not idempotent:\n--- once\n%s\n--- twice\n%s", once, twice)
			}
			if dir := os.Getenv("PASC_FMT_DUMP"); dir != "" {
				_ = os.WriteFile(filepath.Join(dir, filepath.Base(file)), []byte(once), 0o644)
			}
		})
	}
}
