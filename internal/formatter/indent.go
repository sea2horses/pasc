package formatter

import "pascc/internal/lexer"

type block uint8

const (
	blockMain   block = iota /* Algoritmo / Proceso */
	blockFunc                /* Funcion / SubProceso / SubAlgoritmo */
	blockIf                  /* Si */
	blockWhile               /* Mientras */
	blockFor                 /* Para, Para Cada */
	blockSwitch              /* Segun */
	blockCase                /* the body of one Segun case */
	blockRepeat              /* Repetir ... Hasta Que / Mientras Que */
	blockDo                  /* Hacer ... Mientras Que */
)

// closers maps each compound end keyword (and the keyword that follows a
// separate "Fin") to the block it closes.
var closers = map[lexer.Keyword]block{
	lexer.FINALGORITMO: blockMain, lexer.FINPROCESO: blockMain,
	lexer.FINFUNCION: blockFunc, lexer.FINSUBPROCESO: blockFunc,
	lexer.FINSI: blockIf, lexer.FINMIENTRAS: blockWhile,
	lexer.FINPARA: blockFor, lexer.FINSEGUN: blockSwitch,
}

var finTails = map[lexer.Keyword]block{
	lexer.ALGORITMO: blockMain, lexer.PROCESO: blockMain,
	lexer.FUNCION: blockFunc, lexer.SUBPROCESO: blockFunc,
	lexer.SI: blockIf, lexer.MIENTRAS: blockWhile,
	lexer.PARA: blockFor, lexer.SEGUN: blockSwitch,
}

var openers = map[lexer.Keyword]block{
	lexer.ALGORITMO: blockMain, lexer.PROCESO: blockMain,
	lexer.FUNCION: blockFunc, lexer.SUBPROCESO: blockFunc,
	lexer.SI: blockIf, lexer.PARA: blockFor,
	lexer.SEGUN: blockSwitch, lexer.REPETIR: blockRepeat,
}

// indenter tracks the open blocks while lines are rendered in order.
type indenter struct {
	stack []block
}

func (in *indenter) top() (block, bool) {
	if len(in.stack) == 0 {
		return 0, false
	}
	return in.stack[len(in.stack)-1], true
}

func (in *indenter) topIs(blocks ...block) bool {
	top, ok := in.top()
	if !ok {
		return false
	}
	for _, b := range blocks {
		if top == b {
			return true
		}
	}
	return false
}

// close pops b and everything opened after it. Closing a Segun therefore
// also closes its last case. A block that is not open is ignored.
func (in *indenter) close(b block) {
	for i := len(in.stack) - 1; i >= 0; i-- {
		if in.stack[i] == b {
			in.stack = in.stack[:i]
			return
		}
	}
}

// line updates the open blocks with one line and returns its indent level.
//
// A line is indented at the depth in effect when it starts, except that
// closers and SiNo at the start of a line move it out: "FinSi" lines up with
// its "Si". Segun cases sit one level inside the Segun and their bodies one
// level further.
func (in *indenter) line(glyphs []glyph) int {
	level := len(in.stack)
	leading := true /* only closers have been seen so far */
	stmtStart := true

	next := func(i int) *glyph {
		for j := i + 1; j < len(glyphs); j++ {
			if glyphs[j].kind != lexer.COMMENT {
				return &glyphs[j]
			}
		}
		return nil
	}

	for i := 0; i < len(glyphs); i++ {
		g := &glyphs[i]
		if g.kind == lexer.COMMENT {
			continue
		}

		if leading && in.topIs(blockSwitch, blockCase) {
			if end, ok := caseHeader(glyphs, i); ok {
				if in.topIs(blockCase) {
					in.stack = in.stack[:len(in.stack)-1]
				}
				level = len(in.stack)
				in.stack = append(in.stack, blockCase)
				leading, stmtStart, i = false, true, end
				continue
			}
		}

		closed := false
		if g.isKw {
			switch kw := g.kw; {
			case kw == lexer.FIN:
				if tail := next(i); tail != nil && tail.isKw {
					if b, ok := finTails[tail.kw]; ok {
						in.close(b)
						closed = true
						for glyphs[i+1].kind == lexer.COMMENT {
							i++
						}
						i++
					}
				}
			case kw == lexer.SINO:
				if leading {
					level = len(in.stack) - 1
				}
			case kw == lexer.MIENTRAS:
				if in.topIs(blockRepeat, blockDo) && next(i).is(lexer.QUE) && !hasKeywordAfter(glyphs, i, lexer.HACER) {
					// "Mientras Que" closing a Repetir or Hacer loop, the same
					// test the parser applies.
					in.stack = in.stack[:len(in.stack)-1]
					closed = true
				} else {
					in.stack = append(in.stack, blockWhile)
				}
			case kw == lexer.HASTA:
				if in.topIs(blockRepeat) && next(i).is(lexer.QUE) {
					in.close(blockRepeat)
					closed = true
				}
			case kw == lexer.HACER:
				// Only a Hacer that starts a statement opens a loop; elsewhere it
				// is the optional end of a Mientras, Para or Segun header.
				if stmtStart {
					in.stack = append(in.stack, blockDo)
				}
			default:
				if b, ok := closers[kw]; ok {
					in.close(b)
					closed = true
				} else if b, ok := openers[kw]; ok {
					in.stack = append(in.stack, b)
				}
			}
		}

		if closed && leading {
			level = len(in.stack)
		} else if !closed {
			leading = false
		}
		stmtStart = g.text == ";"
	}

	return max(level, 0)
}

// caseHeader recognizes "Caso 1, 2:", "1, 2:" and "De Otro Modo[:]" starting
// at glyph i, and returns the index of the header's last glyph.
func caseHeader(glyphs []glyph, i int) (int, bool) {
	if glyphs[i].is(lexer.DE) {
		for j := i + 1; j < len(glyphs); j++ {
			if glyphs[j].is(lexer.MODO) {
				if j+1 < len(glyphs) && glyphs[j+1].text == ":" {
					j++
				}
				return j, true
			}
		}
		return 0, false
	}
	for j := i; j < len(glyphs) && glyphs[j].text != ";"; j++ {
		if glyphs[j].kind == lexer.COLON && glyphs[j].text == ":" {
			return j, true
		}
	}
	return 0, false
}

func hasKeywordAfter(glyphs []glyph, i int, kw lexer.Keyword) bool {
	for j := i + 1; j < len(glyphs); j++ {
		if glyphs[j].is(kw) {
			return true
		}
	}
	return false
}
