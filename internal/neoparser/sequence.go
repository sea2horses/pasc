package neoparser

type Pair[A, B any] struct {
	First  A
	Second B
}

func Seq2[A, B any](
	a Pattern[A],
	b Pattern[B],
) Pattern[Pair[A, B]] {
	return PatternFunc[Pair[A, B]](func(ctx *Context) Match[Pair[A, B]] {
		start := ctx.Checkpoint()

		ra := a.Match(ctx)
		if ra.Kind != Matched {
			ctx.Restore(start)

			return Match[Pair[A, B]]{
				Kind:  ra.Kind,
				Start: start.Pos,
				End:   ra.End,
				Err:   ra.Err,
			}
		}

		rb := b.Match(ctx)
		if rb.Kind != Matched {
			ctx.Restore(start)

			return Match[Pair[A, B]]{
				Kind:  rb.Kind,
				Start: start.Pos,
				End:   rb.End,
				Err:   rb.Err,
			}
		}

		return Match[Pair[A, B]]{
			Kind: Matched,
			Value: Pair[A, B]{
				First:  ra.Value,
				Second: rb.Value,
			},
			Start: start.Pos,
			End:   ctx.Pos,
		}
	})
}

type Trio[A, B, C any] struct {
	First  A
	Second B
	Third  C
}

func Seq3[A, B, C any](
	a Pattern[A],
	b Pattern[B],
	c Pattern[C],
) Pattern[Trio[A, B, C]] {
	return PatternFunc[Trio[A, B, C]](func(ctx *Context) Match[Trio[A, B, C]] {
		start := ctx.Checkpoint()

		ra := a.Match(ctx)
		if ra.Kind != Matched {
			ctx.Restore(start)

			return Match[Trio[A, B, C]]{
				Kind:  ra.Kind,
				Start: start.Pos,
				End:   ra.End,
				Err:   ra.Err,
			}
		}

		rb := b.Match(ctx)
		if rb.Kind != Matched {
			ctx.Restore(start)

			return Match[Trio[A, B, C]]{
				Kind:  rb.Kind,
				Start: start.Pos,
				End:   rb.End,
				Err:   rb.Err,
			}
		}

		rc := c.Match(ctx)
		if rc.Kind != Matched {
			ctx.Restore(start)

			return Match[Trio[A, B, C]]{
				Kind:  rc.Kind,
				Start: start.Pos,
				End:   rc.End,
				Err:   rc.Err,
			}
		}

		return Match[Trio[A, B, C]]{
			Kind: Matched,
			Value: Trio[A, B, C]{
				First:  ra.Value,
				Second: rb.Value,
				Third:  rc.Value,
			},
			Start: start.Pos,
			End:   ctx.Pos,
		}
	})
}

func Left[A, B any](
	a Pattern[A],
	b Pattern[B],
) Pattern[A] {
	return Map(
		Seq2(a, b),
		func(v Pair[A, B]) A {
			return v.First
		},
	)
}

func Right[A, B any](
	a Pattern[A],
	b Pattern[B],
) Pattern[B] {
	return Map(
		Seq2(a, b),
		func(v Pair[A, B]) B {
			return v.Second
		},
	)
}

func After[A, B any](
	a Pattern[A],
	b Pattern[B],
) Pattern[B] {
	return Right(a, b)
}
