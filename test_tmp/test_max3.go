package main

import (
	"fmt"
	"github.com/arran4/lookup/jsonata"
	"github.com/arran4/lookup"
)

func main() {
	ast, _ := jsonata.Parse("[1,2,3]")
	runner := jsonata.Compile(ast)
	res := runner.Run(lookup.NewScope(nil, nil))
	fmt.Printf("Arg RAW: %v (%T)\n", res.Raw(), res.Raw())
}
