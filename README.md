# lookup

`lookup` is a small library that brings dynamic navigation similar to JSONPath or JSONata to native Go structures. It works with structs, maps, arrays, slices and even functions while remaining null safe. The API exposes a few simple primitives that can be composed into complex queries.

This README outlines the core concepts, shows practical examples and documents the available modifiers and data structures. Many of the examples exist as runnable Go programs under `examples/`.

## Why lookup?

While Go's strong typing is excellent for safety and performance, it can be cumbersome when dealing with deeply nested dynamic data or exploring unknown structures (like complex JSON/YAML configuration).

Standard approaches often involve:
*   Verbose type assertions at every step.
*   Risk of panics if a pointer is nil.
*   Complex reflection code that is hard to maintain.

`lookup` provides a middle ground: **safe, dynamic navigation** without the verbosity or risk. It allows you to write expressive queries to drill down into your data, automatically handling:
*   Nil checks (returns an error-implementing object instead of panicking).
*   Slice/Array indexing (including negative indices).
*   Map lookups.
*   Interface wrapping.

## Installation

To use `lookup` as a library in your Go project:

```bash
go get github.com/arran4/lookup
```

## Concepts

| Concept | Description |
|---------|-------------|
| **Pathor** | Interface returned from all queries. Exposes `Find`, `Raw`, `Type` and `Value`. |
| **Reflector** | Implementation of `Pathor` based on reflection for arbitrary Go values. Use `lookup.Reflect` to create one. |
| **Jsonor** | Lazily unmarshals raw JSON as fields are requested. Use `lookup.Json` to create one. |
| **Yamlor** | Lazily unmarshals raw YAML as fields are requested. Use `lookup.Yaml` to create one. |
| **Interfaceor** | Wraps a user defined `Interface` so you can implement custom lookups. |
| **Constantor** | Holds a constant value and is often used internally by modifiers. |
| **Invalidor** | Represents an invalid path while still implementing `Pathor`. |
| **Relator** | Stores relative lookups used by modifiers such as `This`, `Parent` and `Result`. |

## Quick Start

The following short program demonstrates navigating a struct. You can run it with `go run examples/basic_example.go`.

```go
package main

import (
    "log"

    "github.com/arran4/lookup"
)

type Node struct {
    Name string
    Size int
}

func main() {
    root := &Node{Name: "root", Size: 10}
    r := lookup.Reflect(root)

    log.Printf("name = %s", r.Find("Name").Raw())
    log.Printf("size = %d", r.Find("Size").Raw())
}
```

Running the program prints:

```
name = root
size = 10
```

### JSON Example

`Json` lets you query raw JSON without fully unmarshalling it:

```go
raw := []byte(`{"name":"root","sizes":[1,2,3]}`)
r := lookup.Json(raw)
log.Printf("last size = %d", r.Find("sizes", lookup.Index("-1")).Raw())
```

### YAML Example

`Yaml` behaves the same for YAML input:

```go
raw := []byte("name: root\nsizes:\n  - 1\n  - 2\n  - 3\n")
r := lookup.Yaml(raw)
log.Printf("first size = %d", r.Find("sizes", lookup.Index(0)).Raw())
```

### Query Strings

For quick lookups the library understands a tiny query language that mirrors the
`Find` API. Paths are written using dot notation with optional array/slice
indexes in brackets. Negative indexes count from the end of the collection.
The helper `lookup.QuerySimplePath` parses the expression and runs it against your value:

```go
// Get the value of root.A.B[0].C
result := lookup.QuerySimplePath(root, "A.B[0].C").Raw()

// Last element using a negative index
last := lookup.QuerySimplePath(root, "A.B[-1].C").Raw()
```

If you need to reuse a query repeatedly you can compile it once using
`lookup.ParseSimplePath` which returns a `Relator` that can be executed on any `Pathor`.

## Modifiers

Modifiers are `Runner` implementations that transform the current scope of a lookup. They are passed to `Find` after the path name.

