package analyzer

import (
	"errors"
	"pascc/internal/i18n"
	"strings"

	tast "pascc/internal/ast_typed"
	"pascc/internal/diagnostics"
)

// IndexBase selects the first valid array index of a program. PSeInt lets the
// profile decide between 0-based and 1-based arrays, so the compiler either
// takes it from the user or deduces it from how the program indexes arrays.
type IndexBase int

const (
	IndexBaseAuto IndexBase = iota
	IndexBaseZero
	IndexBaseOne
)

func (b IndexBase) String() string {
	switch b {
	case IndexBaseZero:
		return "0"
	case IndexBaseOne:
		return "1"
	default:
		return "auto"
	}
}

// Offset is the numeric first index. Auto has no offset until it is resolved.
func (b IndexBase) Offset() int {
	if b == IndexBaseOne {
		return 1
	}
	return 0
}

func ParseIndexBase(value string) (IndexBase, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return IndexBaseAuto, nil
	case "0", "zero":
		return IndexBaseZero, nil
	case "1", "one":
		return IndexBaseOne, nil
	}
	return IndexBaseAuto, errors.New(i18n.T(i18n.InvalidIndexBase, value))
}

// resolveIndexBase settles the base used for code generation and runs the
// compile-time checks that depend on it.
// Without a requested base it is deduced from the program's subscripts;
// programs with no evidence either way keep the 0-based default.
func (a *Analyzer) resolveIndexBase(main *tast.MainFunction) IndexBase {
	// A program with errors is never generated, but checking it anyway lets
	// bounds and dimension errors show up next to the type errors instead of
	// only after those are fixed. The pass tolerates holes in the tree; the
	// recover is a last resort so a missed hole costs diagnostics, not a crash.
	if a.Checkpoint().AnyError() {
		defer func() { _ = recover() }()
	}
	pass := newStaticPass(a, main)
	base := a.IndexBase
	if base == IndexBaseAuto {
		base = pass.deduceBase()
		diagnostics.Dbgfmt("deduced %s-based arrays (evidence: 0-based=%d, 1-based=%d)\n", base, pass.zero.score, pass.one.score)
	}
	pass.checkBounds(base)
	return base
}
