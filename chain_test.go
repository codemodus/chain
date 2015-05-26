package chain_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codemodus/chain"
	"golang.org/x/net/context"
)

var (
	bTxt0   = "0"
	bTxt1   = "1"
	bTxtA   = "A"
	bTxtEnd = "_END_"
)

func Example() {
	ctx := context.Background()
	// Add common data to the context.

	// Each wrapper writes either "0", "1", or "A" to the response body before
	// and after ServeHTTPContext() is called.
	// ctxHandler writes "_END_" to the response body and returns.
	chain00 := chain.New(ctxHandlerWrapper0, ctxHandlerWrapper0).SetContext(ctx)
	chain00A1 := chain00.Append(chain.Convert(httpHandlerWrapperA), ctxHandlerWrapper1)

	chain100A1 := chain.New(ctxHandlerWrapper1).SetContext(ctx)
	chain100A1 = chain100A1.Merge(chain00A1)

	mux := http.NewServeMux()
	mux.Handle("/path_implies_body/00_End", chain00.EndFn(ctxHandler))
	mux.Handle("/path_implies_body/00A1_End", chain00A1.EndFn(ctxHandler))
	mux.Handle("/path_implies_body/100A1_End", chain100A1.EndFn(ctxHandler))

	server := httptest.NewServer(mux)

	rBody0, err := getReqBody(server.URL + "/path_implies_body/00_End")
	if err != nil {
		fmt.Println(err)
	}

	rBody1, err := getReqBody(server.URL + "/path_implies_body/00A1_End")
	if err != nil {
		fmt.Println(err)
	}

	rBody2, err := getReqBody(server.URL + "/path_implies_body/100A1_End")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Chain 0 Body:", rBody0)
	fmt.Println("Chain 1 Body:", rBody1)
	fmt.Println("Chain 2 Body:", rBody2)

	// Output:
	// Chain 0 Body: 00_END_00
	// Chain 1 Body: 00A1_END_1A00
	// Chain 2 Body: 100A1_END_1A001
}

func TestChain(t *testing.T) {
	c0 := chain.New(ctxHandlerWrapper0)
	c1 := c0.Append(ctxHandlerWrapper1, chain.Convert(httpHandlerWrapperA))
	cBefore0 := chain.New(ctxHandlerWrapper1)
	c0 = cBefore0.Merge(c0)
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.EndFn(ctxHandler))
	m.Handle(r1, c1.EndFn(ctxHandler))
	s := httptest.NewServer(m)

	rb0, err := getReqBody(s.URL + r0)
	if err != nil {
		t.Error(err)
	}

	rb1, err := getReqBody(s.URL + r1)
	if err != nil {
		t.Error(err)
	}

	want := bTxt1 + bTxt0 + bTxtEnd + bTxt0 + bTxt1
	want += bTxt0 + bTxt1 + bTxtA + bTxtEnd + bTxtA + bTxt1 + bTxt0
	got := rb0 + rb1
	if got != want {
		t.Errorf("Body = %v, want %v", got, want)
	}
}

func TestNilEnd(t *testing.T) {
	c0 := chain.New(emptyCtxHandlerWrapper)
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
	defer func() {
		_ = re0.Body.Close()
	}()
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
	defer func() {
		_ = re1.Body.Close()
	}()
	rs1 := re1.StatusCode

	got = rs1
	if got != want {
		t.Errorf("Status Code = %v, want %v", got, want)
	}
}

func TestContextContinuity(t *testing.T) {
	str := "my_string"
	ctx := context.Background()
	ctx = initPHFC(ctx)
	if conCtx, ok := getPHFC(ctx); ok {
		*conCtx = setMyString(*conCtx, str)
	}

	c0 := chain.New(ctxContinuityWrapper, ctxHandlerWrapper0).SetContext(ctx)
	c0 = c0.Append(ctxHandlerWrapper1, chain.Convert(httpHandlerWrapperA))
	m := http.NewServeMux()
	r0 := "/0"
	m.Handle(r0, c0.EndFn(ctxContinuityHandler))
	s := httptest.NewServer(m)

	rb0, err := getReqBody(s.URL + r0)
	if err != nil {
		t.Error(err)
	}

	want := bTxt0 + bTxt1 + bTxtA + str + bTxtA + bTxt1 + bTxt0 + str
	got := rb0
	if got != want {
		t.Errorf("Body = %v, want %v", got, want)
	}
}

func TestContextChange(t *testing.T) {
	str0 := "my_string_0"
	str1 := "my_string_1"
	ctx0 := context.Background()
	ctx0 = setMyString(ctx0, str0)
	ctx1 := setMyString(ctx0, str1)

	c0 := chain.New(emptyCtxHandlerWrapper).SetContext(ctx0)
	c1 := c0.SetContext(ctx1)
	m := http.NewServeMux()
	r0 := "/0"
	r1 := "/1"
	m.Handle(r0, c0.EndFn(ctxChangeHandler))
	m.Handle(r1, c1.EndFn(ctxChangeHandler))
	s := httptest.NewServer(m)

	rb0, err := getReqBody(s.URL + r0)
	if err != nil {
		t.Error(err)
	}

	rb1, err := getReqBody(s.URL + r1)
	if err != nil {
		t.Error(err)
	}

	want := str0 + str1
	got := rb0 + rb1
	if got != want {
		t.Errorf("Body = %v, want %v", got, want)
	}
}

func getReqBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	_ = resp.Body.Close()
	return string(body), nil
}

func ctxHandlerWrapper0(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(bTxt0))
		n.ServeHTTPContext(ctx, w, r)
		_, _ = w.Write([]byte(bTxt0))
	})
}

func ctxHandlerWrapper1(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(bTxt1))
		n.ServeHTTPContext(ctx, w, r)
		_, _ = w.Write([]byte(bTxt1))
	})
}

func httpHandlerWrapperA(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(bTxtA))
		n.ServeHTTP(w, r)
		_, _ = w.Write([]byte(bTxtA))
	})
}

func emptyCtxHandlerWrapper(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)
	})
}

func ctxHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(bTxtEnd))
	return
}

func ctxContinuityWrapper(n chain.Handler) chain.Handler {
	return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		n.ServeHTTPContext(ctx, w, r)

		if conCtx, ok := getPHFC(ctx); ok {
			if s, ok := getMyString(*conCtx); ok {
				_, _ = w.Write([]byte(s))
			}
		}
	})
}

func ctxContinuityHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if conCtx, ok := getPHFC(ctx); ok {
		if s, ok := getMyString(*conCtx); ok {
			_, _ = w.Write([]byte(s))
		}
	}
	return
}

func ctxChangeHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if s, ok := getMyString(ctx); ok {
		_, _ = w.Write([]byte(s))
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

func setMyString(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, keyString, s)
}

func getMyString(ctx context.Context) (string, bool) {
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
	c0 := chain.New(emptyCtxHandlerWrapper,
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
		_ = re0.Body.Close()
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
		_ = re0.Body.Close()
	}
}
