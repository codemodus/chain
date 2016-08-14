package chain

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	b0   = "0"
	b1   = "1"
	bEnd = "_END_"
)

func Example() {
	// Each nested handler writes either "0", or "1", to the response
	// body before and after ServeHTTPContext() is called.
	//
	// endHandler writes "_END_" to the response body and returns.

	ch00 := New(nestedHandler0, nestedHandler0)
	ch001 := ch00.Append(nestedHandler1)

	ch1 := New(nestedHandler1)
	ch1001 := ch1.Merge(ch001)

	mux := http.NewServeMux()
	mux.Handle("/00_End", ch00.EndFn(endHandler))
	mux.Handle("/001_End", ch001.EndFn(endHandler))
	mux.Handle("/1001_End", ch1001.EndFn(endHandler))

	server := httptest.NewServer(mux)

	resp0, err := getRespBody(server.URL + "/00_End")
	if err != nil {
		fmt.Println(err)
	}

	resp1, err := getRespBody(server.URL + "/001_End")
	if err != nil {
		fmt.Println(err)
	}

	resp2, err := getRespBody(server.URL + "/1001_End")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Chain 00 Resp:", resp0)
	fmt.Println("Chain 001 Resp:", resp1)
	fmt.Println("Chain 1001 Resp:", resp2)

	// Output:
	// Chain 00 Resp: 00_END_00
	// Chain 001 Resp: 001_END_100
	// Chain 1001 Resp: 1001_END_1001
}

func TestNew(t *testing.T) {
	c := New(emptyNestedHandler)
	if c.hs == nil {
		t.Fatal("want new chain with hs set, got nil")
	}
}

func getRespBody(url string) (string, error) {
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

func getRespStatus(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	_ = resp.Body.Close()

	return resp.StatusCode, nil
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

func emptyNestedHandler(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n.ServeHTTP(w, r)
	})
}

func endHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(bEnd))
}

func BenchmarkChain10(b *testing.B) {
	c0 := New(emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler)

	m := http.NewServeMux()
	m.Handle("/", c0.EndFn(nilHandler))

	s := httptest.NewServer(m)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		resp, err := http.Get(s.URL + "/")
		if err != nil {
			b.Error(err)
		}

		_ = resp.Body.Close()
	}
}

func BenchmarkNest10(b *testing.B) {
	h := emptyNestedHandler(emptyNestedHandler(
		emptyNestedHandler(emptyNestedHandler(
			emptyNestedHandler(emptyNestedHandler(
				emptyNestedHandler(emptyNestedHandler(
					emptyNestedHandler(emptyNestedHandler(
						http.HandlerFunc(nilHandler)))))))))))

	m := http.NewServeMux()
	m.Handle("/", h)

	s := httptest.NewServer(m)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		resp, err := http.Get(s.URL + "/")
		if err != nil {
			b.Error(err)
		}

		_ = resp.Body.Close()
	}
}