| Modifier | Purpose |
|----------|---------|
| `Index(i)` | Select an element from an array or slice. Supports negative indexes. |
| `Filter(r)` | Keep elements for which `r` returns true. |
| `Map(r)` | Convert each element using `r`. |
| `Contains(r)` | True if the current collection contains the result of `r`. |
| `In(r)` | True if the current value is present in the collection returned by `r`. |
| `Every(r)` | True if every element in scope matches `r`. |
| `Any(r)` | True if any element in scope matches `r`. |
| `Match(r)` | Proceed only if `r` evaluates to true. |
| `If(c, t, o)` | When `c` is true run `t` otherwise `o`. |
| `Default(v)` | Use `v` whenever the lookup would result in an invalid value. |
| `Union(r)` | Combine the current collection with `r` removing duplicates. |
| `Intersection(r)` | Elements present in both the current collection and `r`. |
| `First(r)` | Return the first value matching `r`. |
| `Last(r)` | Return the last value matching `r`. |
| `Range(s, e)` | Like `Index` but returns a slice from `s` to `e`. |
| `This(p)` `Parent(p)` `Result(p)` | Relative lookups executed from different points in a query. |

See `expression.go` and `collections.go` for the full list of helpers.

## Supported Data Structures

| Input | Description |
|-------|-------------|
| **Reflector** | Uses Go reflection to navigate arbitrary structs, maps, arrays, slices and functions. Channels are not supported. |
| **Invalidor** | Indicates that the search reached an invalid path. It implements the `error` interface. |
| **Constantor** | Similar to `Invalidor` but wraps a constant value. Attempting to navigate it does not change the position. |
| **Interfaceor** | Like `Reflector` but relies on a user supplied interface to obtain children. |
| **Jsonor** | Navigate raw JSON values without unmarshalling everything up front. |
| **Yamlor** | Navigate raw YAML values without unmarshalling everything up front. |
| **Relator** | Stores a path which can be replayed. Mostly used by modifiers for relative lookups. |

### Todo Data Structures

| Data structure | Description |
|----------------|-------------|
| `Simpleor` | A type-switch based version of `Reflector` for a smaller set of inputs. |


## Planned / TODO

| Modifier | Category | Description | Input | Output |
| --- | --- | --- | --- | --- |
| Append(?) | Collections | Combine two results with duplicates | | |
| If(?, ?, ?) | Expression | Conditional | | |
| Error(?) | Invalidor | Returns an invalid / failed result | | |
## Basic Lookup Behaviour

The library works by calling `Find` with field names. Arrays are expanded automatically so subsequent lookups act as map operations over the elements. Each field navigation can be followed by modifiers. For example, `Index` selects a specific element:

```go

lookup.Reflect(root).Find("Node2", lookup.Index("1")).Find("Size")
```

Here `.Find("Node2")` extracts an array and `Index("1")` picks a single element from it.

Functions with no arguments and a single return value (optionally followed by an error) can also be executed as part of a lookup.

```go
log.Printf("%s", lookup.Reflect(root).Find("Method1").Raw())
```

All usage is null-safe. When a path does not exist or an error occurs you receive an object implementing `error` which still satisfies `Pathor`:

```go
result := lookup.Reflect(root).Find("Node1").Find("DoesntExist")
if err, ok := result.(error); ok {
    panic(err)
}
```

Errors returned by functions are wrapped appropriately:

```go
result := lookup.Reflect(root).Find("Method2")
if errors.Is(result, Err1) {
    // expected error
}
```

## Advanced Usage

A runnable advanced example lives in `examples/advanced/advanced_example.go` and demonstrates combining modifiers for more complex queries:

```go
r := lookup.Reflect(root)

// Filter children by tag and fetch their names
names := r.Find("Children",
    lookup.Filter(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))))).
    Find("Name").Raw()

// Select the largest child size
largest := r.Find("Children", lookup.Map(lookup.This("Size")), lookup.Index("-1")).Raw()

// Check if any child has the tag "groupB"
hasB := r.Find("Children",
    lookup.Any(lookup.Map(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupB")))))).Raw()
```

A second runnable example demonstrates the collection helpers defined in
`examples/collections/collections_example.go`:

```go
numbers := []int{1, 2, 3, 3}
r := lookup.Reflect(numbers)

union := r.Find("", lookup.Union(lookup.Array(3, 4))).Raw()
intersection := r.Find("", lookup.Intersection(lookup.Array(2, 3, 4))).Raw()
first := r.Find("", lookup.First(lookup.Equals(lookup.Constant(3)))).Raw()
last := r.Find("", lookup.Last(lookup.Equals(lookup.Constant(3)))).Raw()
slice := r.Find("", lookup.Range(1, 3)).Raw()
```

