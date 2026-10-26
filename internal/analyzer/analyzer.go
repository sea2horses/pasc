package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
)

type Analyzer struct {
	scope           *semantic.Scope
	type_table      *semantic.TypeTable
	readInferences  []*semantic.InferType
	implicit        []*semantic.Symbol
	implicitScope   *semantic.Scope
	functionSymbols map[string]*semantic.Symbol
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
	a.functionSymbols = map[string]*semantic.Symbol{}
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
	a.implicitScope = a.scope
	for _, function := range program.Functions {
		params := make([]semantic.Type, len(function.Arguments))
		for i := range params {
			params[i] = a.type_table.MakeInference()
		}
		result := semantic.Type(semantic.Void_T{})
		if function.HasReturn {
			result = a.type_table.MakeInference()
		}
		byReference := append([]bool(nil), function.ByReference...)
		symbol := &semantic.Symbol{
			Name: function.Name,
			Kind: semantic.SymbolFunction,
			Type: semantic.FunctionType{Params: params, ByReference: byReference, Return: result},
		}
		if !a.scope.Declare(symbol) {
			a.Report(function.NodeSpan(), "function '%s' is already declared", function.Name)
			continue
		}
		a.functionSymbols[semantic.NormalizeName(function.Name)] = symbol
	}

	statements := a.analyze_statement_block(program.Stmts)
	mainImplicit := a.implicit
	a.implicit = nil
	typedFunctions := make([]*tast.Function, 0, len(program.Functions))
	// Analyze callers before callees where possible so calls in a function body
	// can provide argument types to the functions they invoke.
	for i := len(program.Functions) - 1; i >= 0; i-- {
		function := program.Functions[i]
		if a.functionSymbols[semantic.NormalizeName(function.Name)] == nil {
			continue
		}
		typedFunctions = append(typedFunctions, a.analyze_function(function))
	}
	for left, right := 0, len(typedFunctions)-1; left < right; left, right = left+1, right-1 {
		typedFunctions[left], typedFunctions[right] = typedFunctions[right], typedFunctions[left]
	}
	for _, function := range typedFunctions {
		function.Stmts = append(a.implicitDeclarations(function.ImplicitSymbols, false), function.Stmts...)
		for _, symbol := range function.ImplicitSymbols {
			if infer, ok := semantic.Resolve(symbol.Type).(*semantic.InferType); ok && infer.Resolved == nil {
				infer.Resolved = semantic.String_T{}
			}
		}
	}
	for _, function := range program.Functions {
		symbol := a.functionSymbols[semantic.NormalizeName(function.Name)]
		if symbol == nil {
			continue
		}
		fnType := semantic.Resolve(symbol.Type).(semantic.FunctionType)
		for i, param := range fnType.Params {
			if infer, ok := semantic.Resolve(param).(*semantic.InferType); ok && infer.Resolved == nil {
				infer.Resolved = semantic.String_T{}
				name := "argument"
				if i < len(function.Arguments) {
					name = function.Arguments[i]
				}
				a.Warn(function.NodeSpan(), "cannot infer type of function argument '%s'; using string", name)
			}
		}
		if function.HasReturn {
			if infer, ok := semantic.Resolve(fnType.Return).(*semantic.InferType); ok && infer.Resolved == nil {
				infer.Resolved = semantic.String_T{}
				a.Warn(function.NodeSpan(), "cannot infer return type of function '%s'; using string", function.Name)
			}
		}
	}
	for _, infer := range a.readInferences {
		if infer.Resolved == nil {
			infer.Resolved = semantic.String_T{}
		}
	}
	implicitDeclarations := a.implicitDeclarations(mainImplicit, true)
	statements = append(implicitDeclarations, statements...)

	return &tast.MainFunction{
		NodeInfo:  program.NodeInfo,
		Name:      program.Name,
		Stmts:     statements,
		Functions: typedFunctions,
	}, a.Diagnostics()
}

func (a *Analyzer) analyze_function(function *ast.Function) *tast.Function {
	functionSymbol := a.functionSymbols[semantic.NormalizeName(function.Name)]
	fnType := semantic.Resolve(functionSymbol.Type).(semantic.FunctionType)
	previousScope := a.scope
	previousImplicitScope := a.implicitScope
	functionScope := semantic.NewScope(previousScope)
	a.scope = functionScope
	a.implicitScope = functionScope
	a.implicit = nil

	parameters := make([]*semantic.Symbol, len(function.Arguments))
	for i, name := range function.Arguments {
		parameter := &semantic.Symbol{
			Name:        name,
			Kind:        semantic.SymbolParameter,
			Type:        fnType.Params[i],
			Declared:    function.NodeSpan(),
			Mutable:     true,
			ByReference: i < len(function.ByReference) && function.ByReference[i],
		}
		if !functionScope.Declare(parameter) {
			a.Report(function.NodeSpan(), "duplicate function argument '%s'", name)
		}
		parameters[i] = parameter
	}

	statements := make([]tast.Stmt, 0, len(function.Stmts))
	for _, statement := range function.Stmts {
		if typed := a.analyze_statement(statement); typed != nil {
			statements = append(statements, typed...)
		}
	}

	var returnValue tast.TypedExpr
	if function.HasReturn {
		returnNode := &ast.Variable{NodeInfo: function.NodeInfo, Name: function.ReturnVariable}
		returnValue = a.analyze_variable(returnNode)
		var ok bool
		returnValue, ok = coerce_to(returnValue, fnType.Return)
		if !ok {
			a.Report(function.NodeSpan(), "function '%s' returns an incompatible value", function.Name)
			returnValue = &tast.ErrorExpr{NodeInfo: function.NodeInfo}
		}
	}
	implicitSymbols := a.implicit
	a.implicit = nil
	a.scope = previousScope
	a.implicitScope = previousImplicitScope

	return &tast.Function{
		NodeInfo:        function.NodeInfo,
		Name:            function.Name,
		Parameters:      parameters,
		Return:          returnValue,
		HasReturn:       function.HasReturn,
		Stmts:           statements,
		ImplicitSymbols: implicitSymbols,
	}
}

func (a *Analyzer) implicitDeclarations(symbols []*semantic.Symbol, defaultString bool) []tast.Stmt {
	declarations := make([]tast.Stmt, 0, len(symbols))
	for _, symbol := range symbols {
		if infer, ok := semantic.Resolve(symbol.Type).(*semantic.InferType); ok && infer.Resolved == nil {
			if defaultString {
				infer.Resolved = semantic.String_T{}
			}
		}
		declarations = append(declarations, &tast.Declaration{
			NodeInfo: ast.NodeInfo{Span: symbol.Declared},
			Symbols:  []*semantic.Symbol{symbol},
			Type:     semantic.Resolve(symbol.Type),
		})
	}
	return declarations
}

func (a *Analyzer) new_scope() {
	a.scope = semantic.NewScope(a.scope)
}

func (a *Analyzer) destroy_scope() {
	a.scope = a.scope.Parent()
}
