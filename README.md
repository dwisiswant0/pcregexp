# pcregexp

[![Tests](https://github.com/dwisiswant0/pcregexp/actions/workflows/tests.yaml/badge.svg)](https://github.com/dwisiswant0/pcregexp/actions/workflows/tests.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dwisiswant0/pcregexp.svg)](https://pkg.go.dev/github.com/dwisiswant0/pcregexp)

`pcregexp` is a drop‑in replacement for Go's standard [`regexp`](https://pkg.go.dev/regexp) package that uses the full capabilities of [PCRE2](https://github.com/PCRE2Project/pcre2) by loading the shared library dynamically at runtime, which enables cross‑compilation without the need for a C compiler (**no Cgo required!**). The API closely mirrors that of the standard library's `regexp` package while supporting advanced regex features like lookarounds and backreferences that PCRE2 provides.

> [!WARNING]
> PCRE2 supports features that can lead to exponential runtime in some cases. Use `pcregexp` only with *trusted* regex patterns to avoid potential regular expression denial-of-service (ReDoS) issues ([CWE-1333](https://cwe.mitre.org/data/definitions/1333.html)) or configure a global match context to impose limits and control resource usage by using [`SetMatchContext`](https://pkg.go.dev/github.com/dwisiswant0/pcregexp#SetMatchContext) function.

## Requirements

* **go1.20** or later.
* **PCRE2 10.x shared library** must be [installed](https://github.com/PCRE2Project/pcre2#quickstart) on your system.
* Supported platforms:
  * [PCRE2](https://github.com/PCRE2Project/pcre2#platforms)
  * [purego](https://github.com/ebitengine/purego#supported-platforms)

## Install

```bash
go install -v github.com/dwisiswant0/pcregexp@latest
```

## Usage

```go
package main

import (
    "fmt"
    
    "github.com/dwisiswant0/pcregexp"
)

func main() {
    // Compile a pattern. (Panics on error)
    re := pcregexp.MustCompile("p([a-z]+)ch")
    defer re.Close()

    // Check if the string matches the pattern.
    fmt.Println("MatchString(\"peach\"):", re.MatchString("peach"))

    // Find the leftmost match.
    fmt.Println("FindString(\"peach punch\"):", re.FindString("peach punch"))

    // Get the start and end indexes of the match.
    fmt.Println("FindStringIndex(\"peach punch\"):", re.FindStringIndex("peach punch"))

    // Retrieve the match along with its captured submatch.
    fmt.Println("FindStringSubmatch(\"peach punch\"):", re.FindStringSubmatch("peach punch"))

    // Perform global replacement (naively replaces all non-overlapping matches).
    src := "peach punch pinch"
    repl := "<fruit>"
    fmt.Println("ReplaceAllString:", re.ReplaceAllString(src, repl))
}
```

## Wrapped Regexp API

[![Go Reference](https://pkg.go.dev/badge/github.com/dwisiswant0/pcregexp.svg)](https://pkg.go.dev/github.com/dwisiswant0/pcregexp/pkg/regexp)

You may want to use the `regexp` package provided here, which wraps both Go's standard `regexp` package and a PCRE2-based implementation, `pcregexp`. This unified interface automatically selects the appropriate engine based on the regex features used, offering the best of both worlds.

## Benchmark

Execute the performance benchmark by running:

```bash
make bench
```

## TODO

See [TODO.md](/TODO.md).

## Status

> [!CAUTION]
> `pcregexp` has NOT reached 1.0 yet. Therefore, this library is currently not supported and does not offer a stable API; use at your own risk.

There are no guarantees of stability for the APIs in this library, and while they are not expected to change dramatically. API tweaks and bug fixes may occur.

## License

`pcregexp` is released by [**@dwisiswant0**](https://github.com/dwisiswant0) under the Apache 2.0 license. See [LICENSE](/LICENSE).