# JSONata Integration

The `jsonata` package provides a tiny parser and compiler that translate a very
small subset of JSONata syntax into `lookup` queries.

Supported features:

- Dot separated field navigation (`foo.bar`)
- Array indexes (`arr[0]`, `arr[-1]`)
- Equality filters (`books[author="Bob"]`)

Parsing yields an AST which can be compiled into a `lookup.Relator`. The
relator implements the `Runner` interface so it can be executed like other
modifiers.

```go
ast, _ := jsonata.Parse("Children[Name='child1'].Size")
query := jsonata.Compile(ast)

root := lookup.Reflect(data)
size := query.Run(lookup.NewScope(root, root)).Raw()
```

This example selects the `Size` of the child whose `Name` equals `child1`.
