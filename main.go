package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"pascc/cmd"
	"pascc/internal/analyzer"
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/generator"
	"pascc/internal/lexer"
	"pascc/internal/parser"
	"path/filepath"
	"runtime"
)

var (
	debug    bool
	filename string
)

const (
	OUTPUT      = "build"
	CG_FILENAME = "out.go"
)

//go:embed runtime.go
var _runtime []byte

type ProgramErrors string

func (pe ProgramErrors) Error() string {
	return string(pe)
}

const (
	ErrNoFilename = ProgramErrors("expected filename to compile with -file flag")
)

func main() {
	/* Execute cobra's root command */
	cmd.Execute()
	//	if err := run(); err != nil {
	//		fmt.Fprintln(os.Stderr, err)
	//		os.Exit(1)
	//	}
	//
	// diagnostics.Dbg("Program exited successfuly")
}

func run() error {
	flag.BoolVar(&diagnostics.Debug, "debug", false, "Whether to print debug information about the compiler")
	flag.StringVar(&filename, "file", "", "File to compile (*.psc)")

	flag.Parse()

	diagnostics.Dbg("Getting file source code...")
	if filename == "" {
		return ErrNoFilename
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	diagnostics.Dbg("Tokenizing...")

	src := []rune(string(bytes))
	lexer := lexer.NewLexer(src)

	tokens, err := lexer.Tokenize()
	ds := lexer.Diagnostics()

	for i, d := range ds {
		err := diagnostics.PrintFileDiagnostic(filename, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "<err #%d could not be shown (%s)>", i+1, err.Error())
		}
	}

	e, w, _ := diagnostics.Summary(ds)
	if e > 0 || w > 0 {
		fmt.Fprintf(os.Stderr, "(%d errors, %d warnings)\n", e, w)
	}

	if e > 0 {
		return diagnostics.SummaryAsError(ds)
	}

	if err != nil {
		return err
	}

	diagnostics.Dbg("Tokens:")
	for i, token := range tokens {
		diagnostics.Dbgfmt("#%d: %+v\n", i, token)
	}

	diagnostics.Dbg("AST Parsing...")
	parser := parser.Parser{Tokens: tokens}
	ast_tree, ds := parser.Parse()

	for i, d := range ds {
		err := diagnostics.PrintFileDiagnostic(filename, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "<err #%d could not be shown (%s)>", i+1, err.Error())
		}
	}

	e, w, _ = diagnostics.Summary(ds)
	if e > 0 || w > 0 {
		fmt.Fprintf(os.Stderr, "(%d errors, %d warnings)\n", e, w)
	}

	if e > 0 {
		return diagnostics.SummaryAsError(ds)
	}

	diagnostics.Dbg(ast_tree.String())

	diagnostics.Dbg("Analyzing...")
	analyzer := analyzer.New()
	mf, ok := ast_tree.(*ast.MainFunction)
	if !ok {
		return errors.New("main function not found")
	}

	tast_tree, ds := analyzer.Analyze(mf)

	for i, d := range ds {
		err := diagnostics.PrintFileDiagnostic(filename, d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "<err #%d could not be shown (%s)>", i+1, err.Error())
		}
	}

	e, w, _ = diagnostics.Summary(ds)
	if e > 0 || w > 0 {
		fmt.Fprintf(os.Stderr, "(%d errors, %d warnings)\n", e, w)
	}

	if e > 0 {
		return diagnostics.SummaryAsError(ds)
	}

	// Print tast
	diagnostics.Dbg(tast.Dump(tast_tree))

	diagnostics.Dbg("Generating code...")
	codegen := generator.CodeGenerator{}

	code_bytes := codegen.Generate(tast_tree)
	var data []byte = append(_runtime, code_bytes...)

	diagnostics.Dbg("Generated data\n", string(data))

	// Code path
	code_path := filepath.Join(OUTPUT, CG_FILENAME)
	executable_name := "out"
	if runtime.GOOS == "windows" {
		executable_name += ".exe"
	}
	out_path := filepath.Join(OUTPUT, executable_name)

	if err = os.MkdirAll(OUTPUT, 0755); err != nil {
		return err
	}

	err = os.WriteFile(code_path, data, 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", out_path, code_path)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return err
	}

	run := exec.Command(out_path)
	run.Stderr = os.Stderr
	run.Stdout = os.Stdout
	run.Stdin = os.Stdin

	err = run.Run()
	if err != nil {
		return err
	}

	return nil
}
