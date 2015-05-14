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
	m []func(Handler) Handler
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

func New(mw ...func(Handler) Handler) Chain {
	return Chain{m: mw}
}

func (c Chain) Append(mw ...func(Handler) Handler) Chain {
	c.m = append(c.m, mw...)
	return c
}

func (c Chain) Then(h Handler) http.Handler {
	if h == nil {
		return nil
	}

	for i := len(c.m) - 1; i >= 0; i-- {
		h = c.m[i](h)
	}

	f := handlerAdapter{
		ctx: context.Background(), h: h,
	}
	return f
}

func (c Chain) ThenFunc(h HandlerFunc) http.Handler {
	if h == nil {
		return c.Then(nil)
	}
	return c.Then(HandlerFunc(h))
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
