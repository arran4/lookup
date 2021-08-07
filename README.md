# lookup

This is a "simple" lookup library I wrote for go.. It's designed to bring some of the dynamicness you can get with lookup
solutions like Jsonpath and Jsonata to structures inside go. Inspired by Jackson's .Path()

It's minimal to my needs. I look forwards to hearing from others and working with them to expand the scope.

It works by trying to "find" the next component you have requested. It will dynamically create arrays as necessary.

Say for instance you have this structure here:
```go
var (
	Err1 = errors.New("one")
)

type Root struct {
    Node1 Node1
    Node2 *Node2
}

func (r *Root) Method1 () (string, error) {
	return "hi", nil
}

func (r *Root) Method1 () (string, error) {
	return "", Err1
}

root := &Root{
    Node1: Node1{
        Name: "asdf",
    },
    Node2: []*Node2{
        {Size: 1,},
        {Size: 12,},
        {Size: 35,},
    },
}
```

You could run the following code on it:
```go
log.Printf("%#v", lookup.Reflector(root).Find("Node1").Find("Name").Raw()) // "asdf"
log.Printf("%#v", lookup.Reflector(root).Find("Node1").Find("DoesntExist").Raw()) // nil
log.Printf("%#v", lookup.Reflector(root).Find("Node1").Find("DoesntExist", lookup.NewDefault("N/A")).Raw()) // "N/A"
log.Printf("%#v", lookup.Reflector(root).Find("Node2").Find("Size").Raw()) // []int{ 1,12,35 }
log.Printf("%#v", lookup.Reflector(root).Find("Node2", Index("1")).Find("Size").Raw()) // 12
log.Printf("%#v", lookup.Reflector(root).Find("Node2", Index("-1")).Find("Size").Raw()) // 35
```

It will even execute functions (provided they have no arguments, and 1 primitive return, or a primitive and an error return)

```go
log.Printf("%#v", lookup.Reflector(root).Find("Method1").Raw()) // "hi"
```

All usages of the program should be null-safe you shouldn't be able to create a panic from inside the lookup codebase.
(If you write a crashing function it does /not/ call recover())


When you get to an invalid path or an error, the object being returned from `find()` is a valid error
```go
result := lookup.Reflector(root).Find("Node1").Find("DoesntExist")
if err, ok := result.(error); ok {
	panic(err)
}
```

It properly raps errors returned by functions:
```go
result := lookup.Reflector(root).Find("Method2")
if errors.Is(result, Err1); ok {
	// We expected this error
}
```

