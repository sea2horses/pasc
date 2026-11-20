package analyzer

import (
	"math"
	"pascc/internal/i18n"
	"reflect"
	"strconv"

	tast "pascc/internal/ast_typed"
	"pascc/internal/models"
	"pascc/internal/semantic"
)

// The static pass runs over the finished typed tree. It abstractly evaluates
// numeric expressions (constants, ranges of loop variables, and symbolic
// bounds such as "n" or "n - 1") to gather evidence about the array base the
// program was written for.

// linear is a symbolic bound "sym + c"; a nil sym means the constant c.
type linear struct {
	sym *semantic.Symbol
	c   float64
}

// absval is what the static pass knows about a numeric expression.
type absval struct {
	known    bool    // lo/hi hold every value the expression can take
	lo, hi   float64 // numeric range when known
	exact    bool    // a single compile-time constant (lo == hi)
	integral bool    // always a whole number
	ends     []linear
	shifted  bool // ends were offset from a loop variable (e.g. i + 1)
	// lower/upper are the first and last values of a loop variable (possibly
	// offset), when they can be named. They drive the index-base evidence.
	lower, upper *linear
}

func shiftBound(b *linear, by float64) *linear {
	if b == nil {
		return nil
	}
	return &linear{sym: b.sym, c: b.c + by}
}

func point(v float64) absval {
	return absval{known: true, lo: v, hi: v, exact: true, integral: v == math.Trunc(v), ends: []linear{{c: v}}}
}

// pointEnd returns the single symbolic value an expression stands for, if any.
func (v absval) pointEnd() (linear, bool) {
	if len(v.ends) == 1 && (v.exact || !v.known) {
		return v.ends[0], true
	}
	return linear{}, false
}

func shiftEnds(ends []linear, by float64) []linear {
	out := make([]linear, len(ends))
	for i, e := range ends {
		out[i] = linear{sym: e.sym, c: e.c + by}
	}
	return out
}

// indexAccess records one array subscript for evidence and bounds checks.
type indexAccess struct {
	span  models.Span
	array *semantic.Symbol
	axis  int
	value absval
	dim   *absval
	// guarded accesses sit under a condition that mentions one of the index's
	// variables (Si i > 0 Entonces v[i-1]); their loop range is not reliable.
	guarded bool
}

type evidence struct {
	score int
	span  *models.Span
}

type staticPass struct {
	a        *Analyzer
	assigns  map[*semantic.Symbol]int
	env      map[*semantic.Symbol]absval
	dims     map[*semantic.Symbol][]absval
	accesses []indexAccess
	zero     evidence
	one      evidence
	guards   []map[*semantic.Symbol]bool
	// resizes is set when the program uses Redimensionar anywhere; arrays
	// passed to user functions may then come back with a different size.
	resizes bool
}

func newStaticPass(a *Analyzer, main *tast.MainFunction) *staticPass {
	p := &staticPass{
		a:       a,
		assigns: map[*semantic.Symbol]int{},
		env:     map[*semantic.Symbol]absval{},
		dims:    map[*semantic.Symbol][]absval{},
	}
	p.countAssignments(main.Stmts)
	for _, function := range main.Functions {
		p.countAssignments(function.Stmts)
	}
	p.block(main.Stmts, true)
	for _, function := range main.Functions {
		p.block(function.Stmts, false)
		if function.Return != nil {
			p.eval(function.Return)
		}
	}
	return p
}

/* Assignment counting: a variable written exactly once can be folded. */

func (p *staticPass) countAssignments(stmts []tast.Stmt) {
	walkStmts(stmts, func(stmt tast.Stmt) {
		switch node := stmt.(type) {
		case *tast.Assignment:
			p.countWrite(node.Target)
		case *tast.Read:
			p.countWrite(node.Into)
		case *tast.Dimension:
			if node.Resize {
				p.resizes = true
			}
		}
	}, func(expr tast.TypedExpr) {
		call, ok := expr.(*tast.Call)
		if !ok {
			return
		}
		for i, arg := range call.Arguments {
			if i < len(call.ByReference) && call.ByReference[i] {
				p.countWrite(arg)
			}
		}
	})
}

