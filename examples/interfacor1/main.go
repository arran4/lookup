package main

import (
	"github.com/arran4/lookup"
	"log"
	"strings"
)

type TreeNode struct {
	Name          string
	ChildrenNodes []*TreeNode
}

func (tn *TreeNode) String() string {
	children := make([]string, 0, len(tn.ChildrenNodes))
	for _, c := range tn.ChildrenNodes {
		children = append(children, c.String())
	}
	return tn.Name + "(" + strings.Join(children, ",") + ")"
}

type InterfaceorNode struct {
	Node *TreeNode
}

func (i *InterfaceorNode) String() string {
	return i.Node.String()
}

func (i *InterfaceorNode) Get(path string) (interface{}, error) {
	for _, child := range i.Node.ChildrenNodes {
		if path == child.Name {
			return &InterfaceorNode{
				Node: child,
			}, nil
		}
	}
	return nil, nil
}

func (i *InterfaceorNode) Raw() interface{} {
	return i.Node
}

func main() {
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
	log.Printf("A = %s", root.Raw())
	log.Printf("A->B = %v", root.Find("B").Raw())
	log.Printf("A->B->D = %v", root.Find("B").Find("D").Raw())
	log.Printf("A->B->ZZ = %v", root.Find("B").Find("ZZ").Raw())
	log.Printf("A->B->ZZ = %v", root.Find("B").Find("ZZ", lookup.NewDefault("Not found")).Raw())
}
