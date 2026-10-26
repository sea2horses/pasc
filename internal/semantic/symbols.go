package semantic

import (
	"pascc/internal/diagnostics"
	"pascc/internal/models"
	"strings"
)

type SymbolKind uint8

const (
	SymbolVariable SymbolKind = iota
	SymbolParameter
	SymbolFunction
)

type Symbol struct {
	Name     string
	Kind     SymbolKind
	Type     Type
	Declared models.Span
	Mutable  bool
	Builtin  *BuiltinInfo
}

type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: map[string]*Symbol{},
	}
}

/* TODO: Add option to remove case insensitive assignment */
func NormalizeName(name string) string {
	return strings.ToLower(name)
}

func (s *Scope) Parent() *Scope {
	return s.parent
}

func (s *Scope) Declare(symbol *Symbol) bool {
	if symbol == nil {
		panic("GAVE A NIL SYMBOL! CANNOT DECLARE.")
	}

	diagnostics.Dbgfmt("Declared new symbol: %+v\n", symbol)

	key := NormalizeName(symbol.Name)
	if _, exists := s.symbols[key]; exists {
		return false
	}

	s.symbols[key] = symbol
	return true
}

func (s *Scope) LookupLocal(name string) (*Symbol, bool) {
	symbol, ok := s.symbols[NormalizeName(name)]
	return symbol, ok
}

func (s *Scope) Lookup(name string) (*Symbol, bool) {
	for current := s; current != nil; current = current.parent {
		if symbol, ok := current.LookupLocal(name); ok {
			return symbol, true
		}
	}

	return nil, false
}

type TypeTable struct {
	table      map[string]Type
	inferences int
}

var default_table map[string]Type = map[string]Type{
	"entero":   Integer_T{},
	"cadena":   String_T{},
	"caracter": String_T{},
	"texto":    String_T{},
	"logico":   Bool_T{},
	"real":     Real_T{},
	"numero":   Real_T{},
	"numerico": Real_T{},
}

func NewTypeTable() *TypeTable {
	return &TypeTable{
		table: default_table,
	}
}

func (s *TypeTable) MakeInference() Type {
	s.inferences++
	return &InferType{
		ID:       s.inferences,
		Resolved: nil,
	}
}

func (s *TypeTable) Get(name string) Type {
	t, ok := s.table[NormalizeName(name)]
	if !ok {
		return nil
	}
	return t
}
