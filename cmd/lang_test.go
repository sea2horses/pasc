package cmd

import (
	"testing"

	"pascc/internal/i18n"
)

func TestResolveLanguage(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		env  map[string]string
		want i18n.Lang
	}{
		{"default is spanish", nil, nil, i18n.ES},
		{"flag with equals", []string{"build", "--lang=en", "a.psc"}, nil, i18n.EN},
		{"flag with separate value", []string{"--lang", "en", "run", "a.psc"}, nil, i18n.EN},
		{"flag beats environment", []string{"--lang", "es"}, map[string]string{"PASCC_LANG": "en", "LANG": "en_US.UTF-8"}, i18n.ES},
		{"flag after -- is a program argument", []string{"run", "--", "--lang", "en"}, nil, i18n.ES},
		{"PASCC_LANG beats locale", nil, map[string]string{"PASCC_LANG": "en", "LANG": "es_NI.UTF-8"}, i18n.EN},
		{"LC_ALL beats LANG", nil, map[string]string{"LC_ALL": "en_GB.UTF-8", "LANG": "es_ES.UTF-8"}, i18n.EN},
		{"locale with region", nil, map[string]string{"LANG": "en_US.UTF-8"}, i18n.EN},
		{"unsupported locale falls back", nil, map[string]string{"LANG": "C"}, i18n.ES},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range envLanguages {
				t.Setenv(name, tc.env[name])
			}
			if got := resolveLanguage(tc.args); got != tc.want {
				t.Fatalf("resolveLanguage(%q) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}