Running the example prints:

```
union: []interface{}{1, 2, 3, 4}
intersection: []interface{}{2, 3}
first 3: 3
last 3: 3
range [1:3]: []int{2, 3}
```

Run `go test ./examples/...` to execute the examples as tests.

### Custom Interface Example

You can plug your own data structures into `lookup` by implementing the
`Interface` interface. Provide `Get` to return the next element of the path and
`Raw` to expose the underlying value. A runnable demo lives in
[`examples/interfacor1`](examples/interfacor1).

```go
type MyNode struct{}

func (n *MyNode) Get(path string) (interface{}, error) { /* ... */ }
func (n *MyNode) Raw() interface{} { return n }

r := lookup.NewInterfaceor(&MyNode{})
```

## Internals - Scope

Modifiers operate with a `Scope` that tracks the current, parent and position values. Nested and sequential modifiers adjust the scope without escaping the query. Consider the following:

```go
lookup.Reflect(root).Find("Node2", lookup.Index(lookup.Constant("-1")), lookup.Index(lookup.Constant("-2"))).Find("Size", lookup.Index(lookup.Constant("-3")))
```

Given this YAML:

```yaml
Node2:
  - Sizes:
      - 1
      - 2
      - 3
  - Sizes:
      - 4
      - 5
      - 6
  - Sizes:
      - 7
      - 8
      - 9
```

During the query `Index(Constant("-1"))` sees:
* `Scope.Parent` = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* `Scope.Current` = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* `Scope.Position` = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Result: `{Sizes: [7,8,9]}`

`Constant("-1")` sees the same parent, current and position but returns `-1`.

`Index(Constant("-2"))` then sees:
* `Scope.Parent` = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* `Scope.Current` = `{Sizes: [7,8,9]}`
* `Scope.Position` = `{Sizes: [7,8,9]}`
* Result: `8`

With other modifiers `Scope.Current` may differ from `Scope.Position`.

## Command Line Tools

Two helper binaries make navigating YAML and JSON from the shell easy. Both use
lookup's `SimplePath` syntax and share the same set of options.

### yaml-simpe-path

Reads one or more YAML documents and prints selected values. The interface is
inspired by classic Unix text processing tools with jq-style niceties.

```
Usage: yaml-simpe-path [options] PATH [PATH ...]

Options:
  -f string   YAML file to read (default stdin)
  -e string   simple path query (can be repeated)
  -d string   output delimiter (default "\n")
  -json       output as JSON
  -yaml       output as YAML (default)
  -raw        output raw value without formatting
  -grep str   only print results matching the regex
  -v          invert regex match
  -n          prefix results with their index
  -0          use NUL as output delimiter
  -count      only print the number of matched results
```

Example:

```bash
$ cat <<'EOF' > doc.yaml
name: foo
spec:
  replicas: 3
metadata:
  name: prod-service
EOF
$ yaml-simpe-path -f doc.yaml .spec.replicas
3
```

### json-simpe-path

Operates on JSON input with the same flags. It defaults to JSON output but can
emit YAML when `-yaml` is specified.

```bash
$ cat <<'EOF' > doc.json
{"name":"foo","spec":{"replicas":3},"metadata":{"name":"prod-service"}}
EOF
$ json-simpe-path -f doc.json .spec.replicas
3
```

Manual pages generated with `go-md2man` are available in the `man/` directory.

## Releases

