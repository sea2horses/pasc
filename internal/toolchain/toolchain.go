package toolchain

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"pascc/internal/analyzer"
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/generator"
	"pascc/internal/i18n"
	"pascc/internal/lexer"
	"pascc/internal/parser"
	"path/filepath"
	"runtime"
	"runtime/debug"
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
	Filename       string
	OutputDir      string
	Name           string /* Name of the end executable, if not present, it will be infered from the filename */
	Cleanup        bool   /* Whether to remove the generated .go code */
	Run            bool   /* Whether to run the program after building */
	TempBuild      bool   /* This makes the build temporary, overrides the OutputDir and then cleans everything up */
	Runtime        []byte
	ExtRuntime     []byte
	ExtFastRuntime []byte
	RuntimeMod     []byte
	Vendor         fs.FS
	Fast           bool
	IndexBase      analyzer.IndexBase /* Array base; IndexBaseAuto lets the analyzer deduce it */
	Target         BuildTarget        /* Platform to compile for; the zero value is the host */
}

func RunToolchain(config *ToolchainConfig) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = errors.New(i18n.T(i18n.InternalCompilerError, recovered, debug.Stack()))
		}
	}()

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
	t_tree, err := analyze(tree, config.Filename, config.IndexBase)
	if err != nil {
		return err
	}

	diagnostics.Dbg("generating go code... (6/7)")
	codegen, err := generate(t_tree)
	if err != nil {
		return err
	}

	diagnostics.Dbg("building... (7/7)")
	path, err := build(
		codegen,
		output_dir,
		exec_name,
		config.Target,
		config.Runtime,
		config.ExtRuntime,
		config.ExtFastRuntime,
		config.RuntimeMod,
		config.Vendor,
		config.Fast,
	)
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
		return "", errors.New(i18n.T(i18n.InvalidPath))
	}

	diagnostics.Dbg("Base: " + base)
	base, ok := strings.CutSuffix(base, ".psc")

	diagnostics.Dbg("Base (no suffix): " + base)

	if !ok {
		return "", errors.New(i18n.T(i18n.NotPscFile, filename))
	}

	return base, nil
}

func open_file(filename string) ([]byte, error) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.New(i18n.T(i18n.FileNotFound, filename))
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.T(i18n.OpenFileFailed, filename), err)
	}

	return bytes, nil
}

func tokenize(file []byte, filename string) ([]lexer.Token, error) {
	src := decodeSource(file)
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

func analyze(tree ast.Node, filename string, indexBase analyzer.IndexBase) (tast.Node, error) {
	analyzer := analyzer.New()
	analyzer.IndexBase = indexBase
	mf, ok := tree.(*ast.MainFunction)
	if !ok {
		return nil, errors.New(i18n.T(i18n.MainNotFound))
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

type BuildTarget uint8

const (
	Native BuildTarget = iota /* The host's GOOS/GOARCH */
	WASM                      /* WebAssembly with WASI preview 1 (wasmtime, wasmer, node --experimental-wasi...) */
)

func (t BuildTarget) String() string {
	switch t {
	case Native:
		return "native"
	case WASM:
		return "wasm"
	}
	return fmt.Sprintf("BuildTarget(%d)", uint8(t))
}

func ParseBuildTarget(value string) (BuildTarget, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "native":
		return Native, nil
	case "wasm", "wasi", "wasip1":
		return WASM, nil
	}
	return Native, errors.New(i18n.T(i18n.InvalidBuildTarget, value))
}

/* goEnv returns the extra environment the go command needs to cross-compile */
func (t BuildTarget) goEnv() []string {
	switch t {
	case WASM:
		return []string{"GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0"}
	}
	return nil
}

/* outputName adds the extension the target expects to the executable name */
func (t BuildTarget) outputName(name string) string {
	switch t {
	case WASM:
		return name + ".wasm"
	}
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func build(codegen []byte, output_dir string, target_name string, target BuildTarget, _runtime, extRuntime, extFastRuntime, runtimeMod []byte, vendor fs.FS, fast bool) (string, error) {
	var err error
	output_dir, err = filepath.Abs(output_dir)
	if err != nil {
		return "", err
	}

	/* code_path is the generated .go */
	code_path := filepath.Join(output_dir, target_name+".go")
	runtime_path := filepath.Join(output_dir, "runtime.go")
	ext_path := filepath.Join(output_dir, "ext.go")

	out_path := filepath.Join(output_dir, target.outputName(target_name))

	if err := os.MkdirAll(output_dir, 0755); err != nil {
		return "", err
	}

	err = os.WriteFile(runtime_path, _runtime, 0644)
	if err != nil {
		return "", err
	}

	// Like runtime.go and ext.go, the inspectable copy is excluded from Go
	// package builds: an output directory inside a module (the default
	// ./build) would otherwise put one main package per program into
	// `go build ./...`. Naming the file explicitly (go run x.go) still works.
	err = os.WriteFile(code_path, append([]byte("//go:build ignore\n\n"), codegen...), 0644)
	if err != nil {
		return "", err
	}
	if fast {
		extRuntime = extFastRuntime
	}
	if len(extRuntime) == 0 {
		return "", errors.New(i18n.T(i18n.RuntimeExtMissing))
	}
	if err := os.WriteFile(ext_path, extRuntime, 0644); err != nil {
		return "", err
	}

	if !fast && (len(runtimeMod) == 0 || vendor == nil) {
		return "", errors.New(i18n.T(i18n.RuntimeModuleMissing))
	}

	module_dir, err := os.MkdirTemp("", "pascc-generated-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(module_dir)

	if err := os.WriteFile(filepath.Join(module_dir, "main.go"), codegen, 0644); err != nil {
		return "", err
	}
	// runtime.go is excluded from pascc's own build with a build tag; remove
	// that tag from the staged copy so it is compiled into generated programs.
	runtimeSource := bytes.TrimPrefix(_runtime, []byte("//go:build ignore\n\n"))
	if err := os.WriteFile(filepath.Join(module_dir, "runtime.go"), runtimeSource, 0644); err != nil {
		return "", err
	}
	extSource := bytes.TrimPrefix(extRuntime, []byte("//go:build ignore\n\n"))
	if err := os.WriteFile(filepath.Join(module_dir, "ext.go"), extSource, 0644); err != nil {
		return "", err
	}

	args := []string{"build", "-buildvcs=false", "-o", out_path}
	if fast {
		args = append(args, "main.go", "runtime.go", "ext.go")
	} else {
		if err := os.WriteFile(filepath.Join(module_dir, "go.mod"), runtimeMod, 0644); err != nil {
			return "", err
		}
		if err := copyEmbeddedVendor(vendor, module_dir); err != nil {
			return "", err
		}
		args = append(args, "-mod=vendor", ".")
	}
	goCmd, err := findGo(requiredGoVersion(runtimeMod))
	if err != nil {
		return "", err
	}
	cmd := exec.Command(goCmd.path, args...)
	cmd.Dir = module_dir
	cmd.Env = append(append(os.Environ(), goCmd.env...), target.goEnv()...)
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
		return errors.New(i18n.T(i18n.RuntimeVendorBroken))
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
