package lookup_test

import (
	"fmt"
	"log"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
	"github.com/arran4/lookup/jsonata"
)

func Example_jsonata() {
	// 1. Data
	data := map[string]interface{}{
		"Account": map[string]interface{}{
			"Name":    "Firefly",
			"Balance": 100.50,
		},
	}

	// 2. Compile JSONata expression
	// This uses the lookups JSONata compiler which binds functions at parse time.
	ast, err := jsonata.Parse("Account.Name")
	if err != nil {
		log.Fatal(err)
	}
	q := jsonata.Compile(ast)

	// 3. Execute
	result := q.Run(&lookup.Scope{
		Current: lookup.Reflect(data),
	})

	fmt.Println(result.Raw())
	// Output: Firefly
}

func Example_jsonataCustomFunction() {
	// 1. Create a Context with standard functions and register custom one
	// Thread-safe!
	funcs := jsonata.GetStandardFunctions()
	funcs["$greet"] = &GreetFunc{}
	ctx := &evaluator.Context{
		Functions: funcs,
	}

	// 2. Data
	data := map[string]interface{}{
		"Name": "Alice",
	}

	// 3. Compile expression using the custom function
	ast, err := jsonata.Parse("$greet(Name)")
	if err != nil {
		log.Fatal(err)
	}
	q := jsonata.Compile(ast)

	// 4. Run with Context
	result := q.Run(lookup.NewScopeWithContext(nil, lookup.Reflect(data), ctx))

	fmt.Println(result.Raw())

	// Output: Hello, Alice!
}

type GreetFunc struct{}

func (g *GreetFunc) Call(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "Hello!", nil
	}
	name, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("arg 0 must be string")
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}
