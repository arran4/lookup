package lookup

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJsonorBasic(t *testing.T) {
	data := []byte(`{"name":"root","child":{"size":10},"list":[1,2,3]}`)
	r := Json(data)
	assert.Equal(t, "root", r.Find("name").Raw())
	assert.Equal(t, 10.0, r.Find("child").Find("size").Raw())
	assert.Equal(t, 2.0, r.Find("list", Index(1)).Raw())
	assert.Equal(t, "def", r.Find("missing", Default("def")).Raw())
}
