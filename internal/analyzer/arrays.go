package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_dimension(d *ast.Dimension) *tast.Dimension {
	ok := true

	symbol, declared := a.scope.Lookup(d.Name)
	var elemType semantic.Type
	if declared {
		if symbol.Kind != semantic.SymbolVariable {
			a.Report(d.NodeSpan(), "'%s' cannot be dimensioned because it is not a variable", d.Name)
			return nil
		}
		if semantic.IsArray(symbol.Type) {
			a.Report(d.NodeSpan(), "'%s' is already dimensioned", d.Name)
			a.Info(symbol.Declared, "declared here")
			return nil
		}
		elemType = symbol.Type
	} else {
		elemType = a.type_table.MakeInference()
		symbol = &semantic.Symbol{
			Type:     semantic.ArrayType{Elem: elemType, Rank: len(d.Dimensions)},
			Declared: d.NodeSpan(),
			Kind:     semantic.SymbolVariable,
			Mutable:  true,
			Name:     d.Name,
		}
		a.scope.Declare(symbol)
	}

	var dims []tast.TypedExpr
	for _, dim := range d.Dimensions {
		/* Try to parse all expressions */
		expr := a.analyze_expression(dim)

		if expr == nil {
			ok = false
			continue
		}

		/* Try coercing */
		expr, ok = coerce_to(expr, semantic.Integer_T{})
		if !ok {
			a.Report(dim.NodeSpan(), "dimension must be an integer")
			ok = false
			continue
		}

		dims = append(dims, expr)
	}

	if !ok {
		return nil
	}

	// A prior scalar definition supplies the element type. Implicitly
	// dimensioned arrays retain inference and learn their element type later.
	symbol.Type = semantic.ArrayType{Elem: elemType, Rank: len(d.Dimensions)}

	return &tast.Dimension{
		NodeInfo:   d.NodeInfo,
		Symbol:     symbol,
		Dimensions: dims,
	}
}

func (a *Analyzer) analyze_indexing(d *ast.Postfix, typed_expr tast.TypedExpr) tast.TypedExpr {
	/* Right now you can only index arrays */

	if typed_expr == nil || semantic.IsInvalid(typed_expr.Type()) {
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	lv, ok := typed_expr.(tast.LValue)
	if !ok {
		a.Report(d.NodeSpan(), "cannot index non-lvalue")
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	if !semantic.IsArray(typed_expr.Type()) {
		a.Report(d.NodeSpan(), "cannot index non-array type '%s'", typed_expr.Type())
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	underlying_type := (typed_expr.Type().(semantic.ArrayType)).Elem
	array_rank := (typed_expr.Type().(semantic.ArrayType)).Rank

	if len(d.Arguments) != array_rank {
		a.Report(d.NodeSpan(), "index count mismatch: expected %d, got %d", array_rank, len(d.Arguments))
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	ok = true
	indexes := []tast.TypedExpr{}
	for _, index := range d.Arguments {
		curr := a.analyze_expression(index)
		if curr == nil || semantic.IsInvalid(curr.Type()) {
			ok = false
			continue
		}

		curr, ok := coerce_to(curr, semantic.Integer_T{})
		if !ok {
			a.Report(curr.NodeSpan(), "index must be an integer")
			continue
		}
		indexes = append(indexes, curr)
	}

	if !ok {
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	return &tast.Index{
		NodeInfo:  d.NodeInfo,
		Target:    lv,
		Indexes:   indexes,
		IndexType: underlying_type,
	}
}
