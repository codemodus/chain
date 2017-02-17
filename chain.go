// Package chain aids the composition of nested http.Handler instances.
package chain

import "net/http"

// Chain contains the current nested http.Handler data.
type Chain struct {
	hs []func(http.Handler) http.Handler
}

// New receives one or more nested http.Handler instances, and returns a new
// Chain.
func New(handlers ...func(http.Handler) http.Handler) *Chain {
	return &Chain{hs: handlers}
}

// appendHandlers differs from the append built-in in that it does not set the
// cap of the currently tracked handlers slice beyond it's required length.
// This ensures that additional appending produces expected results instead of
// allowing for the potential of collision/overwriting.
func appendHandlers(hs []func(http.Handler) http.Handler, ahs ...func(http.Handler) http.Handler) []func(http.Handler) http.Handler {
	lcur := len(hs)
	ltot := lcur + len(ahs)
	if ltot > cap(hs) {
		nhs := make([]func(http.Handler) http.Handler, ltot)
		copy(nhs, hs)
		hs = nhs
	}

	copy(hs[lcur:], ahs)

	return hs
}

// Append receives one or more nested http.Handler instances, and appends the
// value to the returned Chain.
func (c *Chain) Append(handlers ...func(http.Handler) http.Handler) *Chain {
	c = New(appendHandlers(c.hs, handlers...)...)

	return c
}

// Merge receives one or more Chain instances, and returns a merged Chain.
func (c *Chain) Merge(chains ...*Chain) *Chain {
	for k := range chains {
		c = New(appendHandlers(c.hs, chains[k].hs...)...)
	}

	return c
}

// Copy receives one Chain instance, and copies it's handlers into the
// receiver's handlers slice.
func (c *Chain) Copy(chain *Chain) {
	c.hs = make([]func(http.Handler) http.Handler, len(chain.hs))

	for k := range chain.hs {
		c.hs[k] = chain.hs[k]
	}
}

// End receives an http.Handler, and returns an http.Handler comprised of all
// nested http.Handler data where the received http.Handler is the endpoint.
func (c *Chain) End(handler http.Handler) http.Handler {
	if handler == nil {
		handler = http.HandlerFunc(emptyHandler)
	}

	for i := len(c.hs) - 1; i >= 0; i-- {
		handler = c.hs[i](handler)
	}

	return handler
}

// EndFn receives an instance of http.HandlerFunc, then passes it to End to
// return an http.Handler.
func (c *Chain) EndFn(handlerFunc http.HandlerFunc) http.Handler {
	if handlerFunc == nil {
		handlerFunc = http.HandlerFunc(emptyHandler)
	}

	return c.End(handlerFunc)
}

func emptyHandler(w http.ResponseWriter, r *http.Request) {}
