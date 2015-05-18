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
	mw  func(http.Handler) http.Handler
}

func New(ctx context.Context, mw ...func(Handler) Handler) Chain {
	return Chain{ctx: ctx, m: mw}
}

func (c Chain) Append(mw ...func(Handler) Handler) Chain {
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
func Bridge(h func(http.Handler) http.Handler) func(Handler) Handler {
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

type ctxKey int

const (
	postHandlerFuncCtxKey ctxKey = 0
)

func InitPHFC(ctx context.Context) context.Context {
	return context.WithValue(ctx, postHandlerFuncCtxKey, &ctx)
}

func GetPHFC(ctx context.Context) (*context.Context, bool) {
	cx, ok := ctx.Value(postHandlerFuncCtxKey).(*context.Context)
	return cx, ok
}
