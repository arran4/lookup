package main
import (
	"fmt"
	"encoding/json"
	"github.com/arran4/lookup/jsonata"
)
func main() {
	ast, _ := jsonata.Parse("$max([1,2,3])")
	b, _ := json.MarshalIndent(ast, "", "  ")
	fmt.Println(string(b))
}
