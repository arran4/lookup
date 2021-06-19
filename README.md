# lookup

This is a "simple" lookup library I wrote for go.. It's designed to bring some of the dynamicness you can get with lookup
solutions like Jsonpath and Jsonata to structures inside go.

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
log.Printf("%#v", lookup.Reflector(root).Find("Node2").Find("1").Find("Size").Raw()) // 12
log.Printf("%#v", lookup.Reflector(root).Find("Node2").Find("-1").Find("Size").Raw()) // 35
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

# License

Sorry guys decided to try something new. DUAL licensed. Public license is AGPL contact me about private licenses.

I do not consider tests part of the binary.

# Pricing

Subject to negotiation if you can use the AGPL do so. As always I would like to find more time to work on these this might make
it possible.
