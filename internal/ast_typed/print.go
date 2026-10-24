package tast

import (
	"fmt"
	"pascc/internal/models"
	"pascc/internal/semantic"
	"strings"
)

// Print prints a TAST node to stdout.
func Print(node Node) {
	fmt.Print(Dump(node))
}

// Dump returns a textual tree representation of a TAST node.
func Dump(node Node) string {
	var builder strings.Builder

	printer := treePrinter{
		builder: &builder,
	}

	switch node := node.(type) {
	case *MainFunction:
		printer.printMainFunction(node, 0)

	case MainFunction:
		printer.printMainFunction(&node, 0)

	case TypedExpr:
		printer.printExpr(node, 0)

	default:
		printer.printStmt(node, 0)
	}

	return builder.String()
}

type treePrinter struct {
	builder *strings.Builder
}

func (p *treePrinter) line(depth int, format string, args ...any) {
	p.builder.WriteString(strings.Repeat("  ", depth))
	fmt.Fprintf(p.builder, format, args...)
	p.builder.WriteByte('\n')
}

// -----------------------------------------------------------------------------
// Main function
// -----------------------------------------------------------------------------

func (p *treePrinter) printMainFunction(function *MainFunction, depth int) {
	if function == nil {
		p.line(depth, "<nil MainFunction>")
		return
	}

	p.line(
		depth,
		"MainFunction name=%q span=%s",
		function.Name,
		formatSpan(function.NodeSpan()),
	)

	if len(function.Stmts) == 0 {
		p.line(depth+1, "Statements: <empty>")
		return
	}

	p.line(depth+1, "Statements:")

	for index, stmt := range function.Stmts {
		p.line(depth+2, "[%d]", index)
		p.printStmt(stmt, depth+3)
	}
}

// -----------------------------------------------------------------------------
// Statements
// -----------------------------------------------------------------------------

func (p *treePrinter) printStmt(stmt Stmt, depth int) {
	if stmt == nil {
		p.line(depth, "<nil statement>")
		return
	}

	switch stmt := stmt.(type) {
	case Assignment:
		p.printAssignment(&stmt, depth)

	case *Assignment:
		p.printAssignment(stmt, depth)

	case While:
		p.printWhile(&stmt, depth)

	case *While:
		p.printWhile(stmt, depth)

	case Switch:
		p.printSwitch(&stmt, depth)

	case *Switch:
		p.printSwitch(stmt, depth)

	case Case:
		p.printCase(&stmt, depth)

	case *Case:
		p.printCase(stmt, depth)

	case Write:
		p.printWrite(&stmt, depth)

	case *Write:
		p.printWrite(stmt, depth)

	case Read:
		p.printRead(&stmt, depth)

	case *Read:
		p.printRead(stmt, depth)

	case If:
		p.printIf(&stmt, depth)

	case *If:
		p.printIf(stmt, depth)

	case Else:
		p.printElse(&stmt, depth)

	case *Else:
		p.printElse(stmt, depth)

	case ClearScreen:
		p.printClearScreen(&stmt, depth)

	case *ClearScreen:
		p.printClearScreen(stmt, depth)

		// 	case Declaration:
		// 		p.printDeclaration(&stmt, depth)
		//
		// 	case *Declaration:
		// 		p.printDeclaration(stmt, depth)

	case MainFunction:
		p.printMainFunction(&stmt, depth)

	case *MainFunction:
		p.printMainFunction(stmt, depth)

	default:
		p.line(depth, "UnknownStatement type=%T", stmt)
	}
}

func (p *treePrinter) printStatements(label string, statements []Stmt, depth int) {
	if len(statements) == 0 {
		p.line(depth, "%s: <empty>", label)
		return
	}

	p.line(depth, "%s:", label)
	for index, statement := range statements {
		p.line(depth+1, "[%d]", index)
		p.printStmt(statement, depth+2)
	}
}

