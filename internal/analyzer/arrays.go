package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/semantic"
)

func (a *Analyzer) analyze_dimension(d *ast.Dimension) *tast.Dimension {
	ok := true

	inf := a.type_table.MakeInference()

	var symbol *semantic.Symbol = &semantic.Symbol{
		Type:     semantic.ArrayType{Elem: inf, Rank: len(d.Dimensions)},
		Declared: d.NodeSpan(),
		Kind:     semantic.SymbolVariable,
		Mutable:  false,
		Name:     d.Name,
	}

	/* Check that symbol doesn't exist */
	if s, ok := a.scope.Lookup(d.Name); ok {
		a.Report(d.NodeSpan(), "'%s' is already defined in this scope", d.Name)
		a.Info(s.Declared, "declared here")
		ok = false
	} else {
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
		expr, ok = coerce_to(expr, semantic.IntegerType{})
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

	return &tast.Dimension{
		NodeInfo:   d.NodeInfo,
		Symbol:     symbol,
		Dimensions: dims,
	}
}

func (a *Analyzer) analyze_indexing(d *ast.Index) tast.TypedExpr {
	/* Right now you can only index arrays */

	/* Analyze expression to be indexed */
	typed_expr := a.analyze_expression(d.Target)
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

	if len(d.Indexes) != array_rank {
		a.Report(d.NodeSpan(), "index count mismatch: expected %d, got %d", array_rank, len(d.Indexes))
		return &tast.ErrorExpr{
			NodeInfo: d.NodeInfo,
		}
	}

	ok = true
	indexes := []tast.TypedExpr{}
	for _, index := range d.Indexes {
		curr := a.analyze_expression(index)
		if curr == nil || semantic.IsInvalid(curr.Type()) {
			ok = false
			continue
		}

		curr, ok := coerce_to(curr, semantic.IntegerType{})
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
