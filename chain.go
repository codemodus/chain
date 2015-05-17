package chain

import (
	"net/http"

	"golang.org/x/net/context"
)

type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h HandlerFunc) ServeHTTPContext(c context.Context, w http.ResponseWriter, r *http.Request) {
	h(c, w, r)
}

type Middleware func(Handler) Handler

type Chain struct {
	ctx context.Context
	m   []Middleware
}

type handlerAdapter struct {
	ctx context.Context
	h   Handler
}

func (ha handlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ha.h.ServeHTTPContext(ha.ctx, w, r)
}

type NoCtxMiddleware func(http.Handler) http.Handler

type noCtxHandlerAdapter struct {
	ctx context.Context
	mw  NoCtxMiddleware
	n   Handler
}

func (ha noCtxHandlerAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ha.n.ServeHTTPContext(ha.ctx, w, r)
}

func New(ctx context.Context, mw ...Middleware) Chain {
	return Chain{ctx: ctx, m: mw}
}

func (c Chain) Append(mw ...Middleware) Chain {
	c.m = append(c.m, mw...)
	return c
}

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

func (c Chain) EndFn(h HandlerFunc) http.Handler {
	if h == nil {
		return c.End(nil)
	}
	return c.End(HandlerFunc(h))
}

// Adapt http.Handler into a ContextHandler
func Bridge(mw NoCtxMiddleware) Middleware {
	return func(n Handler) Handler {
		return HandlerFunc(
			func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				x := noCtxHandlerAdapter{ctx: ctx, mw: mw, n: n}
				mw(x).ServeHTTP(w, r)
			},
		)
	}
}
