package main

import (
	"fmt"
	"github.com/arran4/lookup/jsonata"
	"github.com/arran4/lookup"
)

func main() {
	ast, _ := jsonata.Parse("$max([1,2,3])")
	runner := jsonata.Compile(ast)
	scope := lookup.NewScope(nil, nil)
	res := runner.Run(scope)
	raw := res.Raw()
	fmt.Printf("RAW: %v (%T)\n", raw, raw)
}
