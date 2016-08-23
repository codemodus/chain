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
	c := New(nestedHandler(b0), nestedHandler(b1))
	c = c.Append(nestedHandler(b1))

	got, want := len(c.hs), 3
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	// should be same order as above, doubled - nestedHandler(b0) == "00"
	gResp, err := handlersToString(c.hs)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}
	wResp := b0 + b0 + b1 + b1 + b1 + b1
	if gResp != wResp {
		t.Errorf("got %s, want %s\n", gResp, wResp)
	}
}

func TestUnitMerge(t *testing.T) {
	c1 := New(nestedHandler(b0))
	c2 := New(nestedHandler(b1), nestedHandler(b0))
	c3 := c1.Merge(c2)

	got, want := len(c3.hs), 3
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}

	// should be same order as above, doubled - nestedHandler(b0) == "00"
	gResp, err := handlersToString(c3.hs)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}
	wResp := b0 + b0 + b1 + b1 + b0 + b0
	if gResp != wResp {
		t.Errorf("got %s, want %s\n", gResp, wResp)
	}
}

func TestUnitEnd(t *testing.T) {
	c := New(nestedHandler(b0), nestedHandler(b1))
	nh := c.End(nil)
	h1 := c.End(http.HandlerFunc(endHandler))

	w, err := record(nh)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	gCode, wCode := w.Code, http.StatusOK
	if gCode != wCode {
		t.Errorf("got %d, want %d", gCode, wCode)
	}

	w, err = record(h1)
	if nil != err {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	gCode, wCode = w.Code, http.StatusOK
	if gCode != wCode {
		t.Errorf("got %d, want %d", gCode, wCode)
	}

	gResp := w.Body.String()
	wResp := b0 + b1 + bEnd + b1 + b0
	if gResp != wResp {
		t.Errorf("got %s, want %s\n", gResp, wResp)
	}
}

func TestUnitEndFn(t *testing.T) {
	c := New(nestedHandler(b1), nestedHandler(b0))
	nh := c.EndFn(nil)
	h1 := c.EndFn(endHandler)

	w, err := record(nh)
	if err != nil {
		t.Fatalf("unexpected error: %s\n", err.Error())
	}

	gCode, wCode := w.Code, http.StatusOK
	if gCode != wCode {
		t.Errorf("got %d, want %d", gCode, wCode)
	}

	w, err = record(h1)
	if nil != err {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	gCode, wCode = w.Code, http.StatusOK
	if gCode != wCode {
		t.Errorf("got %d, want %d", gCode, wCode)
	}

	gResp := w.Body.String()
	wResp := b1 + b0 + bEnd + b0 + b1
	if gResp != wResp {
		t.Errorf("got %s, want %s\n", gResp, wResp)
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

func handlersToString(hs []func(http.Handler) http.Handler) (string, error) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return "", err
	}

	for _, fn := range hs {
		fn(http.HandlerFunc(emptyHandler)).ServeHTTP(w, r)
	}

	return w.Body.String(), err
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

func emptyNestedHandler(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.ServeHTTP(w, r)
	})
}