func (p *staticPass) countWrite(target tast.TypedExpr) {
	if variable, ok := target.(*tast.VariableExpr); ok {
		p.assigns[variable.Symbol]++
	}
}

/* Statements */

func (p *staticPass) block(stmts []tast.Stmt, topLevel bool) {
	for _, stmt := range stmts {
		p.stmt(stmt, topLevel)
	}
}

func (p *staticPass) stmt(stmt tast.Stmt, topLevel bool) {
	if isNil(stmt) {
		return
	}
	switch node := stmt.(type) {
	case *tast.Assignment:
		value := p.eval(node.Value)
		p.eval(node.Target)
		// Only straight-line code in the main body runs unconditionally and in
		// order, so a single write there is a constant for everything after it.
		if variable, ok := node.Target.(*tast.VariableExpr); ok && topLevel && value.exact && p.assigns[variable.Symbol] == 1 {
			p.env[variable.Symbol] = value
		}
	case *tast.Read:
		p.eval(node.Into)
	case *tast.Write:
		for _, expr := range node.Content {
			p.eval(expr)
		}
	case *tast.ExprStmt:
		p.eval(node.Expr)
	case *tast.TimeOut:
		p.eval(node.Amount)
	case *tast.If:
		p.eval(node.Condition)
		p.guarded(func() {
			p.block(node.Stmts, false)
			if node.Else != nil {
				p.block(node.Else.Stmts, false)
			}
		}, node.Condition)
	case *tast.While:
		p.eval(node.Condition)
		p.guarded(func() { p.block(node.Stmts, false) }, node.Condition)
	case *tast.DoWhile:
		p.guarded(func() { p.block(node.Stmts, false) }, node.Condition)
		p.eval(node.Condition)
	case *tast.Switch:
		p.eval(node.Base)
		p.guarded(func() {
			for _, c := range node.Cases {
				for _, value := range c.Values {
					p.eval(value)
				}
				p.block(c.Stmts, false)
			}
			p.block(node.Default, false)
		}, node.Base)
	case *tast.For:
		p.forLoop(node)
	case *tast.ForEach:
		p.eval(node.Array)
		p.block(node.Stmts, false)
	case *tast.Dimension:
		dims := make([]absval, len(node.Dimensions))
		for i, dim := range node.Dimensions {
			dims[i] = p.eval(dim)
		}
		p.checkDimension(node, dims)
		switch {
		case !node.Resize || topLevel:
			p.dims[node.Symbol] = dims
		default:
			// A conditional resize leaves the size unknown from here on.
			delete(p.dims, node.Symbol)
		}
	}
}

func (p *staticPass) forLoop(f *tast.For) {
	start := p.eval(f.Start)
	until := p.eval(f.Until)
	var step *absval
	if f.Step != nil {
		s := p.eval(f.Step)
		step = &s
	}

	loop := absval{integral: start.integral && (step == nil || step.integral)}
	unitStep := step == nil || (step.exact && math.Abs(step.lo) == 1)

	switch {
	case start.exact && until.exact && (step == nil || step.exact):
		s, e := start.lo, until.lo
		st := 1.0
		if step != nil {
			st = step.lo
		} else if e < s {
			st = -1
		}
		if st == 0 || (st > 0 && s > e) || (st < 0 && s < e) {
			// The body never runs: its accesses say nothing about the program.
			loop.ends = nil
			break
		}
		last := s + math.Floor((e-s)/st)*st
		loop.known = true
		loop.lo, loop.hi = math.Min(s, last), math.Max(s, last)
		loop.ends = []linear{{c: s}, {c: last}}
		loop.lower, loop.upper = &linear{c: loop.lo}, &linear{c: loop.hi}
	default:
		var first, final *linear
		if end, ok := start.pointEnd(); ok {
			loop.ends = append(loop.ends, end)
			first = &end
		}
		// The last value only matches the bound when every value is visited.
		if end, ok := until.pointEnd(); ok && unitStep {
			loop.ends = append(loop.ends, end)
			final = &end
		}
		if step != nil && step.exact && step.lo < 0 {
			loop.lower, loop.upper = final, first
		} else {
			loop.lower, loop.upper = first, final
		}
	}

	var symbol *semantic.Symbol
	if f.Var != nil {
		symbol = f.Var.Symbol
	}
	if symbol != nil && p.assigns[symbol] == 0 {
		p.env[symbol] = loop
		defer delete(p.env, symbol)
	}
	p.block(f.Stmts, false)
}

