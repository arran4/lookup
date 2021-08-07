package main

import (
	"fmt"
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	rootNode := &TreeNode{
		Name: "A",
		ChildrenNodes: []*TreeNode{
			&TreeNode{
				Name: "B",
				ChildrenNodes: []*TreeNode{
					&TreeNode{
						Name:          "D",
						ChildrenNodes: []*TreeNode{},
					},
				},
			},
			&TreeNode{
				Name:          "C",
				ChildrenNodes: []*TreeNode{},
			},
		},
	}
	var root lookup.Pathor = lookup.NewInterfaceor(&InterfaceorNode{
		Node: rootNode,
	})
	assert.Equal(t, "A(B(D()),C())", fmt.Sprintf("%s", root.Raw()), "A failed")
	assert.Equal(t, "B(D())", fmt.Sprintf("%s", root.Find("B").Raw()), "A->B failed")
	assert.Equal(t, "D()", fmt.Sprintf("%s", root.Find("B").Find("D").Raw()), "A->B->D failed")
	assert.Equal(t, nil, root.Find("B").Find("ZZ").Raw(), "A->B->ZZ failed")
	assert.Error(t, lookup.ErrNoSuchPath, root.Find("B").Find("ZZ", lookup.Default("Not found")).Raw(), "A->B->ZZ failed")
}
