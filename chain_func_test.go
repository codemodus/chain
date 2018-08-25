package chain_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codemodus/chain/v2"
)

var (
	b0   = "0"
	b1   = "1"
	bEnd = "_END_"
)

func TestFuncHandlerOrder(t *testing.T) {
	c0 := chain.New(nestedHandler(b0), nestedHandler(b0))
	c0 = c0.Append(nestedHandler(b1), nestedHandler(b1))
	c0 = c0.Append(nestedHandler(b1), nestedHandler(b1))

	c1 := c0.Append(nestedHandler(b1))

	c0 = c0.Append(nestedHandler(b0))

	cm := chain.New(nestedHandler(b0), nestedHandler(b0))
	c0 = c0.Merge(cm)
	c1 = c1.Merge(cm)

	h0 := c0.EndFn(endHandler)
	h1 := c1.EndFn(endHandler)

	w, err := record(h0)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	got := w.Body.String()
	want := b0 + b0 + b1 + b1 + b1 + b1 + b0 + b0 + b0 + bEnd + b0 + b0 + b0 + b1 + b1 + b1 + b1 + b0 + b0
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	w, err = record(h1)
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	got = w.Body.String()
	want = b0 + b0 + b1 + b1 + b1 + b1 + b1 + b0 + b0 + bEnd + b0 + b0 + b1 + b1 + b1 + b1 + b1 + b0 + b0
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func nestedHandler(msg string) func(http.Handler) http.Handler {
	return func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(msg))
			n.ServeHTTP(w, r)
			_, _ = w.Write([]byte(msg))
		})
	}
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
