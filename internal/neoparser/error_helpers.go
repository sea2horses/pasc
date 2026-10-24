package neoparser

func ensureError[T any](ctx *Context, result Match[T], expected string) ParseError {
	if result.Err != nil {
		return result.Err
	}
	return failureAt(ctx, max(ctx.Pos, result.End), expected)
}