/* Expressions */

func (p *staticPass) eval(expr tast.TypedExpr) absval {
	if isNil(expr) {
		return absval{}
	}
	v := p.evalNode(expr)
	if t := expr.Type(); t != nil && semantic.IsInteger(semantic.Resolve(t)) {
		v.integral = true
	}
	return v
}

func (p *staticPass) evalNode(expr tast.TypedExpr) absval {
	switch node := expr.(type) {
	case *tast.NumberLiteral:
		text := node.Int
		if node.Frac != "" {
			text += "." + node.Frac
		}
		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return absval{}
		}
		return point(value)
	case *tast.VariableExpr:
		if node.Symbol == nil {
			return absval{}
		}
		if value, ok := p.env[node.Symbol]; ok {
			return value
		}
		if semantic.IsNumeric(semantic.Resolve(node.Symbol.Type)) {
			return absval{ends: []linear{{sym: node.Symbol}}}
		}
		return absval{}
	case *tast.Cast:
		inner := p.eval(node.Expr)
		if node.ConversionKind != nil && *node.ConversionKind == semantic.ConversionRealToIntegerExact {
			p.checkRealToInteger(node, inner)
		}
		return inner
	case *tast.UnaryOperation:
		inner := p.eval(node.Expr)
		if node.Op == nil || node.Op.Op != semantic.SUBTRACT {
			return absval{}
		}
		out := absval{known: inner.known, lo: -inner.hi, hi: -inner.lo, exact: inner.exact, integral: inner.integral}
		for _, e := range inner.ends {
			if e.sym == nil {
				out.ends = append(out.ends, linear{c: -e.c})
			}
		}
		return out
	case *tast.BinaryOperation:
		return p.binary(node)
	case *tast.Index:
		p.index(node)
		return absval{}
	case *tast.Call:
		args := make([]absval, len(node.Arguments))
		for i, arg := range node.Arguments {
			args[i] = p.eval(arg)
		}
		if p.resizes && !isBuiltinCall(node) {
			for _, arg := range node.Arguments {
				if variable, ok := arg.(*tast.VariableExpr); ok {
					delete(p.dims, variable.Symbol)
				}
			}
		}
		return builtinValue(node, args)
	}
	return absval{}
}

