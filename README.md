# chain
--
    import "github.com/codemodus/chain"

Package chain aids in the ordering and reuse of context-aware Handler wrapper
chains.

Review the test file for examples covering chain manipulation, and a way to pass
a context across scopes (for common use cases race conditions will not be
encountered, but caution is warranted). Benchmarks are available showing a
negligible increase in processing time and memory consumption, and no increase
in memory allocations relative to nesting functions without an aid.

## Usage

View the [GoDoc](http://godoc.org/github.com/codemodus/chain)
