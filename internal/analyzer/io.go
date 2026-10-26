package analyzer

import (
	"pascc/internal/ast"
	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
	"pascc/internal/semantic"
	"slices"
)

var allowed_write_types []semantic.Type = []semantic.Type{
	semantic.Integer_T{},
	semantic.Bool_T{},
	semantic.Real_T{},
	semantic.String_T{},
}

func (a *Analyzer) analyze_write(write *ast.Write) *tast.Write {
	diagnostics.Dbg("Analyzing write...")
	content := []tast.TypedExpr{}
	ok := true

	for _, exp := range write.Print {
		typed := defaultUntypedNumber(a.analyze_expression(exp))

		if semantic.IsInvalid(typed.Type()) {
			ok = false
			continue
		}
		/* Write cannot have a void value */
		n := a.assert_nonvoid(typed)
		if !n {
			ok = false
			continue
		}
		/* Check that the type is supported */
		if !slices.Contains(allowed_write_types, semantic.Resolve(typed.Type())) {
			a.Report(typed.NodeSpan(), "%s", ErrUnsupportedType{Type: typed.Type()}.Error())
			ok = false
			continue
		}

		content = append(content, typed)
	}

	if !ok {
		return nil
	}

	return &tast.Write{NodeInfo: write.NodeInfo, Content: content, Newline: write.Newline}
}

func (a *Analyzer) analyze_read(read *ast.Read) []*tast.Read {
	diagnostics.Dbg("analyzing read")
	reads := []*tast.Read{}
	ok := true

	for _, exp := range read.Into {
		/* Check each expression */
		typed := a.analyze_expression(exp)

		if semantic.IsInvalid(typed.Type()) {
			ok = false
			continue
		}

		/* Assert it to lvalue */
		lvalue, good /* Im running out of names */ := typed.(tast.LValue)
		if !good {
			a.Report(typed.NodeSpan(), ErrLValueError)
			ok = false
			continue
		}

		/* Check that it is mutable */
		if !lvalue.IsMutable() {
			a.Report(typed.NodeSpan(), ErrImmutable)
			a.Info(lvalue.AssignmentOrigin().Span, "%s", lvalue.AssignmentOrigin().Message)
			continue
		}

		/* All checks pass, add it */
		reads = append(reads, &tast.Read{NodeInfo: read.NodeInfo, Into: lvalue})
	}

	if !ok {
		return nil
	}

	return reads
}

func (a *Analyzer) analyze_clear_screen(cls *ast.ClearScreen) *tast.ClearScreen {
	return &tast.ClearScreen{NodeInfo: cls.NodeInfo}
}
