package chain_test

import (
	"bytes"
	"fmt"
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

func ctxWrapper0(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt0)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt0)
	})
}

func ctxWrapper1(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt1)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt1)
	})
}

func stdWrapperA(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bTxtA)
		n.ServeHTTP(w, r)
		w.Write(bTxtA)
	})
}

func emptyCtxWrapper(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)
	})
}

func ctxEndPoint(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write(bTxtEnd)
	return
}

func TestChain(t *testing.T) {
	c0 := chain.New(context.Background(), ctxWrapper0)
	c1 := c0.Append(ctxWrapper1, chain.Meld(stdWrapperA))
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.EndFn(ctxEndPoint))
	m.Handle(r1, c1.EndFn(ctxEndPoint))
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
	c0 := chain.New(context.Background(), emptyCtxWrapper)
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

	want := http.StatusOK
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

func Example() {
	// ctxWrapper0 writes "0" to the response body before and after
	// ServeHTTPContext() is called.
	// ctxWrapper1 writes "1", and stdWrapperA writes "A", in the same manner
	// (though, stdWrapperA calls ServeHTTP()).
	// ctxEndPoint writes "_END_" to the response body and returns.
	chain0 := chain.New(context.Background(), ctxWrapper0)
	chain1 := chain0.Append(ctxWrapper1, chain.Meld(stdWrapperA))

	m := http.NewServeMux()
	m.Handle("/test0", chain0.EndFn(ctxEndPoint))
	m.Handle("/test01AEnd", chain1.EndFn(ctxEndPoint))

	s := httptest.NewServer(m)

	resp0, err := http.Get(s.URL + "/test0")
	if err != nil {
		fmt.Println(err)
	}
	defer resp0.Body.Close()
	rBody0, err := ioutil.ReadAll(resp0.Body)
	if err != nil {
		fmt.Println(err)
	}

	resp1, err := http.Get(s.URL + "/test01AEnd")
	if err != nil {
		fmt.Println(err)
	}
	defer resp1.Body.Close()
	rBody1, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Chain 0:", string(rBody0))
	fmt.Println("Chain 1:", string(rBody1))

	// Output:
	// Chain 0: 0_END_0
	// Chain 1: 01A_END_A10
}
