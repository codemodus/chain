package chain_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codemodus/chain"
)

var (
	b0   = "0"
	b1   = "1"
	bEnd = "_END_"
)

func TestFuncHandlerOrder(t *testing.T) {
	c := chain.New(nestedHandler0, nestedHandler0)
	c = c.Append(nestedHandler1, nestedHandler1)

	mc := chain.New(nestedHandler0, nestedHandler0)
	c = c.Merge(mc)

	h := c.EndFn(endHandler)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	resp := w.Body.String()
	wResp := b0 + b0 + b1 + b1 + b0 + b0 + bEnd + b0 + b0 + b1 + b1 + b0 + b0
	if wResp != resp {
		t.Fatalf("want response %s, got %s\n", wResp, resp)
	}
}

func nestedHandler0(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(b0))
		n.ServeHTTP(w, r)
		_, _ = w.Write([]byte(b0))
	})
}

func nestedHandler1(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(b1))
		n.ServeHTTP(w, r)
		_, _ = w.Write([]byte(b1))
	})
}

func endHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(bEnd))
}

func record(h http.Handler) (*httptest.ResponseRecorder, error) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return nil, err
	}

	h.ServeHTTP(w, r)

	return w, nil
}
