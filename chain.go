// Package chain aids the composition of Handler wrapper chains that carry
// request-scoped data.
//
// Review the test file for examples covering chain manipulation, and a way to
// pass data to detached scopes (for common  use cases race conditions will not
// be encountered, but caution is warranted). Benchmarks are available showing
// a negligible increase in processing time and memory consumption, and no
// increase in memory allocations compared to nesting functions without an aid.
package chain

import (
	"net/http"

	"golang.org/x/net/context"
)

// Handler interface must be implemented for a function to be able to be
// wrapped, or served.
type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is an adapter which allows a function with the appropriate
// signature to be treated as a Handler.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPContext calls h(ctx, w, r)
func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h(ctx, w, r)
}

// Chain holds the basic components used to order Handler wrapper chains.
type Chain struct {
	ctx context.Context
	hws []func(Handler) Handler
}

type handlerAdapter struct {
	ctx context.Context
	h   Handler
}

func (ha handlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ha.h.ServeHTTPContext(ha.ctx, w, r)
}

// New takes one or more Handler wrappers, and returns a new Chain.
func New(hws ...func(Handler) Handler) Chain {
	return Chain{hws: hws}
}

// SetContext takes a context.Context, and updates the stored/initial context
// of the returned Chain.
func (c Chain) SetContext(ctx context.Context) Chain {
	c.ctx = ctx
	return c
}

// Append takes one or more Handler wrappers, and appends the value to the
// returned Chain.
func (c Chain) Append(hws ...func(Handler) Handler) Chain {
	c.hws = append(c.hws, hws...)
	return c
}

// Merge takes one or more Chain objects, and appends the values' Handler
// wrappers to the returned Chain.
func (c Chain) Merge(cs ...Chain) Chain {
	for k := range cs {
		c.hws = append(c.hws, cs[k].hws...)
	}
	return c
}

// End takes a Handler and returns an http.Handler.
func (c Chain) End(h Handler) http.Handler {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	if h == nil {
		h = HandlerFunc(nilHandler)
	}

	for i := len(c.hws) - 1; i >= 0; i-- {
		h = c.hws[i](h)
	}

	r := handlerAdapter{
		ctx: c.ctx, h: h,
	}
	return r
}

// EndFn takes a func that matches the HandlerFunc type, then passes it to End.
func (c Chain) EndFn(h HandlerFunc) http.Handler {
	if h == nil {
		h = HandlerFunc(nilHandler)
	}
	return c.End(h)
}

// Convert takes a http.Handler wrapper and returns a Handler wrapper.  This is
// useful for making standard http.Handler wrappers compatible with a Chain.
func Convert(hw func(http.Handler) http.Handler) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			ha := handlerAdapter{ctx: ctx, h: h}
			hw(ha).ServeHTTP(w, r)
		})
	}
}

func nilHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return
}