func (p *treePrinter) printAssignment(assignment *Assignment, depth int) {
	if assignment == nil {
		p.line(depth, "<nil Assignment>")
		return
	}

	p.line(
		depth,
		"Assignment declarative=%t span=%s",
		assignment.Declarative,
		formatSpan(assignment.NodeSpan()),
	)

	p.line(depth+1, "Target:")
	p.printLValue(assignment.Target, depth+2)

	p.line(depth+1, "Value:")
	p.printExpr(assignment.Value, depth+2)
}

// func (p *treePrinter) printDeclaration(declaration *Declaration, depth int) {
// 	if declaration == nil {
// 		p.line(depth, "<nil Declaration>")
// 		return
// 	}
//
// 	p.line(depth, "Declaration type=%s span=%s", formatType(declaration.Type), formatSpan(declaration.NodeSpan()))
// 	if declaration.Symbol == nil {
// 		p.line(depth+1, "Symbol: <nil>")
// 		return
// 	}
//
// 	p.line(depth+1, "Symbol:")
// 	p.line(depth+2, "Name: %q", declaration.Symbol.Name)
// 	p.line(depth+2, "Kind: %s", formatSymbolKind(declaration.Symbol.Kind))
// 	p.line(depth+2, "Mutable: %t", declaration.Symbol.Mutable)
// 	p.line(depth+2, "Declared: %s", formatSpan(declaration.Symbol.Declared))
// }

func (p *treePrinter) printWhile(stmt *While, depth int) {
	if stmt == nil {
		p.line(depth, "<nil While>")
		return
	}

	p.line(
		depth,
		"While span=%s",
		formatSpan(stmt.NodeSpan()),
	)

	p.line(depth+1, "Condition:")
	p.printExpr(stmt.Condition, depth+2)

	if len(stmt.Stmts) == 0 {
		p.line(depth+1, "Body: <empty>")
		return
	}

	p.line(depth+1, "Body:")

	for index, child := range stmt.Stmts {
		p.line(depth+2, "[%d]", index)
		p.printStmt(child, depth+3)
	}
}

func (p *treePrinter) printSwitch(stmt *Switch, depth int) {
	if stmt == nil {
		p.line(depth, "<nil Switch>")
		return
	}

	p.line(depth, "Switch span=%s", formatSpan(stmt.NodeSpan()))
	p.line(depth+1, "Base:")
	p.printExpr(stmt.Base, depth+2)

	if len(stmt.Cases) == 0 {
		p.line(depth+1, "Cases: <empty>")
	} else {
		p.line(depth+1, "Cases:")
		for index, item := range stmt.Cases {
			p.line(depth+2, "[%d]", index)
			p.printCase(item, depth+3)
		}
	}

	p.printStatements("Default", stmt.Default, depth+1)
}

func (p *treePrinter) printCase(stmt *Case, depth int) {
	if stmt == nil {
		p.line(depth, "<nil Case>")
		return
	}

	p.line(depth, "Case span=%s", formatSpan(stmt.NodeSpan()))
	p.line(depth+1, "Value:")
	p.printExpr(stmt.Value, depth+2)
	p.printStatements("Body", stmt.Stmts, depth+1)
}

func (p *treePrinter) printWrite(stmt *Write, depth int) {
	if stmt == nil {
		p.line(depth, "<nil Write>")
		return
	}

	p.line(
		depth,
		"Write span=%s",
		formatSpan(stmt.NodeSpan()),
	)

	if len(stmt.Content) == 0 {
		p.line(depth+1, "Content: <empty>")
		return
	}

	p.line(depth+1, "Content:")

	for index, expression := range stmt.Content {
		p.line(depth+2, "[%d]", index)
		p.printExpr(expression, depth+3)
	}
}

func (p *treePrinter) printRead(stmt *Read, depth int) {
	if stmt == nil {
		p.line(depth, "<nil Read>")
		return
	}

	p.line(depth, "Read span=%s", formatSpan(stmt.NodeSpan()))
	p.line(depth+1, "Into:")
	p.printLValue(stmt.Into, depth+2)
}

func (p *treePrinter) printIf(stmt *If, depth int) {
	if stmt == nil {
		p.line(depth, "<nil If>")
		return
	}

	p.line(depth, "If span=%s", formatSpan(stmt.NodeSpan()))
	p.line(depth+1, "Condition:")
	p.printExpr(stmt.Condition, depth+2)
	p.printStatements("Then", stmt.Stmts, depth+1)

	if stmt.Else == nil {
		p.line(depth+1, "Else: <none>")
		return
	}

	p.line(depth+1, "Else:")
	p.printElse(stmt.Else, depth+2)
}

