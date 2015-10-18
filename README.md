# chain

    go get "github.com/codemodus/chain"

Package chain aids the composition of Handler wrapper chains that carry 
request-scoped data.

Nesting functions is a simple concept.  If your handler wrapper order does not 
need to be composable, do not use a package and avoid adding a dependency 
to your project.  However, nesting functions quickly becomes burdensome as the 
need for flexibility increases.  Add to that the need for request-scoped data, 
and Chain is a lightweight and complete solution.

## Usage

```go
func Convert(hw func(http.Handler) http.Handler) func(Handler) Handler
type Chain
    func New(hws ...func(Handler) Handler) Chain
    func (c Chain) Append(hws ...func(Handler) Handler) Chain
    func (c Chain) End(h Handler) http.Handler
    func (c Chain) EndFn(h HandlerFunc) http.Handler
    func (c Chain) Merge(cs ...Chain) Chain
    func (c Chain) SetContext(ctx context.Context) Chain
type Handler
type HandlerFunc
    func (h HandlerFunc) ServeHTTPContext(ctx context.Context, w http.ResponseWriter, r *http.Request)
```

### Setup

```go
import (
    // ...

    "github.com/codemodus/chain"
    "golang.org/x/net/context"
)

func main() {
    // ...

    ctx := context.Background()
    // Add common data to the context.

    ch0 := chain.New(firstWrapper, secondWrapper).SetContext(ctx)
    ch1 := ch0.Append(chain.Convert(httpHandlerWrapper), fourthWrapper)

    ch2 := chain.New(beforeFirstWrapper).SetContext(ctx)
    ch2 = ch2.Merge(ch1)

    m := http.NewServeMux()
    m.Handle("/1w2w_End1", ch0.EndFn(ctxHandler))
    m.Handle("/1w2w_End2", ch0.EndFn(anotherCtxHandler))
    m.Handle("/1w2wHw4w_End1", ch1.EndFn(ctxHandler))
    m.Handle("/0w1w2wHw4w_End1", ch2.EndFn(ctxHandler))

    // ...
}
```

### Handler Wrapper And Context Usage (Set)

```go
func firstWrapper(n chain.Handler) chain.Handler {
    return chain.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
        // ...
        
        ctx = setMyString(ctx, "Send this down the line.")
    	
        n.ServeHTTPContext(ctx, w, r)
    	
        // ...
    })
}
```
This function signature will make wrappers compatible with chain.  It's simple 
to make existing wrappers capable of carrying request-scoped data.

### Handler Function And Context Usage (Get)

```go
func ctxHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    // ...
    
    if s, ok := getMyString(ctx); ok {
        // s = "Send this down the line."
    }
    
    // ...
}
```
End-point functions will need to be adapted using chain.HandlerFunc.  As a 
convenience, EndFn will adapt functions with compatible signatures.  The 
prescribed signature is in accordance with practices outlined in the Go Blog.

### HTTP Handler Wrapper

```go
func httpHandlerWrapper(n http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...

        n.ServeHTTP(w, r)

        // ...
    })
}
```
A standard http.Handler wrapper is a perfect candidate for chain.Convert. If 
chain.Convert is used, the added http.Handler wrapper will be compatible with 
a Chain, but will not be able to make use of the request context.

## More Info 

### net/context?

net/context is made for this need and enables some interesting capabilities.
[The Go Blog: Context](https://blog.golang.org/context)

### What if SetContext goes unused?

A context.Background result is provided as a default initial context.Context.  
Conversely, it is also possible to set different initial context.Context 
objects for different chains which may or may not share wrapper content/order.

### Context Scope

By not using more broadly scoped context access, a small trick is needed to 
move data to and from certain points in the request life cycle.  For instance, 
if a final handler adds any data to the context, that data will not be 
accessible to any wrapper code residing after calls to 
ServeHTTP/ServeHTTPContext.

An example of resolving this is not being included here as it leaves the scope 
of the package itself. Though, this package is tested for the capability, so 
review the relevant test if need be.  Convenience functions can be found.

### catena

If a project is not in need of a request context, consider using 
[catena](https://github.com/codemodus/catena). The API is nearly identical to 
chain, so adding a request context is easy when needs change.

## Documentation

View the [GoDoc](http://godoc.org/github.com/codemodus/chain)

## Benchmarks

These results are for comparison of normally nested functions, and chained 
functions.  Each benchmark includes 10 functions prior to the final handler.

    benchmark           iter      time/iter   bytes alloc         allocs
    ---------           ----      ---------   -----------         ------
    BenchmarkChain10   50000    48.08 μs/op     4644 B/op   53 allocs/op
    BenchmarkNest10    50000    47.23 μs/op     4639 B/op   53 allocs/op
