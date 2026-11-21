package cmd

import "testing"

func TestParseIndent(t *testing.T) {
	for value, want := range map[string]string{"tab": "\t", "TAB": "\t", "2": "  ", "4": "    "} {
		if got, err := parseIndent(value); err != nil || got != want {
			t.Errorf("parseIndent(%q) = %q, %v; want %q", value, got, err, want)
		}
	}
	for _, value := range []string{"", "0", "9", "-1", "spaces"} {
		if _, err := parseIndent(value); err == nil {
			t.Errorf("parseIndent(%q) succeeded, want an error", value)
		}
	}
}
