package ast

import (
	"fmt"
	"pseint-compiled/internal/models"
	"pseint-compiled/internal/semantic"
	"pseint-compiled/internal/utils"
	"strings"
)

type NodeInfo struct {
	Span models.Span
}

func (ni NodeInfo) JoinSpan(span models.Span) {
	ni.Span = models.JoinSpans(ni.Span, span)
}

func (ni NodeInfo) Info() NodeInfo {
	return ni
}

func (ni NodeInfo) NodeSpan() models.Span {
	return ni.Span
}

/* Node implemented by every AST Node */
type Node interface {
	Info() NodeInfo
	NodeSpan() models.Span
	String() string
}

type Expr interface {
	Node
}

type Stmt interface {
	Node
}

type StringLiteral struct {
	NodeInfo
	Content string
}

func (sl StringLiteral) String() string {
	return fmt.Sprintf("\"%s\"", sl.Content)
}

type NumberLiteral struct {
	NodeInfo
	Int  string
	Frac string /* TODO: THIS IS FUCKED, because of 1.001 */
}

func (nl NumberLiteral) String() string {
	return fmt.Sprintf("%s.%s", nl.Int, nl.Frac)
}

type BoolLiteral struct {
	NodeInfo
	Value bool
}

func (bl BoolLiteral) String() string {
	return fmt.Sprintf("%t", bl.Value)
}

type Variable struct {
	NodeInfo
	Name string
}

func (v Variable) String() string {
	return v.Name
}

type TypeRef struct {
	NodeInfo
	Name string
}

func (t TypeRef) String() string {
	return fmt.Sprintf("type<%s>", t.Name)
}

type Operator struct {
	NodeInfo
	Type semantic.OperatorType
}

func (o Operator) String() string {
	return o.Type.String()
}

func (o Operator) Precedence() int8 {
	return semantic.BinaryOperatorPrecedence(o.Type)
}

type UnaryOperation struct {
	NodeInfo
	Op   Operator
	Expr Expr
}

func (uo UnaryOperation) String() string {
	return fmt.Sprintf("(%s %s)", uo.Op, uo.Expr)
}

type BinaryOperation struct {
	NodeInfo
	LHS Expr
	Op  Operator
	RHS Expr
}

func (bo BinaryOperation) String() string {
	return fmt.Sprintf("(%s %s %s)", bo.LHS, bo.Op, bo.RHS)
}

type Index struct {
	NodeInfo
	Indexes []Expr
	Target  Expr
}

func (i Index) String() string {
	return fmt.Sprintf("%s[%s]", i.Target, strings.Join(utils.Map(i.Indexes, func(e Expr) string {
		return e.String()
	}), ", "))
}

/* Expression as Statement */
type ExprStmt struct {
	NodeInfo
	Expr Expr
}

func (e ExprStmt) String() string {
	return fmt.Sprintf("%s", e.String())
}

type Declaration struct {
	NodeInfo
	Names []string
	Type  *TypeRef
}

func (d Declaration) String() string {
	return fmt.Sprintf("decl %s as %s", strings.Join(d.Names, ", "), d.Type)
}

type Assignment struct {
	NodeInfo
	Target  Expr
	Content Expr
}

func (a Assignment) String() string {
	return fmt.Sprintf("%s = %s", a.Target, a.Content)
}

type Dimension struct {
	NodeInfo
	Name       string
	Dimensions []Expr
}

func (d Dimension) String() string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Dimension %s [", d.Name)
	for _, dim := range d.Dimensions {
		builder.WriteString(dim.String())
	}
	builder.WriteRune(']')

	return builder.String()
}

type Write struct {
	NodeInfo
	Newline bool
	Print   []Expr
}

func (w Write) String() string {
	return fmt.Sprintf("Write %s", w.Print)
}

type Read struct {
	NodeInfo
	Into []Expr
}

func (r Read) String() string {
	return fmt.Sprintf("Read %s", r.Into)
}

type TimeOut struct {
	NodeInfo
	Amount Expr
	Unit   semantic.TimeUnit
}

func (to TimeOut) String() string {
	return fmt.Sprintf("Wait %s %s", to.Amount, to.Unit)
}

type If struct {
	NodeInfo
	Condition Expr
	Stmts     []Stmt
	Else      *Else
}

func (i If) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("If '%s' {\n", i.Condition))
	for _, stmt := range i.Stmts {
		builder.WriteString(stmt.String())
		builder.WriteRune('\n')
	}
	builder.WriteString("} ")
	if i.Else != nil {
		builder.WriteString(i.Else.String())
	}

	return builder.String()
}

type Else struct {
	NodeInfo
	Stmts []Stmt
}

func (e Else) String() string {
	var builder strings.Builder

	builder.WriteString("Else {\n")
	for _, stmt := range e.Stmts {
		builder.WriteString(stmt.String())
		builder.WriteRune('\n')
	}
	builder.WriteString("}")

	return builder.String()
}

type While struct {
	NodeInfo
	Condition Expr
	Stmts     []Stmt
}

func (w While) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("While '%s' {\n", w.Condition))
	for _, stmt := range w.Stmts {
		builder.WriteString(stmt.String())
		builder.WriteRune('\n')
	}
	builder.WriteString("}")

	return builder.String()
}

type For struct {
	NodeInfo
	Start *Assignment
	Until Expr
	Step  Expr
	Stmts []Stmt
}

func (f For) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("For %s -> %s", f.Start, f.Until))
	if f.Step != nil {
		builder.WriteString(fmt.Sprintf(" wth step %s", f.Step))
	}
	builder.WriteString(" {")

	for _, s := range f.Stmts {
		builder.WriteRune('\n')
		builder.WriteString(s.String())
	}
	builder.WriteRune('\n')
	builder.WriteString("}")

	return builder.String()
}

type Switch struct {
	NodeInfo
	Base    Expr
	Cases   []*Case
	Default []Stmt
}

func (s Switch) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Switch '%s': {\n", s.Base))
	for _, c := range s.Cases {
		builder.WriteString(c.String())
		builder.WriteRune('\n')
	}

	if s.Default != nil {
		builder.WriteString("Default: {\n")
		for _, stmt := range s.Default {
			builder.WriteString(stmt.String())
			builder.WriteRune('\n')
		}
		builder.WriteString("}")
	}
	builder.WriteString("}")

	return builder.String()
}

type Case struct {
	NodeInfo
	Clause Expr
	Stmts  []Stmt
}

func (c Case) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Case '%s': {\n", c.Clause))
	for _, stmt := range c.Stmts {
		builder.WriteString(stmt.String())
		builder.WriteRune('\n')
	}
	builder.WriteString("}")

	return builder.String()
}

type Call struct {
	NodeInfo
	Callable  Expr
	Arguments []Expr
}

func (c Call) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Call '%s': (\n", c.Callable))

	args := utils.Map(c.Arguments, func(a Expr) string {
		return a.String()
	})

	builder.WriteString(strings.Join(args, ", "))
	builder.WriteString(")")

	return builder.String()
}

type MainFunction struct {
	NodeInfo
	Name  string
	Stmts []Stmt
}

type ClearScreen struct {
	NodeInfo
}

func (cs ClearScreen) String() string {
	return "clear screen"
}

func (mf MainFunction) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("MainFunction '%s' {\n", mf.Name))
	for _, stmt := range mf.Stmts {
		builder.WriteString(stmt.String())
		builder.WriteRune('\n')
	}
	builder.WriteString("}")

	return builder.String()
}
