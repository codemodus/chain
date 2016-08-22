package chain

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	b0   = "0"
	b1   = "1"
	bEnd = "_END_"
)

func TestUnitAppend(t *testing.T) {
	c := New(emptyNestedHandler)
	c = c.Append(emptyNestedHandler)

	if 2 != len(c.hs) {
		t.Fatalf("want chain hs with len %d, got %d\n", 2, len(c.hs))
	}
}

func TestUnitMerge(t *testing.T) {
	c1 := New(emptyNestedHandler)
	c2 := New(emptyNestedHandler, emptyNestedHandler)
	c3 := c1.Merge(c2)

	if 3 != len(c3.hs) {
		t.Fatalf("want chain hs with len %d, got %d\n", 3, len(c3.hs))
	}
}

func TestUnitEnd(t *testing.T) {
	c := New(nestedHandler0)
	h := c.End(http.HandlerFunc(endHandler))

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	if http.StatusOK != w.Code {
		t.Fatalf("want status %d, got %d\n", http.StatusOK, w.Code)
	}

	resp := w.Body.String()
	wResp := b0 + bEnd + b0
	if wResp != resp {
		t.Fatalf("want response %s, got %s\n", wResp, resp)
	}
}

func TestUnitEndNilHandler(t *testing.T) {
	c := New(emptyNestedHandler)
	h := c.End(nil)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	if http.StatusOK != w.Code {
		t.Fatalf("want status %d, got %d\n", http.StatusOK, w.Code)
	}
}

func TestUnitEndFn(t *testing.T) {
	c := New(emptyNestedHandler)
	h := c.EndFn(endHandler)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	if http.StatusOK != w.Code {
		t.Fatalf("want status %d, got %d\n", http.StatusOK, w.Code)
	}
}

func TestUnitEndFnNilHandler(t *testing.T) {
	c := New(emptyNestedHandler)
	h := c.EndFn(nil)

	w, err := record(h)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	if http.StatusOK != w.Code {
		t.Fatalf("want status %d, got %d\n", http.StatusOK, w.Code)
	}
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

func emptyNestedHandler(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.ServeHTTP(w, r)
	})
}
