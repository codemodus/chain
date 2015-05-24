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

func TestChain(t *testing.T) {
	c0 := chain.New(context.Background(), ctxHandlerWrapper0)
	c1 := c0.Append(ctxHandlerWrapper1, chain.Meld(httpHandlerWrapperA))
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.EndFn(ctxHandler))
	m.Handle(r1, c1.EndFn(ctxHandler))
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
	c0 := chain.New(context.Background(), emptyCtxHandlerWrapper)
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

func TestContextContinuity(t *testing.T) {
	tStr := "test_string"
	ctx := context.Background()
	ctx = initPHFC(ctx)
	if conCtx, ok := getPHFC(ctx); ok {
		*conCtx = setString(*conCtx, tStr)
	}

	c0 := chain.New(ctx, ctxContinuityWrapper, ctxHandlerWrapper0)
	c0 = c0.Append(ctxHandlerWrapper1, chain.Meld(httpHandlerWrapperA))
	m := http.NewServeMux()
	r0 := "/0"
	m.Handle(r0, c0.EndFn(ctxContinuityHandler))
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

	bb := &bytes.Buffer{}
	bb.Write(bTxt0)
	bb.Write(bTxt1)
	bb.Write(bTxtA)
	bb.Write([]byte(tStr))
	bb.Write(bTxtA)
	bb.Write(bTxt1)
	bb.Write(bTxt0)
	bb.Write([]byte(tStr))

	want := string(bb.Bytes())
	got := string(rb0)
	if got != want {
		t.Errorf("Body = %v, want %v", got, want)
	}
}

func Example() {
	// ctxHandlerWrapper0 writes "0" to the response body before and after
	// ServeHTTPContext() is called.
	// ctxHandlerWrapper1 writes "1" to the response body before and after
	// ServeHTTPContext() is called.
	// httpHandlerWrapperA writes "A" to the response body before and after
	// ServeHTTP() is called.
	// ctxHandler writes "_END_" to the response body and returns.
	ctx := context.Background()
	chain0 := chain.New(ctx, ctxHandlerWrapper0, ctxHandlerWrapper1)
	chain1 := chain0.Append(chain.Meld(httpHandlerWrapperA), ctxHandlerWrapper1)

	m := http.NewServeMux()
	m.Handle("/test/01_End", chain0.EndFn(ctxHandler))
	m.Handle("/test/01A1_End", chain1.EndFn(ctxHandler))

	s := httptest.NewServer(m)

	resp0, err := http.Get(s.URL + "/test/01_End")
	if err != nil {
		fmt.Println(err)
	}
	defer resp0.Body.Close()
	rBody0, err := ioutil.ReadAll(resp0.Body)
	if err != nil {
		fmt.Println(err)
	}

	resp1, err := http.Get(s.URL + "/test/01A1_End")
	if err != nil {
		fmt.Println(err)
	}
	defer resp1.Body.Close()
	rBody1, err := ioutil.ReadAll(resp1.Body)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Chain 0 Body:", string(rBody0))
	fmt.Println("Chain 1 Body:", string(rBody1))

	// Output:
	// Chain 0 Body: 01_END_10
	// Chain 1 Body: 01A1_END_1A10
}

func ctxHandlerWrapper0(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt0)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt0)
	})
}

func ctxHandlerWrapper1(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write(bTxt1)
		n.ServeHTTPContext(ctx, w, r)
		w.Write(bTxt1)
	})
}

func httpHandlerWrapperA(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(bTxtA)
		n.ServeHTTP(w, r)
		w.Write(bTxtA)
	})
}

func emptyCtxHandlerWrapper(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)
	})
}

func ctxHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	w.Write(bTxtEnd)
	return
}

func ctxContinuityWrapper(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)

		if conCtx, ok := getPHFC(ctx); ok {
			if s, ok := getString(*conCtx); ok {
				w.Write([]byte(s))
			}
		}
	})
}

func ctxContinuityHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if conCtx, ok := getPHFC(ctx); ok {
		if s, ok := getString(*conCtx); ok {
			w.Write([]byte(s))
		}
	}
	return
}

func nilHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	return
}

type adapter struct {
	ctx context.Context
	h   chain.Handler
}

func (a *adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.h.ServeHTTPContext(a.ctx, w, r)
}

type reqCtxKey int

const (
	postHandlerFuncCtxKey reqCtxKey = iota
	keyString
)

func setString(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, keyString, s)
}

func getString(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(keyString).(string)
	return s, ok
}

// initPHFC takes a context.Context and places a pointer to it within itself.
// This is useful for carrying data into the post ServeHTTPContext area of
// Handler wraps.  PHFC stands for Post HandlerFunc Context.
func initPHFC(ctx context.Context) context.Context {
	return context.WithValue(ctx, postHandlerFuncCtxKey, &ctx)
}

// getPHFC takes a context.Context and returns a pointer to the context.Context
// set in InitPHFC.
func getPHFC(ctx context.Context) (*context.Context, bool) {
	cx, ok := ctx.Value(postHandlerFuncCtxKey).(*context.Context)
	return cx, ok
}

func BenchmarkChain10(b *testing.B) {
	c0 := chain.New(context.Background(), emptyCtxHandlerWrapper,
		emptyCtxHandlerWrapper, emptyCtxHandlerWrapper, emptyCtxHandlerWrapper,
		emptyCtxHandlerWrapper, emptyCtxHandlerWrapper, emptyCtxHandlerWrapper,
		emptyCtxHandlerWrapper, emptyCtxHandlerWrapper, emptyCtxHandlerWrapper)
	m := http.NewServeMux()
	m.Handle("/", c0.EndFn(nilHandler))
	s := httptest.NewServer(m)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		re0, err := http.Get(s.URL + "/")
		if err != nil {
			b.Error(err)
		}
		defer re0.Body.Close()
	}
}

func BenchmarkNest10(b *testing.B) {
	h := &adapter{
		ctx: context.Background(),
		h: emptyCtxHandlerWrapper(emptyCtxHandlerWrapper(
			emptyCtxHandlerWrapper(emptyCtxHandlerWrapper(
				emptyCtxHandlerWrapper(emptyCtxHandlerWrapper(
					emptyCtxHandlerWrapper(emptyCtxHandlerWrapper(
						emptyCtxHandlerWrapper(emptyCtxHandlerWrapper(
							chain.HandlerFunc(nilHandler))))))))))),
	}
	m := http.NewServeMux()
	m.Handle("/", h)
	s := httptest.NewServer(m)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		re0, err := http.Get(s.URL + "/")
		if err != nil {
			b.Error(err)
		}
		defer re0.Body.Close()
	}
}
