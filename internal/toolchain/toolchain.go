package toolchain

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
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

func tempDir() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(cache, "pasc", ".pasctmp")
	return path, nil
}

type ToolchainConfig struct {
	Filename   string
	OutputDir  string
	Name       string /* Name of the end executable, if not present, it will be infered from the filename */
	Cleanup    bool   /* Whether to remove the generated .go code */
	Run        bool   /* Whether to run the program after building */
	TempBuild  bool   /* This makes the build temporary, overrides the OutputDir and then cleans everything up */
	Runtime    []byte
	RuntimeMod []byte
	Vendor     fs.FS
}

func RunToolchain(config *ToolchainConfig) (err error) {
	output_dir := config.OutputDir
	if config.TempBuild {
		output_dir, err = tempDir()
		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, clean_temp_directory())
		}()
	}

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
	codegen, err := generate(t_tree)
	if err != nil {
		return err
	}

	diagnostics.Dbg("building... (7/7)")
	path, err := build(codegen, output_dir, exec_name, config.Runtime, config.RuntimeMod, config.Vendor)
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

	tokens, _ := lexer.Tokenize()
	diagnostics.PrintDiagnostics(lexer.Diagnostics(), filename)

	return tokens, diagnostics.SummaryAsError(lexer.Diagnostics())
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

func generate(tree tast.Node) ([]byte, error) {
	codegen := generator.NewGenerator()

	data, err := codegen.Gen(tree)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func build(codegen []byte, output_dir string, target string, _runtime, runtimeMod []byte, vendor fs.FS) (string, error) {
	var err error
	output_dir, err = filepath.Abs(output_dir)
	if err != nil {
		return "", err
	}

	/* code_path is the generated .go */
	code_path := filepath.Join(output_dir, target+".go")
	runtime_path := filepath.Join(output_dir, "runtime.go")

	executable_name := target

	/* Add windows executable name */
	if runtime.GOOS == "windows" {
		executable_name += ".exe"
	}
	out_path := filepath.Join(output_dir, executable_name)

	if err := os.MkdirAll(output_dir, 0755); err != nil {
		return "", err
	}

	err = os.WriteFile(runtime_path, _runtime, 0644)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(code_path, codegen, 0644)
	if err != nil {
		return "", err
	}

	if len(runtimeMod) == 0 || vendor == nil {
		return "", errors.New("runtime Go module or embedded vendor tree is missing")
	}

	module_dir, err := os.MkdirTemp("", "pascc-generated-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(module_dir)

	if err := os.WriteFile(filepath.Join(module_dir, "go.mod"), runtimeMod, 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(module_dir, "main.go"), codegen, 0644); err != nil {
		return "", err
	}
	// runtime.go is excluded from pascc's own build with a build tag; remove
	// that tag from the staged copy so it is compiled into generated programs.
	runtimeSource := bytes.TrimPrefix(_runtime, []byte("//go:build ignore\n\n"))
	if err := os.WriteFile(filepath.Join(module_dir, "runtime.go"), runtimeSource, 0644); err != nil {
		return "", err
	}
	if err := copyEmbeddedVendor(vendor, module_dir); err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-buildvcs=false", "-mod=vendor", "-o", out_path, ".")
	cmd.Dir = module_dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	return out_path, nil
}

func copyEmbeddedVendor(vendor fs.FS, moduleDir string) error {
	const sourceRoot = "vendor_clean"
	if _, err := fs.Stat(vendor, sourceRoot+"/modules.txt"); err != nil {
		return errors.New("embedded runtime vendor tree is missing vendor/modules.txt")
	}

	return fs.WalkDir(vendor, sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(path, sourceRoot+"/")
		if path == sourceRoot {
			rel = "."
		}
		dest := filepath.Join(moduleDir, "vendor", filepath.FromSlash(rel))
		if entry.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := fs.ReadFile(vendor, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}

func run(path string) error {
	cmd := exec.Command(path)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return err
	}

	/* All this to prevent Cntrl-C to NOT cleanup .temppasc */
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err

	case sig := <-signals:
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
		/* discard done to not throw an error on interrupt! */
		<-done
		return nil
	}
}

func clean_temp_directory() error {
	path, _ := tempDir()
	return os.RemoveAll(path)
}
