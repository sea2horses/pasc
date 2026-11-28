package cmd

import (
	"errors"
	"os"
	"strings"

	"pascc/internal/i18n"

	"github.com/spf13/cobra"
)

// LangFlag holds --lang. The language is resolved before cobra runs (see
// Execute) so that --help is rendered in it too.
var LangFlag string

// envLanguages are checked in order when --lang is absent: an explicit
// pascc override first, then the POSIX locale variables by precedence.
var envLanguages = []string{"PASCC_LANG", "LC_ALL", "LC_MESSAGES", "LANG"}

func init() {
	rootCmd.PersistentFlags().StringVar(&LangFlag, "lang", "", "")
	rootCmd.PersistentPreRunE = func(*cobra.Command, []string) error {
		if LangFlag == "" {
			return nil
		}
		if _, ok := i18n.Parse(LangFlag); !ok {
			return errors.New(i18n.T(i18n.InvalidLang, LangFlag))
		}
		return nil
	}
}

// resolveLanguage picks the message language from --lang, then the
// environment; unknown values fall back to Spanish.
func resolveLanguage(args []string) i18n.Lang {
	if value, ok := langArg(args); ok {
		lang, _ := i18n.Parse(value)
		return lang
	}
	for _, name := range envLanguages {
		if value := os.Getenv(name); value != "" {
			lang, _ := i18n.Parse(value)
			return lang
		}
	}
	return i18n.ES
}

// langArg finds --lang=x or --lang x ahead of cobra's own parsing.
func langArg(args []string) (string, bool) {
	for i, arg := range args {
		switch {
		case arg == "--":
			return "", false
		case strings.HasPrefix(arg, "--lang="):
			return strings.TrimPrefix(arg, "--lang="), true
		case arg == "--lang" && i+1 < len(args):
			return args[i+1], true
		}
	}
	return "", false
}

var commandTexts = map[*cobra.Command]struct{ short, long i18n.Msg }{
	rootCmd:      {i18n.CmdRootShort, i18n.CmdRootLong},
	buildCmd:     {i18n.CmdBuildShort, i18n.CmdBuildLong},
	runCmd:       {i18n.CmdRunShort, i18n.CmdRunLong},
	fmtCmd:       {i18n.CmdFmtShort, i18n.CmdFmtLong},
	highlightCmd: {i18n.CmdHighlightShort, i18n.CmdHighlightLong},
}

var flagTexts = map[string]i18n.Msg{
	"debug":      i18n.FlagDebug,
	"fast":       i18n.FlagFast,
	"temp":       i18n.FlagTemp,
	"index-base": i18n.FlagIndexBase,
	"output":     i18n.FlagOutput,
	"target":     i18n.FlagTarget,
	"lang":       i18n.FlagLang,
	"write":      i18n.FlagWrite,
	"check":      i18n.FlagCheck,
	"indent":     i18n.FlagIndent,
	"format":     i18n.FlagFormat,
	"color":      i18n.FlagColor,
}

// localize translates everything cobra shows: descriptions, flag help, the
// usage template, the version line and the built-in help/completion commands.
func localize() {
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	rootCmd.InitDefaultVersionFlag()
	rootCmd.SetVersionTemplate(`{{with .DisplayName}}{{printf "%s " .}}{{end}}{{printf "` +
		i18n.T(i18n.VersionWord) + ` %s" .Version}}` + "\n")
	rootCmd.SetUsageTemplate(usageTemplate(rootCmd.UsageTemplate()))

	// cobra appends a literal " [flags]" to usage lines; write our own instead.
	for _, cmd := range []*cobra.Command{buildCmd, runCmd} {
		cmd.Use = cmd.Name() + " " + i18n.T(i18n.CmdFileArg) + " " + i18n.T(i18n.UsageFlagsPlaceholder)
		cmd.DisableFlagsInUseLine = true
	}
	for _, cmd := range []*cobra.Command{fmtCmd, highlightCmd} {
		cmd.Use = cmd.Name() + " " + i18n.T(i18n.CmdFilesArg) + " " + i18n.T(i18n.UsageFlagsPlaceholder)
		cmd.DisableFlagsInUseLine = true
	}
	// pflag hardcodes "(default ...)" in flag listings.
	cobra.AddTemplateFunc("localizeFlags", func(usages string) string {
		return strings.ReplaceAll(usages, " (default ", " ("+i18n.T(i18n.UsageDefault)+" ")
	})

	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if texts, ok := commandTexts[cmd]; ok {
			cmd.Short, cmd.Long = i18n.T(texts.short), i18n.T(texts.long)
		}
		switch cmd.Name() {
		case "help":
			cmd.Short = i18n.T(i18n.CmdHelpShort)
			cmd.Long = i18n.T(i18n.CmdHelpLong, rootCmd.Name())
		case "completion":
			cmd.Short = i18n.T(i18n.CmdCompletionShort)
		}
		cmd.InitDefaultHelpFlag()
		translateFlags(cmd)
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}

func translateFlags(cmd *cobra.Command) {
	if f := cmd.Flags().Lookup("help"); f != nil {
		f.Usage = i18n.T(i18n.FlagHelp, cmd.Name())
	}
	if f := cmd.Flags().Lookup("version"); f != nil {
		f.Usage = i18n.T(i18n.FlagVersion, cmd.Name())
	}
	for name, msg := range flagTexts {
		if f := cmd.Flags().Lookup(name); f != nil {
			f.Usage = i18n.T(msg)
		}
		if f := cmd.PersistentFlags().Lookup(name); f != nil {
			f.Usage = i18n.T(msg)
		}
	}
}

// usageTemplate swaps the labels of cobra's default usage template.
func usageTemplate(template string) string {
	return strings.NewReplacer(
		".FlagUsages |", ".FlagUsages | localizeFlags |",
		"Usage:", i18n.T(i18n.UsageUsage),
		"Aliases:", i18n.T(i18n.UsageAliases),
		"Examples:", i18n.T(i18n.UsageExamples),
		"Available Commands:", i18n.T(i18n.UsageCommands),
		"Additional Commands:", i18n.T(i18n.UsageAdditionalCommands),
		"Global Flags:", i18n.T(i18n.UsageGlobalFlags),
		"Flags:", i18n.T(i18n.UsageFlags),
		"Additional help topics:", i18n.T(i18n.UsageHelpTopics),
		`Use "{{.CommandPath}} [command] --help" for more information about a command.`,
		i18n.T(i18n.UsageMoreInfo, "{{.CommandPath}} "+i18n.T(i18n.UsageCommandPlaceholder)),
		"[command]", i18n.T(i18n.UsageCommandPlaceholder),
	).Replace(template)
}

// exactlyOneFile is cobra.ExactArgs(1) with a translated message.
func exactlyOneFile(_ *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New(i18n.T(i18n.ArgsExactlyOneFile, len(args)))
	}
	return nil
}
