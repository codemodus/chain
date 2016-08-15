package chain

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkChain10(b *testing.B) {
	c := New(emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler,
		emptyNestedHandler, emptyNestedHandler, emptyNestedHandler)

	m := http.NewServeMux()
	m.Handle("/", c.EndFn(emptyHandler))

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
						http.HandlerFunc(emptyHandler)))))))))))

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
