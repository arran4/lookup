# lookup

`lookup` is a small library that brings dynamic navigation similar to JSONPath or JSONata to native Go structures. It works with structs, maps, arrays, slices and even functions while remaining null safe. The API exposes a few simple primitives that can be composed into complex queries.

This README outlines the core concepts, shows practical examples and documents the available modifiers and data structures. Many of the examples exist as runnable Go programs under `examples/`.

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
| `Default(v)` | Use `v` whenever the lookup would result in an invalid value. |
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
| Map(?) | Collections | Runs a modifier over a collection and converts it to another value based on content | | |
| Union(?) | Collections | Combine two results with no duplicates | | |
| Append(?) | Collections | Combine two results with duplicates | | |
| Intersection(?) | Collections | Combine two results only returning common values | | |
| First(?) | Collections | Returns the first value only that matches a predicate, using a Modifier as a predicate | | |
| Last(?) | Collections | Returns the last value only that matches a predicate, using a Modifier as a predicate | | |
| Range(?, ?) | Collections | Like Index but returns an array | | |
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

## Extensions

Please contribute any external libraries that build upon `lookup` here:
* ...

## Contributing

Bug reports and pull requests are welcome on GitHub. Feel free to open issues for discussion or ideas.

A JSONata parser built on top of this library is planned. Once available it will be linked here.

## License

This project is publicly available under the Affero GPL license. See `LICENSE` for details.

### Custom Licensing

If the AGPL does not suit your needs, log an issue or email to discuss alternatives.

## Q&A

### Can I use it as part of tests in a private library?

Yes. Tests are not considered part of the released binary.
### Is lookup production ready?

The core API is stable but still evolving. Feedback and contributions are encouraged before locking it down.

### Where can I get help?

Open an issue on GitHub if you have questions or run into problems.
