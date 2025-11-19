package generic

import (
	"context"
)

type Context[C context.Context] context.Context

type SubContext[C context.Context] struct {
	Context[C]
}

func (c *SubContext[C]) BaseContext() C {
	return c.Context.(C)
}

var _ context.Context = (*SubContext[context.Context])(nil)
