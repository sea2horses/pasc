package toolchain

import (
	"path/filepath"
	"testing"
)

func TestVersionLess(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"1.24.9", "1.25.0", true},
		{"1.25.0", "1.25.0", false},
		{"1.27.3", "1.25.0", false},
		{"1.25", "1.25.0", false},
		{"1.26rc1", "1.25.0", false},
		{"1.9", "1.25.0", true},
	} {
		if got := versionLess(tc.a, tc.b); got != tc.want {
			t.Errorf("versionLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestRequiredGoVersion(t *testing.T) {
	if got := requiredGoVersion([]byte("module m\n\ngo 1.25.0\n")); got != "1.25.0" {
		t.Fatalf("got %q", got)
	}
}

func TestSafeJoinRejectsEscapes(t *testing.T) {
	dest := t.TempDir()
	if _, err := safeJoin(dest, "go/bin/go"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"../evil", "go/../../evil", filepath.Join("..", "x")} {
		if _, err := safeJoin(dest, name); err == nil {
			t.Errorf("safeJoin accepted %q", name)
		}
	}
}
