package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

type Analyzer struct {
	scope          *semantic.Scope
	type_table     *semantic.TypeTable
	readInferences []*semantic.InferType
	implicit       []*semantic.Symbol
	diagnostics.DiagnosticEmitter
}

func New() *Analyzer {
	/* Make base scope */
	return &Analyzer{scope: semantic.NewScope(nil), DiagnosticEmitter: diagnostics.NewDiagnosticEmitter()}
}

func (a *Analyzer) Analyze(program *ast.MainFunction) (*tast.MainFunction, []diagnostics.Diagnostic) {
	a.scope = semantic.NewScope(nil)
	a.type_table = semantic.NewTypeTable()
	a.readInferences = nil
	a.implicit = nil
	a.DiagnosticEmitter = diagnostics.NewDiagnosticEmitter()

	a.scope.Declare(
		&semantic.Symbol{
			Name:     program.Name,
			Kind:     semantic.SymbolFunction,
			Type:     semantic.Void_T{},
			Mutable:  false,
			Declared: program.Span,
		},
	)

	/* Builtins */
	a.declare_builtins()

	statements := a.analyze_statement_block(program.Stmts)
	for _, infer := range a.readInferences {
		if infer.Resolved == nil {
			infer.Resolved = semantic.String_T{}
		}
	}
	var implicitDeclarations []tast.Stmt
	for _, symbol := range a.implicit {
		if infer, ok := semantic.Resolve(symbol.Type).(*semantic.InferType); ok && infer.Resolved == nil {
			infer.Resolved = semantic.String_T{}
		}
		implicitDeclarations = append(implicitDeclarations, &tast.Declaration{
			NodeInfo: ast.NodeInfo{Span: symbol.Declared},
			Symbols:  []*semantic.Symbol{symbol},
			Type:     semantic.Resolve(symbol.Type),
		})
	}
	statements = append(implicitDeclarations, statements...)

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
