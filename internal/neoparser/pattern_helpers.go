package neoparser

import "pseint-compiled/internal/models"

func Map[A, B any](p Pattern[A], fn func(A) B) Pattern[B] {
	return PatternFunc[B](func(ctx *Context) Match[B] {
		result := p.Match(ctx)

		if result.Kind != Matched {
			return Match[B]{
				Kind:  result.Kind,
				Start: result.Start,
				End:   result.End,
				Err:   result.Err,
			}
		}

		return Match[B]{
			Kind:  Matched,
			Value: fn(result.Value),
			Start: result.Start,
			End:   result.End,
		}
	})
}

func MapWithLocation[A, B any](
	p Pattern[A],
	fn func(A, models.Span) B,
) Pattern[B] {
	return Map(
		Locate(p),
		func(l Located[A]) B {
			return fn(l.Value, l.Span)
		},
	)
}

func If[A, B any](
	p Pattern[A],
	val B,
) Pattern[B] {
	return PatternFunc[B](func(ctx *Context) Match[B] {
		result := p.Match(ctx)

		if result.Kind != Matched {
			return Match[B]{
				Kind:  result.Kind,
				Start: result.Start,
				End:   result.End,
				Err:   result.Err,
			}
		}

		return Match[B]{
			Kind:  Matched,
			Value: val,
			Start: result.Start,
			End:   result.End,
		}
	})
}

type OptionalValue[T any] struct {
	Value T
	Some  bool
}

func (o OptionalValue[T]) Or(def T) T {
	if o.Some {
		return o.Value
	} else {
		return def
	}
}

func Optional[T any](p Pattern[T]) Pattern[OptionalValue[T]] {
	return PatternFunc[OptionalValue[T]](func(ctx *Context) Match[OptionalValue[T]] {
		start := ctx.Pos
		result := p.Match(ctx)

		switch result.Kind {
		case Matched:
			return Match[OptionalValue[T]]{
				Kind: Matched,
				Value: OptionalValue[T]{
					Value: result.Value,
					Some:  true,
				},
				Start: result.Start,
				End:   result.End,
			}
		case Failed:
			return Match[OptionalValue[T]]{
				Kind:  Failed,
				Start: result.Start,
				End:   result.End,
				Err:   result.Err,
			}
		default:
			ctx.Reset(start)

			return Match[OptionalValue[T]]{
				Kind: Matched,
				Value: OptionalValue[T]{
					Some: false,
				},
				Start: start,
				End:   start,
			}
		}
	})
}

func AsAny[T any](p Pattern[T]) Pattern[any] {
	return Map(p, func(value T) any {
		return value
	})
}