`find()` is the main implementation. It is designed to be simple and "null safe" (as in doesn't create any itself you can create them though!)
It hasn't been fully edge tested as I wrote it for my own testing - quickly. But expect the Find() function to be relatively stable.

Feel free to log issues and PRs:
* For any reason really
* Opinions welcome but not obligated to

# How to use the library

The basic idea behind the library is to act a lot like a meta language for jsonpath or some such. Such as you would write a query as such:
```
Root.Field.ChildField.ArrayElements.Field
```

If it encounters an array, it selects every element and every subsequent field become an implicit map operation. Each field navigation is followed by a "modifier" such as
in the following query, the "index" is a modifier.
```go
lookup.Reflector(root).Find("Node2", Index("1")).Find("Size")
```
So `.Find("Node2"` extracts the array. Each modifier then is run over the results of "Node2", in this case the modifier "Index" takes the array and returns the single element.

# Supported data structures

| Input Data structure | Description |
| --- | --- |
| Reflector | The most developed data structure and the basis. It takes any go input and will attempt to use it. It will not support channels however. It uses reflection for navigation, that includes functions.
| Invalidor | This is typically to indicate that the search function has reached and invalid path. It provides an `error` interface, however doesn't necessarily mean that an error has occurred, it could simply be that there was no where to go. You can use this in conjunction with the modifiers to simply mean "false"
| Constantor | This is similar to the invalidor however it contains a constant and can mean true or false. Attempting to navigate a constant will not change your position. Use a Reflector if you need to navigation. Constantor can mean the end of a search. It's often used just for nagivation events.
| Interfacteor | This is like Reflector but it expects the data structure passed in to adhere to a interface `Interface` it is a naive implementation and is likely to change.
| Relator | This stores a path, which can be replayed. It's used mostly in modifiers for the purpose of providing relative queries (Via `This()` `Parent()` or `Find()`. On it's own it will act as a modifier meaning "If Exists". Such as `lookup.Reflector(root).Find("Node2", This("Name")).Find("Size")` Will filter Node2 and return an array of Size for all elements which have a valid `.Name` Field.

## Todo data structures

| Data structure | Description |
| --- | --- |
| json.Raw / Jsonor | A version I wish to develop which does on-demand deserialization of Json based on the query - In a way which would also work for yaml etc if possible without including them as libraries |
| Simpleor | A typecast version of Reflector which works using type switching, type assertions rather than reflection, but will only work with a much more limited set of input |

# Modifiers (AKA Runners)

The modifier functions available:

| Modifier | Category | Description | Input | Output |
| --- | --- | --- | --- | --- |
| Index(i) | Collections | Selects a single index in an array | An int (raw or as a string.) Another modifier which is run for another compatible result. | The element referred to by index, or an Invalidor |
| Filter(?) | Collections | Runs a Modifier over a collection and filters out value based on boolean returned | | |
| Map(?) | Collections | Runs a modifier over a collection and converts it to another value based on content | | |
| Contains(?) | Collections | Returns Constantor(True) if scope contains ? | | |
| In(?) | Collections | Returns Constantor(True) if scope is in ? | | |
| Every(?) | Collections | Returns Constantor(True) if every element in scope is in ? | | |
| Any(?) | Collections | Returns Constantor(True) if any element in scope is in ? | | |
| Constant(?) | Constant | Returns Constantor(?) | | |
| True() | Constant | Returns Constantor(True) | | |
| False() | Constant | Constantor(False) | | |
| Array(?...) | Constant | Constantor(? as an array) | | |
| Match(?) | Expression | Permits scope if ? is True or Not Is Zero otherwise Invalidor | Boolean | |
| ToBool(?) | Expression | Converts scope to Constantor(bool) if possible otehrwise returns Invalidor | | Boolean |
| Truthy(?) | Expression | Converts scope to bool using truthy like logic otherwise returns Invalidor | | |
| Not(?) | Expression | Toggles Constantor(bool) | | |
| IsZero(?) | Expression | Uses Go Reflect's Value.IsZero to return Constantor(bool) | | |
| Default(?) | Expression | If scope is Invalidor Converts it to Constantor(?) | | |
| Find(?) | Relator | Runs a series of paths and Runners against the Scope.Current position | | |
| Parent(?) | Relator | Runs a series of paths and Runners against the Scope.Parent position (Note this changes) | | |
| This(?) | Relator | Runs a series of paths and Runners against the Scope.Current position | | |
| Result(?) | Relator | Runs a series of paths and Runners against the Scope.Position position | | |
| ValueOf(?) | Valuor | Evaluates ? as a Pathor and returns it as scope. | | |

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
| Error(?) | Invalidor | Returns an invalid / failed result  | | |

## Internals - Scope

(IIRC) Modifiers run with a scope. Depending on if they are Nested, or sequential modifies the scope. Scope doesn't escape out of
a query.

So with:
```go
lookup.Reflector(root).Find("Node2", Index(Constant("-1")), Index(Constant("-2"))).Find("Size", Index(Constant("-3"))
```

and

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

In all of the examples:

`Index(Constant("-1"))` sees
* Scope.Parent = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Current = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Position = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Result: `{Sizes: [7,8,9]}`

`Constant("-1")` sees
* Scope.Parent = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Current = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Position = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Result: `-1`

`Index(Constant("-2"))` sees
* Scope.Parent = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Current = `{Sizes: [7,8,9]}`
* Scope.Position = `{Sizes: [7,8,9]}`
* Result: `8`

`Constant("-2")` sees
* Scope.Parent = `[ { Sizes: [1,2,3] }, {Sizes: [4,5,6]}, {Sizes: [7,8,9]} ]`
* Scope.Current = `{Sizes: [7,8,9]}`
* Scope.Position = `{Sizes: [7,8,9]}`
* Result: `-2`

Note: With other Modifiers than `index` Scope.Current would be different to Scope.Position.

# Public Extensions

Please put any library that extends this in this section here:
* ...

# Public License

This project is publicly available under the Affero GPL license.

# Custom Licensing

If the AGPL is not suitable for your purposes, please log an issue or email me, and let's talk.

# Q/A

## Can I use it as part of tests in a private library

Yes. Tests are not considered part of the released binary.
