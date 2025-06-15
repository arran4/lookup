package lookup

import "testing"

type benchTreeNode struct {
	Name          string
	ChildrenNodes []*benchTreeNode
}

type benchInterfaceorNode struct{ Node *benchTreeNode }

func (i *benchInterfaceorNode) Get(path string) (interface{}, error) {
	for _, child := range i.Node.ChildrenNodes {
		if path == child.Name {
			return &benchInterfaceorNode{Node: child}, nil
		}
	}
	return nil, nil
}

func (i *benchInterfaceorNode) Raw() interface{} { return i.Node }

func BenchmarkFindNestedStruct(b *testing.B) {
	type Child struct{ Value int }
	type Root struct{ C *Child }
	r := &Root{C: &Child{Value: 42}}
	for i := 0; i < b.N; i++ {
		Reflect(r).Find("C").Find("Value").Raw()
	}
}

func BenchmarkFindNestedMap(b *testing.B) {
	root := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]int{"c": 5},
		},
	}
	for i := 0; i < b.N; i++ {
		Reflect(root).Find("a").Find("b").Find("c").Raw()
	}
}

func BenchmarkFindIndex(b *testing.B) {
	arr := []int{1, 2, 3, 4, 5}
	r := Reflect(arr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("", Index(3)).Raw()
	}
}

func BenchmarkFilterChildren(b *testing.B) {
	type Node struct {
		Name     string
		Size     int
		Tags     []string
		Children []*Node
	}
	root := &Node{
		Name: "root",
		Size: 3,
		Tags: []string{"root", "groupA"},
		Children: []*Node{
			{Name: "child1", Size: 1, Tags: []string{"groupA"}},
			{Name: "child2", Size: 2, Tags: []string{"groupB"}},
		},
	}
	r := Reflect(root)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("Children", Filter(This("Tags").Find("", Contains(Constant("groupA"))))).Find("Name").Raw()
	}
}

func BenchmarkMapIndex(b *testing.B) {
	type Node struct {
		Name     string
		Size     int
		Tags     []string
		Children []*Node
	}
	root := &Node{
		Name: "root",
		Size: 3,
		Tags: []string{"root", "groupA"},
		Children: []*Node{
			{Name: "child1", Size: 1, Tags: []string{"groupA"}},
			{Name: "child2", Size: 2, Tags: []string{"groupB"}},
		},
	}
	r := Reflect(root)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("Children", Map(This("Size")), Index("-1")).Raw()
	}
}

func BenchmarkAnyContains(b *testing.B) {
	type Node struct {
		Name     string
		Size     int
		Tags     []string
		Children []*Node
	}
	root := &Node{
		Name: "root",
		Size: 3,
		Tags: []string{"root", "groupA"},
		Children: []*Node{
			{Name: "child1", Size: 1, Tags: []string{"groupA"}},
			{Name: "child2", Size: 2, Tags: []string{"groupB"}},
		},
	}
	r := Reflect(root)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Find("Children", Any(Map(This("Tags").Find("", Contains(Constant("groupB")))))).Raw()
	}
}

func BenchmarkInterfaceorNested(b *testing.B) {
	rootNode := &benchTreeNode{
		Name: "A",
		ChildrenNodes: []*benchTreeNode{
			{
				Name: "B",
				ChildrenNodes: []*benchTreeNode{
					{Name: "D", ChildrenNodes: []*benchTreeNode{}},
				},
			},
			{Name: "C", ChildrenNodes: []*benchTreeNode{}},
		},
	}
	root := NewInterfaceor(&benchInterfaceorNode{Node: rootNode})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Find("B").Find("D").Raw()
	}
}
