package generator

import (
	"bytes"
	"fmt"
	tast "pascc/internal/ast_typed"

	"github.com/dave/jennifer/jen"
)

type Generator struct {
	file *jen.File
}

func NewGenerator() *Generator {
	return &Generator{
		file: jen.NewFile("main"),
	}
}

func (g *Generator) Gen(tree tast.Node) ([]byte, error) {
	g.generate(tree)

	buf := &bytes.Buffer{}
	err := g.file.Render(buf)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (g *Generator) generate(tree tast.Node) *jen.Statement {
	switch node := tree.(type) {
	case *tast.VariableExpr:
		return g.genVariable(node)
	case *tast.Declaration:
		return g.genDeclaration(node)
	case *tast.Assignment:
		return g.genAssignment(node)
	case *tast.Cast:
		return g.genCast(node)
	case *tast.TimeOut:
		return g.genTimeout(node)
	case *tast.WaitKey:
		return g.genWaitKey(node)
	case *tast.MainFunction:
		g.genMainFunction(node)
		return nil
	case *tast.NumberLiteral:
		return g.genNumberLiteral(node)
	case *tast.StringLiteral:
		return g.genStringLiteral(node)
	case *tast.BooleanLiteral:
		return g.genBoolLiteral(node)
	case *tast.Write:
		return g.genWrite(node)
	case *tast.Read:
		return g.genRead(node)
	case *tast.ClearScreen:
		return g.genClearScreen(node)
	case *tast.Call:
		return g.genCall(node)
	case *tast.ExprStmt:
		return g.generate(node.Expr)
	case *tast.UnaryOperation:
		return g.genUnExpr(node)
	case *tast.BinaryOperation:
		return g.genBinExpr(node)
	case *tast.If:
		return g.genIf(node)
	case *tast.While:
		return g.genWhile(node)
	case *tast.Case:
		return g.genCase(node)
	case *tast.Switch:
		return g.genSwitch(node)
	case *tast.For:
		return g.genFor(node)
	case *tast.Dimension:
		return g.genDimension(node)
	case *tast.Index:
		return g.genIndex(node)
	case *tast.DoWhile:
		return g.genDoWhile(node)
	default:
		panic(fmt.Sprintf("unsupported typed AST node %T", tree))
	}
}