func (p *staticPass) binary(node *tast.BinaryOperation) absval {
	l, r := p.eval(node.LHS), p.eval(node.RHS)
	if node.Op == nil {
		return absval{}
	}
	switch node.Op.Op {
	case semantic.ADD, semantic.SUBTRACT:
		sign := 1.0
		if node.Op.Op == semantic.SUBTRACT {
			sign = -1
		}
		out := absval{integral: l.integral && r.integral, exact: l.exact && r.exact}
		if l.known && r.known {
			out.known = true
			if sign > 0 {
				out.lo, out.hi = l.lo+r.lo, l.hi+r.hi
			} else {
				out.lo, out.hi = l.lo-r.hi, l.hi-r.lo
			}
		}
		switch {
		case out.exact:
			out.ends = []linear{{c: out.lo}}
		case r.exact:
			out.ends = shiftEnds(l.ends, sign*r.lo)
			out.shifted = l.shifted || r.lo != 0
			out.lower, out.upper = shiftBound(l.lower, sign*r.lo), shiftBound(l.upper, sign*r.lo)
		case l.exact && sign > 0:
			out.ends = shiftEnds(r.ends, l.lo)
			out.shifted = r.shifted || l.lo != 0
			out.lower, out.upper = shiftBound(r.lower, l.lo), shiftBound(r.upper, l.lo)
		}
		return out
	case semantic.MULTIPLY:
		out := absval{integral: l.integral && r.integral}
		if l.known && r.known {
			products := []float64{l.lo * r.lo, l.lo * r.hi, l.hi * r.lo, l.hi * r.hi}
			out.known, out.lo, out.hi = true, slicesMin(products), slicesMax(products)
			if l.exact && r.exact {
				out.exact = true
				out.ends = []linear{{c: out.lo}}
			}
		}
		return out
	case semantic.DIVIDE:
		if l.exact && r.exact && r.lo != 0 {
			return point(l.lo / r.lo)
		}
	case semantic.MODULO:
		if l.exact && r.exact && r.lo != 0 {
			return point(math.Mod(l.lo, r.lo))
		}
		return absval{integral: l.integral && r.integral}
	case semantic.POWER:
		if l.exact && r.exact {
			return point(math.Pow(l.lo, r.lo))
		}
	}
	return absval{}
}

func isBuiltinCall(call *tast.Call) bool {
	variable, ok := call.Callable.(*tast.VariableExpr)
	return ok && variable.Symbol.Builtin != nil
}

func builtinValue(call *tast.Call, args []absval) absval {
	variable, ok := call.Callable.(*tast.VariableExpr)
	if !ok || variable.Symbol.Builtin == nil || len(args) != 1 {
		return absval{}
	}
	arg := args[0]
	switch variable.Symbol.Builtin.Kind {
	case semantic.BuiltinTrunc:
		if arg.exact {
			return point(math.Trunc(arg.lo))
		}
		return absval{integral: true}
	case semantic.BuiltinRound:
		if arg.exact {
			return point(math.Round(arg.lo))
		}
		return absval{integral: true}
	case semantic.BuiltinAbs:
		if arg.exact {
			return point(math.Abs(arg.lo))
		}
		return absval{integral: arg.integral}
	}
	return absval{}
}

func (p *staticPass) index(node *tast.Index) {
	var array *semantic.Symbol
	if variable, ok := node.Target.(*tast.VariableExpr); ok && variable != nil {
		array = variable.Symbol
	} else {
		p.eval(node.Target)
	}
	dims := p.dims[array]
	for axis, idx := range node.Indexes {
		access := indexAccess{span: spanOf(idx), array: array, axis: axis, value: p.eval(idx), guarded: p.isGuarded(idx)}
		if array != nil && axis < len(dims) {
			access.dim = &dims[axis]
		}
		p.accesses = append(p.accesses, access)
	}
}

// guarded runs body with the variables of condition marked as guards.
func (p *staticPass) guarded(body func(), condition tast.TypedExpr) {
	vars := map[*semantic.Symbol]bool{}
	walkExpr(condition, func(e tast.TypedExpr) {
		if v, ok := e.(*tast.VariableExpr); ok {
			vars[v.Symbol] = true
		}
	})
	p.guards = append(p.guards, vars)
	defer func() { p.guards = p.guards[:len(p.guards)-1] }()
	body()
}

func (p *staticPass) isGuarded(index tast.TypedExpr) bool {
	guarded := false
	walkExpr(index, func(e tast.TypedExpr) {
		v, ok := e.(*tast.VariableExpr)
		if !ok {
			return
		}
		for _, guard := range p.guards {
			if guard[v.Symbol] {
				guarded = true
			}
		}
	})
	return guarded
}

