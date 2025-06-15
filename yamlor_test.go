package lookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlorBasic(t *testing.T) {
	data := []byte("name: root\nchild:\n  size: 10\nlist:\n  - 1\n  - 2\n  - 3\n")
	r := Yaml(data)
	assert.Equal(t, "root", r.Find("name").Raw())
	assert.Equal(t, 10, r.Find("child").Find("size").Raw())
	assert.Equal(t, 2, r.Find("list", Index(1)).Raw())
	assert.Equal(t, "def", r.Find("missing", Default("def")).Raw())
}
