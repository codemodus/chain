package chain_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codemodus/chain"
	"golang.org/x/net/context"
)

var (
	bTxt0   = []byte("0")
	bTxt1   = []byte("1")
	bTxtA   = []byte("A")
	bTxtEnd = []byte("_END_")
)

func wrap0(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt0)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt0)
	})
}

func wrap1(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt1)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt1)
	})
}

func stdWrapA(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bTxtA)
		n.ServeHTTP(w, r)
		w.Write(bTxtA)
	})
}

func emptyWrap0(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)
	})
}

func end0(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write(bTxtEnd)
	return
}

func TestChain(t *testing.T) {
	c0 := chain.New(context.Background(), wrap0)
	c1 := c0.Append(wrap1, chain.Meld(stdWrapA))
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.EndFn(end0))
	m.Handle(r1, c1.EndFn(end0))
	s := httptest.NewServer(m)

	re0, err := http.Get(s.URL + r0)
	if err != nil {
		t.Error(err)
	}
	defer re0.Body.Close()
	rb0, err := ioutil.ReadAll(re0.Body)
	if err != nil {
		t.Error(err)
	}

	re1, err := http.Get(s.URL + r1)
	if err != nil {
		t.Error(err)
	}
	defer re1.Body.Close()
	rb1, err := ioutil.ReadAll(re1.Body)
	if err != nil {
		t.Error(err)
	}

	bb := &bytes.Buffer{}
	bb.Write(bTxt0)
	bb.Write(bTxtEnd)
	bb.Write(bTxt0)
	bb.Write(bTxt0)
	bb.Write(bTxt1)
	bb.Write(bTxtA)
	bb.Write(bTxtEnd)
	bb.Write(bTxtA)
	bb.Write(bTxt1)
	bb.Write(bTxt0)

	want := string(bb.Bytes())
	got := string(rb0) + string(rb1)
	if got != want {
		t.Errorf("Body = %v, want %v", got, want)
	}
}

func TestNilEnd(t *testing.T) {
	c0 := chain.New(context.Background(), emptyWrap0)
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.End(nil))
	m.Handle(r1, c0.EndFn(nil))
	s := httptest.NewServer(m)

	re0, err := http.Get(s.URL + r0)
	if err != nil {
		t.Error(err)
	}
	defer re0.Body.Close()
	rs0 := re0.StatusCode

	want := http.StatusNoContent
	got := rs0
	if got != want {
		t.Errorf("Status Code = %v, want %v", got, want)
	}

	re1, err := http.Get(s.URL + r1)
	if err != nil {
		t.Error(err)
	}
	defer re1.Body.Close()
	rs1 := re1.StatusCode

	got = rs1
	if got != want {
		t.Errorf("Status Code = %v, want %v", got, want)
	}
}
