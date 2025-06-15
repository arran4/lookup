package lookup

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNode struct {
	A struct {
		B []struct {
			C int
		}
	}
}

func newTestNode() *testNode {
	n := &testNode{}
	n.A.B = []struct{ C int }{{C: 1}, {C: 2}}
	return n
}

func TestParseSimplePath(t *testing.T) {
	root := newTestNode()
	r := ParseSimplePath("A.B[0].C")
	res := r.Run(NewScope(nil, Reflect(root)))
	assert.Equal(t, 1, res.Raw())
}

func TestQuerySimplePath(t *testing.T) {
	root := newTestNode()
	res := QuerySimplePath(root, "A.B[-1].C")
	assert.Equal(t, 2, res.Raw())
}

func TestQueryInvalid(t *testing.T) {
	root := newTestNode()
	res := QuerySimplePath(root, "A.B[10].C")
	assert.IsType(t, &Invalidor{}, res)
}

func TestQueryNested(t *testing.T) {
	type nested struct{ A [][]int }
	n := &nested{A: [][]int{{1, 2}, {3, 4}}}
	res := QuerySimplePath(n, "A[1][0]")
	assert.Equal(t, 3, res.Raw())
}

func TestParseUnmatchedBracket(t *testing.T) {
	root := newTestNode()
	res := QuerySimplePath(root, "A.B[0.C")
	assert.IsType(t, &Invalidor{}, res)
}