/* Compile-time safety checks */

// checkRealToInteger validates an implicit real -> integer conversion. The
// runtime rejects non-whole values, so a constant fraction is an error and a
// value that cannot be proven whole is a warning pointing at Redon()/Trunc().
func (p *staticPass) checkRealToInteger(cast *tast.Cast, value absval) {
	span := spanOf(cast)
	switch {
	case value.exact && !value.integral:
		p.a.errorf(span, i18n.RealToInteger, formatNumber(value.lo))
	case !value.integral:
		p.a.warnf(span, i18n.RealToIntegerRuntime)
	}
}

// checkBounds validates dimensions and subscripts whose values are known at
// compile time against the array base the program uses.
func (p *staticPass) checkBounds(base IndexBase) {
	for _, access := range p.accesses {
		if access.dim == nil || !access.dim.exact || !access.value.known || access.dim.lo < 1 {
			continue
		}
		first := float64(base.Offset())
		last := access.dim.lo - 1 + first
		value := access.value
		if value.lo >= first && value.hi <= last {
			continue
		}
		name := access.array.Name
		if !value.exact && access.guarded {
			continue
		}
		if value.exact {
			p.a.errorf(access.span, i18n.IndexOutOfBounds,
				formatNumber(value.lo), name, formatNumber(first), formatNumber(last), base)
			continue
		}
		p.a.warnf(access.span, i18n.IndexMaybeOutOfBounds,
			formatNumber(value.lo), formatNumber(value.hi), name, formatNumber(first), formatNumber(last), base)
	}
}

func (p *staticPass) checkDimension(node *tast.Dimension, dims []absval) {
	for i, dim := range dims {
		if dim.exact && dim.lo < 1 {
			p.a.errorf(spanOf(node.Dimensions[i]), i18n.DimensionAtLeastOne, formatNumber(dim.lo))
		}
	}
}

func formatNumber(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

/* Index base deduction */

func (e *evidence) add(score int, span models.Span) {
	e.score += score
	if e.span == nil {
		e.span = &span
	}
}

// collectEvidence scores how each subscript reads under both bases:
//   - an index that can be 0 only makes sense 0-based;
//   - an index that reaches the dimension size only makes sense 1-based;
//   - a loop whose last index is "size - 1" points to 0-based, and a loop
//     whose first index is 1 weakly points to 1-based.
func (p *staticPass) collectEvidence() {
	for _, access := range p.accesses {
		value := access.value
		matchesSize := func(b linear, offset float64) bool {
			if access.dim == nil {
				return false
			}
			for _, size := range access.dim.ends {
				if size.sym == b.sym && b.c == size.c+offset {
					return true
				}
			}
			return false
		}
		switch {
		case value.exact:
			if value.lo == 0 {
				p.zero.add(2, access.span)
			}
			if matchesSize(linear{c: value.lo}, 0) {
				p.one.add(2, access.span)
			}
		case value.lower != nil || value.upper != nil:
			if access.guarded {
				continue
			}
			if lower := value.lower; lower != nil && lower.sym == nil {
				if lower.c == 0 {
					p.zero.add(2, access.span)
				}
				if lower.c == 1 && !value.shifted {
					p.one.add(1, access.span)
				}
			}
			if upper := value.upper; upper != nil {
				if matchesSize(*upper, 0) {
					p.one.add(2, access.span)
				} else if matchesSize(*upper, -1) {
					p.zero.add(2, access.span)
				}
			}
		default:
			// A plain variable equal to the size, as in v[n] after Dimensionar v[n].
			for _, end := range value.ends {
				if matchesSize(end, 0) {
					p.one.add(2, access.span)
				}
			}
		}
	}
}

func (p *staticPass) deduceBase() IndexBase {
	p.collectEvidence()
	base, other := IndexBaseZero, p.one
	if p.one.score > p.zero.score {
		base, other = IndexBaseOne, p.zero
	}
	if other.score > 0 && other.span != nil {
		p.a.warnf(*other.span, i18n.MixedIndexBase, base.flip(), base)
	}
	return base
}

func (b IndexBase) flip() IndexBase {
	if b == IndexBaseOne {
		return IndexBaseZero
	}
	return IndexBaseOne
}

/* Helpers */

func spanOf(expr tast.TypedExpr) models.Span {
	for {
		cast, ok := expr.(*tast.Cast)
		if !ok {
			return expr.NodeSpan()
		}
		expr = cast.Expr
	}
}

func slicesMin(values []float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		m = math.Min(m, v)
	}
	return m
}

