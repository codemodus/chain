package chain_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/codemodus/chain"
)

func Example() {
	// Nested handlers write either "0" or "1" to the response body before
	// and after ServeHTTP() is called.
	//
	// endHandler writes "_END_" to the response body.

	ch00 := chain.New(nestedHandler0, nestedHandler0)
	ch001 := ch00.Append(nestedHandler1)

	ch1 := chain.New(nestedHandler1)
	ch1001 := ch1.Merge(ch001)

	mux := http.NewServeMux()
	mux.Handle("/00_End", ch00.EndFn(endHandler))
	mux.Handle("/001_End", ch001.EndFn(endHandler))
	mux.Handle("/1001_End", ch1001.EndFn(endHandler))

	server := httptest.NewServer(mux)

	resp00, err := respBody(server.URL + "/00_End")
	if err != nil {
		fmt.Println(err)
	}

	resp001, err := respBody(server.URL + "/001_End")
	if err != nil {
		fmt.Println(err)
	}

	resp1001, err := respBody(server.URL + "/1001_End")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Chain 00 Resp:", resp00)
	fmt.Println("Chain 001 Resp:", resp001)
	fmt.Println("Chain 1001 Resp:", resp1001)

	// Output:
	// Chain 00 Resp: 00_END_00
	// Chain 001 Resp: 001_END_100
	// Chain 1001 Resp: 1001_END_1001
}

func respBody(url string) (string, error) {
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
