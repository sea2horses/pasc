package parser

type MatchKind uint8

const (
	NoMatch MatchKind = iota
	Matched
	Failed
)

type Match[T any] struct {
	Kind  MatchKind
	Value T

	Start uint32
	End   uint32

	Err ParseError
}

func (m Match[T]) OK() bool {
	return m.Kind == Matched
}

func (m Match[T]) Consumed() bool {
	return m.End > m.Start
}
