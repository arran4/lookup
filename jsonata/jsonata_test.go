package jsonata

import (
	"encoding/json"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

type Node struct {
	Name     string
	Size     int
	Tags     []string
	Children []*Node
}

func runQuery(t *testing.T, data interface{}, q string) interface{} {
	ast, err := Parse(q)
	assert.NoError(t, err)
	r := Compile(ast)
	root := lookup.Reflect(data)
	res := r.Run(lookup.NewScope(root, root))
	return res.Raw()
}

func TestStructQueries(t *testing.T) {
	root := &Node{
		Name: "root",
		Size: 3,
		Children: []*Node{
			{Name: "child1", Size: 1},
			{Name: "child2", Size: 2},
		},
	}

	assert.Equal(t, "root", runQuery(t, root, "Name"))
	assert.Equal(t, "child2", runQuery(t, root, "Children[1].Name"))
	assert.Equal(t, []int{2}, runQuery(t, root, "Children[Name='child2'].Size"))
}

func TestJSONQueries(t *testing.T) {
	jsonData := []byte(`{"users":[{"name":"bob","age":5},{"name":"sam","age":7}]}`)
	var v struct {
		Users []struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"users"`
	}
	if err := json.Unmarshal(jsonData, &v); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	assert.Equal(t, []int{7}, runQuery(t, v, "Users[Name='sam'].Age"))
}
