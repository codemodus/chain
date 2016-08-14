// Package chain aids the composition of nested http.Handler instances.
package chain

import "net/http"

// Chain contains the current nested http.Handler data.
type Chain struct {
	hs []func(http.Handler) http.Handler
}

// New receives one or more nested http.Handler instances, and returns a new
// Chain.
func New(handlers ...func(http.Handler) http.Handler) Chain {
	return Chain{hs: handlers}
}

// Append receives one or more nested http.Handler instances, and appends the
// value to the returned Chain.
func (c Chain) Append(handlers ...func(http.Handler) http.Handler) Chain {
	c.hs = append(c.hs, handlers...)
	return c
}

// Merge receives one or more Chain instances, and returns a merged Chain.
func (c Chain) Merge(cs ...Chain) Chain {
	for k := range cs {
		c.hs = append(c.hs, cs[k].hs...)
	}
	return c
}

// End receives an http.Handler, and returns an http.Handler comprised of all
// nested http.Handler data where the received http.Handler is the endpoint.
func (c Chain) End(h http.Handler) http.Handler {
	if h == nil {
		h = http.HandlerFunc(emptyHandler)
	}

	for i := len(c.hs) - 1; i >= 0; i-- {
		h = c.hs[i](h)
	}

	return h
}

// EndFn receives an instance of http.HandlerFunc, then passes it to End to
// return an http.Handler.
func (c Chain) EndFn(h http.HandlerFunc) http.Handler {
	if h == nil {
		h = http.HandlerFunc(emptyHandler)
	}

	return c.End(h)
}

func emptyHandler(w http.ResponseWriter, r *http.Request) {
	return
}
