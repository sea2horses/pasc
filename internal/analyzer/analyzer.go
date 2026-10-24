package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

type Analyzer struct {
	scope      *semantic.Scope
	type_table *semantic.TypeTable
	diagnostics.DiagnosticEmitter
}

func New() *Analyzer {
	/* Make base scope */
	return &Analyzer{scope: semantic.NewScope(nil), DiagnosticEmitter: diagnostics.NewDiagnosticEmitter()}
}

func (a *Analyzer) Analyze(program *ast.MainFunction) (*tast.MainFunction, []diagnostics.Diagnostic) {
	a.scope = semantic.NewScope(nil)
	a.type_table = semantic.NewTypeTable()
	a.DiagnosticEmitter = diagnostics.NewDiagnosticEmitter()

	a.scope.Declare(
		&semantic.Symbol{
			Name:     program.Name,
			Kind:     semantic.SymbolFunction,
			Type:     semantic.VoidType{},
			Mutable:  false,
			Declared: program.Span,
		},
	)

	/* Builtins */
	a.declare_builtins()

	statements := a.analyze_statement_block(program.Stmts)

	return &tast.MainFunction{
		NodeInfo: program.NodeInfo,
		Name:     program.Name,
		Stmts:    statements,
	}, a.Diagnostics()
}

func (a *Analyzer) new_scope() {
	a.scope = semantic.NewScope(a.scope)
}

func (a *Analyzer) destroy_scope() {
	a.scope = a.scope.Parent()
}