func (p *treePrinter) printElse(stmt *Else, depth int) {
	if stmt == nil {
		p.line(depth, "<nil Else>")
		return
	}

	p.line(depth, "Else span=%s", formatSpan(stmt.NodeSpan()))
	p.printStatements("Body", stmt.Stmts, depth+1)
}

func (p *treePrinter) printClearScreen(stmt *ClearScreen, depth int) {
	if stmt == nil {
		p.line(depth, "<nil ClearScreen>")
		return
	}

	p.line(depth, "ClearScreen span=%s", formatSpan(stmt.NodeSpan()))
}

// -----------------------------------------------------------------------------
// Lvalues
// -----------------------------------------------------------------------------

func (p *treePrinter) printLValue(value LValue, depth int) {
	if value == nil {
		p.line(depth, "<nil lvalue>")
		return
	}

	p.printExpr(value, depth)

	p.line(depth+1, "Mutable: %t", value.IsMutable())

	origin := value.AssignmentOrigin()
	if origin == nil {
		p.line(depth+1, "AssignmentOrigin: <none>")
		return
	}

	p.line(depth+1, "AssignmentOrigin:")
	p.line(depth+2, "Span: %s", formatSpan(origin.Span))
	p.line(depth+2, "Message: %q", origin.Message)
}

// -----------------------------------------------------------------------------
// Expressions
// -----------------------------------------------------------------------------

func (p *treePrinter) printExpr(expression TypedExpr, depth int) {
	if expression == nil {
		p.line(depth, "<nil expression>")
		return
	}

	switch expression := expression.(type) {
	case ErrorExpr:
		p.printErrorExpr(&expression, depth)

	case *ErrorExpr:
		p.printErrorExpr(expression, depth)

	case *StringLiteral:
		p.printStringLiteral(expression, depth)

	case *NumberLiteral:
		p.printNumberLiteral(expression, depth)

	case BooleanLiteral:
		p.printBooleanLiteral(&expression, depth)

	case *BooleanLiteral:
		p.printBooleanLiteral(expression, depth)

	case VariableExpr:
		p.printVariableExpr(&expression, depth)

	case *VariableExpr:
		p.printVariableExpr(expression, depth)

	case UnaryOperation:
		p.printUnaryOperation(&expression, depth)

	case *UnaryOperation:
		p.printUnaryOperation(expression, depth)

	case BinaryOperation:
		p.printBinaryOperation(&expression, depth)

	case *BinaryOperation:
		p.printBinaryOperation(expression, depth)

	case *Cast:
		p.printCast(expression, depth)

	case Cast:
		p.printCast(&expression, depth)

	default:
		p.line(
			depth,
			"UnknownExpression type=%T semanticType=%s span=%s",
			expression,
			formatType(expression.Type()),
			formatSpan(expression.NodeSpan()),
		)
	}
}

func (p *treePrinter) printErrorExpr(expression *ErrorExpr, depth int) {
	if expression == nil {
		p.line(depth, "<nil ErrorExpr>")
		return
	}

	p.line(
		depth,
		"ErrorExpr type=%s span=%s",
		formatType(expression.Type()),
		formatSpan(expression.NodeSpan()),
	)
}

func (p *treePrinter) printStringLiteral(
	expression *StringLiteral,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil StringLiteral>")
		return
	}

	p.line(
		depth,
		"StringLiteral value=%q type=%s span=%s",
		expression.Value,
		formatType(expression.Type()),
		formatSpan(expression.NodeSpan()),
	)
}

func (p *treePrinter) printNumberLiteral(
	expression *NumberLiteral,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil NumberLiteral>")
		return
	}

	p.line(
		depth,
		"NumberLiteral int=%s frac=%s type=%s span=%s",
		expression.Int,
		expression.Frac,
		formatType(expression.Type()),
		formatSpan(expression.NodeSpan()),
	)
}

