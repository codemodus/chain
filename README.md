# chain

    go get github.com/codemodus/chain

Package chain aids the composition of nested http.Handler instances.

Nesting functions is a simple concept.  If your nested handler order does not 
need to be composable, please do not use this or any similar package and avoid 
adding a dependency to your project.

## Usage

```go
type Chain
    func New(handlers ...func(http.Handler) http.Handler) Chain
    func (c Chain) Append(handlers ...func(http.Handler) http.Handler) Chain
    func (c Chain) End(handler http.Handler) http.Handler
    func (c Chain) EndFn(handlerFunc http.HandlerFunc) http.Handler
    func (c Chain) Merge(chains ...Chain) Chain
```

### Setup

```go
import (
    // ...

    "github.com/codemodus/chain"
)

func main() {
    // ...

  	// Nested handlers write either "0" or "1" to the response body before
	// and after ServeHTTP() is called.
	//
	// endHandler writes "_END_" to the response body.

	ch00 := New(nestedHandler0, nestedHandler0)
	ch001 := ch00.Append(nestedHandler1)

	ch1 := New(nestedHandler1)
	ch1001 := ch1.Merge(ch001)

	mux := http.NewServeMux()
	mux.Handle("/00_End", ch00.EndFn(endHandler))     // Resp Body: "00_END_00"
	mux.Handle("/001_End", ch001.EndFn(endHandler))   // Resp Body: "001_END_100"
	mux.Handle("/1001_End", ch1001.EndFn(endHandler)) // Resp Body: "1001_END_1001"

    // ...
}
```

### Nestable http.Handler

```go
func nestableHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
        
        next.ServeHTTP(w, r)
    	
        // ...
    })
}
```

## More Info

### Changes in go1.7+/chain2.0+

As of version 1.7, the http package's Request type includes a field which holds 
an implementation of context.Context.  Further, the context package has been
added to the standard library. There is now no need for the custom Handler
defined in previous versions of chain.  Please refer to the following command to
ease the process of updating your source.

    sed -r -e 's/chain\.Handler/http.Handler/g' \
        -e 's/[a-zA-Z0-9]+ context\.Context, ([a-zA-Z0-9]+) (http\.ResponseWriter)/\1 \2/' \
        -e 's/ServeHTTPContext\([a-zA-Z0-9]+, /ServeHTTP(/'

Beyond this, any usage of chain.Set(context.Context) will need to be modified
manually. Adding the affected logic as a nested handler is a simple and 
effective alternative. Don't forget to run gofmt/goimports.

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/chain)

## Benchmarks

These results are for comparison of normally nested functions, and chained 
functions.  Each benchmark includes 10 functions prior to the final handler.

    go1.7rc6
    benchmark             iter       time/iter   bytes alloc         allocs
    ---------             ----       ---------   -----------         ------
    BenchmarkChain10     30000     61.08 μs/op     3684 B/op   51 allocs/op
    BenchmarkChain10-4   20000     75.28 μs/op     3691 B/op   51 allocs/op
    BenchmarkChain10-8   20000     76.21 μs/op     3695 B/op   51 allocs/op
    BenchmarkNest10      30000     59.57 μs/op     3684 B/op   51 allocs/op
    BenchmarkNest10-4    20000     74.74 μs/op     3692 B/op   51 allocs/op
    BenchmarkNest10-8    20000     75.65 μs/op     3697 B/op   51 allocs/op
