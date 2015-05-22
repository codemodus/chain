// Package chain enables flexible reordering and reuse of nested context-aware
// functions.  Some convenience functions are also provided for easing the
// passing of data through instances of Chain where the mutated data would
// normally become out of scope.
package chain

import (
	"net/http"

	"golang.org/x/net/context"
)

// Handler interface must be implemented for an object to be included within
// a Chain.
type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is an adapter which allows functions with the appropriate
// signature to be, subsequently, treated as a Handler.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPContext calls h(ctx, w, r)
func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h(ctx, w, r)
}

// Chain holds the basic components used to order handler wraps.
type Chain struct {
	ctx context.Context
	m   []func(Handler) Handler
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
	mw func(http.Handler) http.Handler
}

// New takes one or more Handler wraps, and returns a new Chain.
func New(ctx context.Context, mw ...func(Handler) Handler) Chain {
	return Chain{ctx: ctx, m: mw}
}

// Append takes one or more Handler wraps, and appends it/them to the returned
// Chain.
func (c Chain) Append(mw ...func(Handler) Handler) Chain {
	c.m = append(c.m, mw...)
	return c
}

// End takes a Handler and returns an http.Handler.
func (c Chain) End(h Handler) http.Handler {
	if h == nil {
		return nil
	}

	for i := len(c.m) - 1; i >= 0; i-- {
		h = c.m[i](h)
	}

	f := handlerAdapter{
		ctx: c.ctx, h: h,
	}
	return f
}

// EndFn takes a func that matches the HandlerFunc type, assigns it as such if
// it is not already so, then passes it to End.
func (c Chain) EndFn(h HandlerFunc) http.Handler {
	if h == nil {
		return c.End(nil)
	}
	return c.End(h)
}

// Meld takes a standard http.Handler wrapping function and returns a Handler
// wrap.  This is useful for making non-context aware http.Handler wraps
// compatible with the rest of a Chain.
func Meld(h func(http.Handler) http.Handler) func(Handler) Handler {
	return func(n Handler) Handler {
		return HandlerFunc(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				x := noCtxHandlerAdapter{
					mw: h, handlerAdapter: handlerAdapter{ctx: ctx, h: n},
				}
				h(x).ServeHTTP(w, r)
			},
		)
	}
}