func (p *treePrinter) printBooleanLiteral(
	expression *BooleanLiteral,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil BooleanLiteral>")
		return
	}

	p.line(
		depth,
		"BooleanLiteral value=%t type=%s span=%s",
		expression.Value,
		formatType(expression.Type()),
		formatSpan(expression.NodeSpan()),
	)
}

func (p *treePrinter) printVariableExpr(
	expression *VariableExpr,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil VariableExpr>")
		return
	}

	if expression.Symbol == nil {
		p.line(
			depth,
			"VariableExpr symbol=<nil> type=<nil> span=%s",
			formatSpan(expression.NodeSpan()),
		)
		return
	}

	symbol := expression.Symbol

	p.line(
		depth,
		"VariableExpr name=%q type=%s span=%s",
		symbol.Name,
		formatType(symbol.Type),
		formatSpan(expression.NodeSpan()),
	)

	p.line(depth+1, "Symbol:")
	p.line(depth+2, "Kind: %s", formatSymbolKind(symbol.Kind))
	p.line(depth+2, "Mutable: %t", symbol.Mutable)
	p.line(depth+2, "Declared: %s", formatSpan(symbol.Declared))
}

func (p *treePrinter) printUnaryOperation(
	expression *UnaryOperation,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil UnaryOperation>")
		return
	}

	p.line(
		depth,
		"UnaryOperation type=%s span=%s",
		formatType(expression.ResolvedType),
		formatSpan(expression.NodeSpan()),
	)

	p.line(depth+1, "Operator:")
	p.printOperator(expression.Op, depth+2)

	p.line(depth+1, "Expression:")
	p.printExpr(expression.Expr, depth+2)
}

func (p *treePrinter) printBinaryOperation(
	expression *BinaryOperation,
	depth int,
) {
	if expression == nil {
		p.line(depth, "<nil BinaryOperation>")
		return
	}

	p.line(
		depth,
		"BinaryOperation type=%s span=%s",
		formatType(expression.ResolvedType),
		formatSpan(expression.NodeSpan()),
	)

	p.line(depth+1, "Left:")
	p.printExpr(expression.LHS, depth+2)

	p.line(depth+1, "Operator:")
	p.printOperator(expression.Op, depth+2)

	p.line(depth+1, "Right:")
	p.printExpr(expression.RHS, depth+2)
}

func (p *treePrinter) printCast(expression *Cast, depth int) {
	if expression == nil {
		p.line(depth, "<nil Cast>")
		return
	}

	p.line(
		depth,
		"Cast targetType=%s conversion=%s span=%s",
		formatType(expression.TargetType),
		formatConversionKind(expression.ConversionKind),
		formatSpan(expression.NodeSpan()),
	)

	p.line(depth+1, "Expression:")
	p.printExpr(expression.Expr, depth+2)
}

func (p *treePrinter) printOperator(operator *Operator, depth int) {
	if operator == nil {
		p.line(depth, "<nil Operator>")
		return
	}

	p.line(
		depth,
		"Operator value=%v span=%s",
		operator.Op,
		formatSpan(operator.NodeSpan()),
	)
}

// -----------------------------------------------------------------------------
// Formatting
// -----------------------------------------------------------------------------

func formatSpan(span models.Span) string {
	return fmt.Sprintf(
		"%d:%d@%d-%d:%d@%d",
		span.Start.Line,
		span.Start.Column,
		span.Start.Offset,
		span.End.Line,
		span.End.Column,
		span.End.Offset,
	)
}

func formatType(value semantic.Type) string {
	if value == nil {
		return "<nil>"
	}

	return fmt.Sprint(value)
}

func formatConversionKind(kind *semantic.ConversionKind) string {
	if kind == nil {
		return "<none>"
	}

	return fmt.Sprint(*kind)
}

func formatSymbolKind(kind semantic.SymbolKind) string {
	switch kind {
	case semantic.SymbolVariable:
		return "Variable"

	case semantic.SymbolParameter:
		return "Parameter"

	case semantic.SymbolFunction:
		return "Function"

	default:
		return fmt.Sprintf("Unknown(%d)", kind)
	}
}
