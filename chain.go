// Package chain enables flexible ordering and reuse of context-aware Handler
// wrapper chains.  Review the test file for examples covering chain
// manipulation and a way to pass a context across scopes.
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

type noCtxHandlerAdapter struct {
	handlerAdapter
	hw func(http.Handler) http.Handler
}

// New takes one or more Handler wrappers, and returns a new Chain.
func New(ctx context.Context, hws ...func(Handler) Handler) Chain {
	return Chain{ctx: ctx, hws: hws}
}

// Append takes one or more Handler wrappers, and appends the value to the
// returned Chain.
func (c Chain) Append(hws ...func(Handler) Handler) Chain {
	c.hws = append(c.hws, hws...)
	return c
}

// End takes a Handler and returns an http.Handler.
func (c Chain) End(h Handler) http.Handler {
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

// Meld takes a http.Handler wrapper and returns a Handler wrapper.  This is
// useful for making non-context aware http.Handler wrappers compatible with
// the rest of a Handler Chain.
func Meld(hw func(http.Handler) http.Handler) func(Handler) Handler {
	return func(h Handler) Handler {
		return HandlerFunc(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				x := noCtxHandlerAdapter{
					hw: hw, handlerAdapter: handlerAdapter{ctx: ctx, h: h},
				}
				hw(x).ServeHTTP(w, r)
			},
		)
	}
}

func nilHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return
}
