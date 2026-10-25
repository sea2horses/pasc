package toolchain

import (
	"errors"
	"os"
	"os/exec"
	"pascc/internal/analyzer"
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/generator"
	"pascc/internal/lexer"
	"pascc/internal/parser"
	"path/filepath"
	"runtime"
	"strings"
)

type ToolchainConfig struct {
	Filename  string
	OutputDir string
	Name      string /* Name of the end executable, if not present, it will be infered from the filename */
	Cleanup   bool   /* Whether to remove the generated .go code */
	Run       bool   /* Whether to run the program after building */
	Runtime   []byte
}

func RunToolchain(config *ToolchainConfig) error {
	diagnostics.Dbg("validating filename... (1/7)")

	exec_name, err := validate_filename(config.Filename)
	if err != nil {
		return err
	}

	if config.Name != "" {
		/* If there's a defined name, use that! */
		exec_name = config.Name
	}

	diagnostics.Dbgfmt("opening file at %s... (2/7)\n", config.Filename)

	data, err := open_file(config.Filename)
	if err != nil {
		return err
	}

	diagnostics.Dbg("tokenizing... (3/7)")

	tokens, err := tokenize(data, config.Filename)
	if err != nil {
		return err
	}

	diagnostics.Dbg("parsing... (4/7)")
	tree, err := parse(tokens, config.Filename)
	if err != nil {
		return err
	}

	diagnostics.Dbg("analyzing... (5/7)")
	t_tree, err := analyze(tree, config.Filename)
	if err != nil {
		return err
	}

	diagnostics.Dbg("generating go code... (6/7)")
	codegen := generate(t_tree, config.Runtime)

	diagnostics.Dbg("building... (7/7)")
	path, err := build(codegen, config.OutputDir, exec_name)
	if err != nil {
		return err
	}

	if config.Run {
		return run(path)
	}

	return nil
}

func validate_filename(filename string) (string, error) {
	base := filepath.Base(filename)

	if base == "." {
		return "", errors.New("path is empty or invalid!")
	}

	diagnostics.Dbg("Base: " + base)
	base, ok := strings.CutSuffix(base, ".psc")

	diagnostics.Dbg("Base (no suffix): " + base)

	if !ok {
		return "", errors.New("path does not point to a valid .psc file!")
	}

	return base, nil
}

func open_file(filename string) ([]byte, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func tokenize(file []byte, filename string) ([]lexer.Token, error) {
	src := []rune(string(file))
	lexer := lexer.NewLexer(src)

	tokens, err := lexer.Tokenize()
	diagnostics.PrintDiagnostics(lexer.Diagnostics(), filename)

	return tokens, err
}

func parse(tokens []lexer.Token, filename string) (ast.Node, error) {
	parser := parser.Parser{Tokens: tokens}
	tree, ds := parser.Parse()

	diagnostics.PrintDiagnostics(ds, filename)
	return tree, diagnostics.SummaryAsError(ds)
}

func analyze(tree ast.Node, filename string) (tast.Node, error) {
	analyzer := analyzer.New()
	mf, ok := tree.(*ast.MainFunction)
	if !ok {
		return nil, errors.New("main function not found")
	}
	t_tree, ds := analyzer.Analyze(mf)

	diagnostics.PrintDiagnostics(ds, filename)
	return t_tree, diagnostics.SummaryAsError(ds)
}

func generate(tree tast.Node, runtime []byte) []byte {
	codegen := generator.CodeGenerator{}

	code_bytes := codegen.Generate(tree)
	var data []byte = append(runtime, code_bytes...)

	return data
}

func build(codegen []byte, output_dir string, target string) (string, error) {
	/* code_path is the generated .go */
	code_path := filepath.Join(output_dir, target+".go")
	executable_name := target

	/* Add windows executable name */
	if runtime.GOOS == "windows" {
		executable_name += ".exe"
	}
	out_path := filepath.Join(output_dir, executable_name)

	if err := os.MkdirAll(output_dir, 0755); err != nil {
		return "", err
	}

	err := os.WriteFile(code_path, codegen, 0644)
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-o", out_path, code_path)
	diagnostics.Dbgfmt("Command: ", cmd.String())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return out_path, nil
}

func run(path string) error {
	run := exec.Command(path)
	run.Stderr = os.Stderr
	run.Stdout = os.Stdout
	run.Stdin = os.Stdin

	err := run.Run()
	return err
}
