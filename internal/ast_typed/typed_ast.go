package tast

import (
	"fmt"
	"pascc/internal/ast"
	"pascc/internal/models"
	"pascc/internal/semantic"
)

type Node interface {
	Info() ast.NodeInfo
	NodeSpan() models.Span
}

type Stmt interface {
	Node
}

type TypedExpr interface {
	Node
	exprNode()
	Type() semantic.Type
}

type AssignmentOrigin struct {
	Span    models.Span
	Message string
}

type LValue interface {
	TypedExpr
	lvalue()
	AssignmentOrigin() *AssignmentOrigin
	IsMutable() bool
}

type ErrorExpr struct {
	ast.NodeInfo
}

func (ErrorExpr) exprNode() {}

func (ErrorExpr) Type() semantic.Type {
	return semantic.Invalid_T{}
}

type StringLiteral struct {
	ast.NodeInfo
	Value string
}

func (*StringLiteral) exprNode() {}

func (*StringLiteral) Type() semantic.Type {
	return semantic.String_T{}
}

type NumberLiteral struct {
	ast.NodeInfo
	Int          string
	Frac         string
	ResolvedType semantic.Type
}

func (*NumberLiteral) exprNode() {}

/* ALL number literals are real by default */
/* TODO: Add warning to round any operation when assigning to an integer variable */
func (nl *NumberLiteral) Type() semantic.Type {
	return semantic.Real_T{}
}

type BooleanLiteral struct {
	ast.NodeInfo
	Value bool
}

func (BooleanLiteral) exprNode() {}

func (BooleanLiteral) Type() semantic.Type {
	return semantic.Bool_T{}
}

type VariableExpr struct {
	ast.NodeInfo
	Symbol *semantic.Symbol
}

func (VariableExpr) exprNode() {}

func (VariableExpr) lvalue() {}

func (v VariableExpr) Type() semantic.Type {
	return v.Symbol.Type
}

func (v VariableExpr) AssignmentOrigin() *AssignmentOrigin {
	return &AssignmentOrigin{
		Span:    v.Symbol.Declared,
		Message: fmt.Sprintf("%q declared here with type %s", v.Symbol.Name, v.Symbol.Type),
	}
}

func (v VariableExpr) IsMutable() bool {
	return v.Symbol.Mutable
}

type Operator struct {
	ast.NodeInfo
	Op semantic.OperatorType
}

type UnaryOperation struct {
	ast.NodeInfo
	Op           *Operator
	Expr         TypedExpr
	ResolvedType semantic.Type
}

func (UnaryOperation) exprNode() {}

func (uo UnaryOperation) Type() semantic.Type {
	return uo.ResolvedType
}

type BinaryOperation struct {
	ast.NodeInfo
	LHS          TypedExpr
	Op           *Operator
	RHS          TypedExpr
	ResolvedType semantic.Type
}

func (BinaryOperation) exprNode() {}

func (uo BinaryOperation) Type() semantic.Type {
	return uo.ResolvedType
}

type Declaration struct {
	ast.NodeInfo
	Symbols []*semantic.Symbol
	Type    semantic.Type
}

type Assignment struct {
	ast.NodeInfo
	Declarative bool /* MUST be used if it's first use */
	Target      LValue
	Value       TypedExpr
}

type Cast struct {
	ast.NodeInfo
	TargetType     semantic.Type
	ConversionKind *semantic.ConversionKind
	Expr           TypedExpr
}

func (Cast) exprNode() {}

func (c Cast) Type() semantic.Type {
	return c.TargetType
}

type Index struct {
	ast.NodeInfo
	Target    LValue
	Indexes   []TypedExpr
	IndexType semantic.Type
}

func (Index) exprNode() {}

func (Index) lvalue() {}

func (i Index) Type() semantic.Type {
	return i.IndexType
}

type Call struct {
	ast.NodeInfo
	Callable  TypedExpr
	Arguments []TypedExpr
	Return    semantic.Type
}

func (*Call) exprNode() {}

func (i Call) Type() semantic.Type {
	return i.Return
}

func (i Index) AssignmentOrigin() *AssignmentOrigin {
	return i.Target.AssignmentOrigin()
}

func (i Index) IsMutable() bool {
	/* TODO: ACTUALLY define mutability on indexing */
	return true
}

type While struct {
	ast.NodeInfo
	Condition TypedExpr
	Stmts     []Stmt
}

type DoWhile struct {
	ast.NodeInfo
	Condition TypedExpr
	Stmts     []Stmt
}

/* PSEINT Restriction: not any lvalue can be used for a for, only a variable one */
type For struct {
	ast.NodeInfo
	Var   *VariableExpr
	Start TypedExpr
	Until TypedExpr
	Step  TypedExpr
	Stmts []Stmt
}

type TimeOut struct {
	ast.NodeInfo
	Amount TypedExpr
	Unit   semantic.TimeUnit
}

type Case struct {
	ast.NodeInfo
	Value TypedExpr
	Stmts []Stmt
}

type Switch struct {
	ast.NodeInfo
	Base    TypedExpr
	Cases   []*Case
	Default []Stmt
}

type Write struct {
	ast.NodeInfo
	Newline bool
	Content []TypedExpr
}

type Dimension struct {
	ast.NodeInfo
	Symbol     *semantic.Symbol
	Dimensions []TypedExpr
}

/* Although multiple semantically, in the TAST reads are treated as individual */
type Read struct {
	ast.NodeInfo
	Into LValue
}

type If struct {
	ast.NodeInfo
	Condition TypedExpr
	Stmts     []Stmt
	Else      *Else
}

type Else struct {
	ast.NodeInfo
	Stmts []Stmt
}

type ClearScreen struct {
	ast.NodeInfo
}

type MainFunction struct {
	ast.NodeInfo
	Name  string
	Stmts []Stmt
}