Versioned releases are published automatically when a Git tag starting with
`v` is pushed. The release workflow runs [GoReleaser](https://goreleaser.com)
to build binaries for all supported platforms, package the man pages and upload
the archives to GitHub.

## Extensions

Please contribute any external libraries that build upon `lookup` here:
* ...

## Contributing

Bug reports and pull requests are welcome on GitHub. Feel free to open issues for discussion or ideas.

See [docs/jsonata.md](docs/jsonata.md) for a minimal JSONata parser built on top
of this package.

## License

This project is publicly available under the 3-Clause BSD License. See `LICENSE` for details.

## Q&A

### Can I use it as part of tests in a private library?

Yes. Tests are not considered part of the released binary.
### Is lookup production ready?

The core API is stable but still evolving. Feedback and contributions are encouraged before locking it down.

### Where can I get help?

Open an issue on GitHub if you have questions or run into problems.

## JSONata Feature Compatibility Matrix

Test results generated from `go test ./jsonata`.

| Feature Group | Passed | Failed |
|---|---|---|
| array-constructor | 0 | 21 |
| blocks | 0 | 7 |
| boolean-expresssions | 0 | 31 |
| closures | 0 | 2 |
| coalescing-operator | 0 | 13 |
| comments | 0 | 4 |
| comparison-operators | 0 | 29 |
| conditionals | 0 | 9 |
| context | 0 | 4 |
| default-operator | 0 | 14 |
| descendent-operator | 0 | 17 |
| encoding | 0 | 4 |
| errors | 0 | 27 |
| fields | 3 | 5 |
| flattening | 0 | 47 |
| function-abs | 0 | 4 |
| function-append | 0 | 6 |
| function-applications | 0 | 22 |
| function-assert | 0 | 8 |
| function-average | 0 | 13 |
| function-boolean | 0 | 24 |
| function-ceil | 0 | 4 |
| function-contains | 0 | 7 |
| function-count | 0 | 14 |
| function-decodeUrl | 0 | 3 |
| function-decodeUrlComponent | 0 | 3 |
| function-each | 0 | 3 |
| function-encodeUrl | 0 | 3 |
| function-encodeUrlComponent | 0 | 3 |
| function-error | 0 | 11 |
| function-eval | 0 | 8 |
| function-exists | 0 | 25 |
| function-floor | 0 | 4 |
| function-formatBase | 0 | 9 |
| function-formatNumber | 0 | 37 |
| function-fromMillis | 0 | 3 |
| function-join | 0 | 12 |
| function-keys | 0 | 7 |
| function-length | 0 | 17 |
| function-lookup | 0 | 4 |
| function-lowercase | 0 | 2 |
| function-max | 0 | 27 |
| function-merge | 0 | 5 |
| function-number | 0 | 34 |
| function-pad | 0 | 13 |
| function-power | 0 | 7 |
| function-replace | 0 | 12 |
| function-reverse | 0 | 4 |
| function-round | 0 | 18 |
| function-shuffle | 0 | 4 |
| function-sift | 0 | 5 |
| function-signatures | 0 | 35 |
| function-sort | 0 | 11 |
| function-split | 0 | 19 |
| function-spread | 0 | 4 |
| function-sqrt | 0 | 4 |
| function-string | 0 | 31 |
| function-substring | 0 | 19 |
| function-substringAfter | 0 | 5 |
| function-substringBefore | 0 | 5 |
| function-sum | 0 | 7 |
| function-tomillis | 0 | 13 |
| function-trim | 0 | 3 |
| function-typeOf | 0 | 13 |
| function-uppercase | 0 | 2 |
| function-zip | 0 | 6 |
| higher-order-functions | 0 | 3 |
| hof-filter | 0 | 4 |
| hof-map | 0 | 12 |
| hof-reduce | 0 | 11 |
| hof-single | 0 | 11 |
| hof-zip-map | 0 | 4 |
| inclusion-operator | 0 | 9 |
| lambdas | 0 | 14 |
| literals | 0 | 20 |
| matchers | 0 | 2 |
| missing-paths | 4 | 2 |
| multiple-array-selectors | 0 | 3 |
| null | 1 | 6 |
| numeric-operators | 0 | 19 |
| object-constructor | 0 | 27 |
| parentheses | 0 | 8 |
| partial-application | 0 | 5 |
| performance | 0 | 2 |
| predicates | 0 | 4 |
| quoted-selectors | 0 | 8 |
| range-operator | 0 | 25 |
| regex | 0 | 39 |
| simple-array-selectors | 4 | 19 |
| sorting | 0 | 21 |
| string-concat | 0 | 12 |
| tail-recursion | 0 | 10 |
| token-conversion | 0 | 4 |
| transform | 10 | 94 |
| transforms | 0 | 15 |
| variables | 0 | 13 |
| wildcards | 0 | 10 |