func Many[T any](p Pattern[T]) Pattern[[]T] {
	return PatternFunc[[]T](func(ctx *Context) Match[[]T] {
		start := ctx.Pos
		values := []T{}

		for {
			before := ctx.Pos
			result := p.Match(ctx)

			if result.Kind == Failed {
				return Match[[]T]{
					Kind:  Failed,
					Start: start,
					End:   result.End,
					Err:   result.Err,
				}
			}

			if result.Kind == NoMatch {
				ctx.Reset(before)
				break
			}

			if ctx.Pos == before {
				panic("Many() received a pattern that matched without consuming input")
			}

			values = append(values, result.Value)
		}

		return Match[[]T]{
			Kind:  Matched,
			Value: values,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func Many1[T any](p Pattern[T]) Pattern[[]T] {
	return PatternFunc[[]T](func(ctx *Context) Match[[]T] {
		start := ctx.Pos

		first := p.Match(ctx)

		if first.Kind != Matched {
			ctx.Reset(start)

			return Match[[]T]{
				Kind:  first.Kind,
				Start: start,
				End:   first.End,
				Err:   first.Err,
			}
		}

		values := []T{first.Value}

		rest := Many(p).Match(ctx)

		if rest.Kind == Failed {
			return rest
		}

		values = append(values, rest.Value...)

		return Match[[]T]{
			Kind:  Matched,
			Value: values,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func SepBy[T, S any](
	item Pattern[T],
	separator Pattern[S],
) Pattern[[]T] {
	return PatternFunc[[]T](func(ctx *Context) Match[[]T] {
		start := ctx.Pos
		values := []T{}

		first := item.Match(ctx)

		if first.Kind == NoMatch {
			ctx.Reset(start)

			return Match[[]T]{
				Kind:  Matched,
				Value: values,
				Start: start,
				End:   start,
			}
		}

		if first.Kind == Failed {
			return Match[[]T]{
				Kind:  Failed,
				Start: start,
				End:   first.End,
				Err:   first.Err,
			}
		}

		values = append(values, first.Value)

		for {
			beforeSeparator := ctx.Pos

			sep := separator.Match(ctx)

			if sep.Kind == NoMatch {
				ctx.Reset(beforeSeparator)
				break
			}

			if sep.Kind == Failed {
				return Match[[]T]{
					Kind:  Failed,
					Start: start,
					End:   sep.End,
					Err:   sep.Err,
				}
			}

			next := item.Match(ctx)

			if next.Kind != Matched {
				return Match[[]T]{
					Kind:  Failed,
					Start: start,
					End:   ctx.Pos,
					Err: &ParseError{
						Position: ctx.Pos,
						Message:  "expected item after separator",
					},
				}
			}

			values = append(values, next.Value)
		}

		return Match[[]T]{
			Kind:  Matched,
			Value: values,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func SepBy1[T, S any](
	item Pattern[T],
	separator Pattern[S],
) Pattern[[]T] {
	return PatternFunc[[]T](func(ctx *Context) Match[[]T] {
		start := ctx.Pos

		result := SepBy(item, separator).Match(ctx)

		if result.Kind != Matched {
			return result
		}

		if len(result.Value) == 0 {
			ctx.Reset(start)

			return Match[[]T]{
				Kind:  NoMatch,
				Start: start,
				End:   start,
			}
		}

		return result
	})
}

func Between[L, T, R any](
	left Pattern[L],
	content Pattern[T],
	right Pattern[R],
) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		start := ctx.Pos

		l := left.Match(ctx)
		if l.Kind != Matched {
			ctx.Reset(start)

			return Match[T]{
				Kind:  l.Kind,
				Start: start,
				End:   l.End,
				Err:   l.Err,
			}
		}

		value := content.Match(ctx)
		if value.Kind != Matched {
			return Match[T]{
				Kind:  Failed,
				Start: start,
				End:   value.End,
				Err:   value.Err,
			}
		}

		r := right.Match(ctx)
		if r.Kind != Matched {
			return Match[T]{
				Kind:  Failed,
				Start: start,
				End:   r.End,
				Err:   r.Err,
			}
		}

		return Match[T]{
			Kind:  Matched,
			Value: value.Value,
			Start: start,
			End:   ctx.Pos,
		}
	})
}

func OneOf[T any](patterns ...Pattern[T]) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		start := ctx.Pos

		var best *Match[T]

		for _, pattern := range patterns {
			ctx.Reset(start)

			result := pattern.Match(ctx)

			if result.Kind == Matched {
				return result
			}

			if result.Kind == Failed {
				if best == nil || result.End > best.End {
					copy := result
					best = &copy
				}
			}
		}

		ctx.Reset(start)

		if best != nil {
			return *best
		}

		return Match[T]{
			Kind:  NoMatch,
			Start: start,
			End:   start,
		}
	})
}

func Bind[A, B any](
	first Pattern[A],
	next func(A) Pattern[B],
) Pattern[B] {
	return PatternFunc[B](func(ctx *Context) Match[B] {
		start := ctx.Pos

		a := first.Match(ctx)

		if a.Kind != Matched {
			ctx.Reset(start)

			return Match[B]{
				Kind:  a.Kind,
				Start: start,
				End:   a.End,
				Err:   a.Err,
			}
		}

		b := next(a.Value).Match(ctx)

		if b.Kind != Matched {
			return b
		}

		b.Start = start
		return b
	})
}

func ChainBind[T any](
	initial Pattern[T],
	next func(T) Pattern[T],
) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		first := initial.Match(ctx)

		if first.Kind != Matched {
			return first
		}

		value := first.Value
		start := first.Start

		for {
			pos := ctx.Pos

			result := next(value).Match(ctx)

			switch result.Kind {
			case NoMatch:
				ctx.Pos = pos

				return Match[T]{
					Kind:  Matched,
					Value: value,
					Start: start,
					End:   ctx.Pos,
				}

			case Failed:
				return result

			case Matched:
				if ctx.Pos == pos {
					panic("ChainBind: pattern matched without consuming input")
				}

				value = result.Value
			}
		}
	})
}

type Located[T any] struct {
	Value T
	Span  models.Span
}

func Locate[T any](p Pattern[T]) Pattern[Located[T]] {
	return PatternFunc[Located[T]](func(ctx *Context) Match[Located[T]] {
		result := p.Match(ctx)

		if result.Kind != Matched {
			return Match[Located[T]]{
				Kind:  result.Kind,
				Start: result.Start,
				End:   result.End,
				Err:   result.Err,
			}
		}

		span := ctx.Span(result.Start, result.End)

		return Match[Located[T]]{
			Kind: Matched,
			Value: Located[T]{
				Value: result.Value,
				Span:  span,
			},
			Start: result.Start,
			End:   result.End,
		}
	})
}

func Peek[T any](pattern Pattern[T]) Pattern[T] {
	return PatternFunc[T](func(ctx *Context) Match[T] {
		start := ctx.Pos

		result := pattern.Match(ctx)

		ctx.Pos = start

		return result
	})
}

func Until[T any, U any](
	pattern Pattern[T],
	terminator Pattern[U],
) Pattern[[]T] {
	return PatternFunc[[]T](func(ctx *Context) Match[[]T] {
		start := ctx.Pos
		values := []T{}

		for {
			end := Peek(terminator).Match(ctx)

			if end.OK() {
				break
			}

			result := pattern.Match(ctx)

			if !result.OK() {
				return Match[[]T]{
					Kind:  result.Kind,
					Start: start,
					End:   ctx.Pos,
					Err:   result.Err,
				}
			}

			if ctx.Pos == result.Start {
				// Prevent an infinite loop if the pattern
				// succeeds without consuming tokens.
				panic("Until: pattern consumed no tokens")
			}

			values = append(values, result.Value)
		}

		return Match[[]T]{
			Kind:  Matched,
			Value: values,
			Start: start,
			End:   ctx.Pos,
		}
	})
}
