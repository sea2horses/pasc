package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
)

func (a *Analyzer) analyze_operator(op ast.Operator) *tast.Operator {
	return &tast.Operator{
		NodeInfo: op.NodeInfo,
		Op:       op.Type,
	}
}