func slicesMax(values []float64) float64 {
	m := values[0]
	for _, v := range values[1:] {
		m = math.Max(m, v)
	}
	return m
}

// walkStmts visits every statement and expression in a block.
func walkStmts(stmts []tast.Stmt, onStmt func(tast.Stmt), onExpr func(tast.TypedExpr)) {
	for _, stmt := range stmts {
		walkStmt(stmt, onStmt, onExpr)
	}
}

func walkStmt(stmt tast.Stmt, onStmt func(tast.Stmt), onExpr func(tast.TypedExpr)) {
	if isNil(stmt) {
		return
	}
	onStmt(stmt)
	expr := func(e tast.TypedExpr) { walkExpr(e, onExpr) }
	switch node := stmt.(type) {
	case *tast.Assignment:
		expr(node.Target)
		expr(node.Value)
	case *tast.Read:
		expr(node.Into)
	case *tast.Write:
		for _, e := range node.Content {
			expr(e)
		}
	case *tast.ExprStmt:
		expr(node.Expr)
	case *tast.TimeOut:
		expr(node.Amount)
	case *tast.Dimension:
		for _, e := range node.Dimensions {
			expr(e)
		}
	case *tast.If:
		expr(node.Condition)
		walkStmts(node.Stmts, onStmt, onExpr)
		if node.Else != nil {
			walkStmts(node.Else.Stmts, onStmt, onExpr)
		}
	case *tast.While:
		expr(node.Condition)
		walkStmts(node.Stmts, onStmt, onExpr)
	case *tast.DoWhile:
		expr(node.Condition)
		walkStmts(node.Stmts, onStmt, onExpr)
	case *tast.ForEach:
		expr(node.Array)
		walkStmts(node.Stmts, onStmt, onExpr)
	case *tast.For:
		expr(node.Start)
		expr(node.Until)
		expr(node.Step)
		walkStmts(node.Stmts, onStmt, onExpr)
	case *tast.Switch:
		expr(node.Base)
		for _, c := range node.Cases {
			for _, e := range c.Values {
				expr(e)
			}
			walkStmts(c.Stmts, onStmt, onExpr)
		}
		walkStmts(node.Default, onStmt, onExpr)
	}
}

func walkExpr(expr tast.TypedExpr, onExpr func(tast.TypedExpr)) {
	if isNil(expr) {
		return
	}
	onExpr(expr)
	switch node := expr.(type) {
	case *tast.Cast:
		walkExpr(node.Expr, onExpr)
	case *tast.UnaryOperation:
		walkExpr(node.Expr, onExpr)
	case *tast.BinaryOperation:
		walkExpr(node.LHS, onExpr)
		walkExpr(node.RHS, onExpr)
	case *tast.Index:
		walkExpr(node.Target, onExpr)
		for _, e := range node.Indexes {
			walkExpr(e, onExpr)
		}
	case *tast.Call:
		walkExpr(node.Callable, onExpr)
		for _, e := range node.Arguments {
			walkExpr(e, onExpr)
		}
	}
}

// isNil reports nil interfaces and typed nil pointers. Programs with errors
// keep holes in their typed tree (e.g. a (*tast.For)(nil) statement), and the
// static pass still walks them to report every problem in one run.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}
