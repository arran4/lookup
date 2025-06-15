package main

import (
	"log"

	"github.com/arran4/lookup"
)

type Node struct {
	Name     string
	Size     int
	Tags     []string
	Children []*Node
}

func main() {
	root := &Node{
		Name: "root",
		Size: 3,
		Tags: []string{"root", "groupA"},
		Children: []*Node{
			{Name: "child1", Size: 1, Tags: []string{"groupA"}},
			{Name: "child2", Size: 2, Tags: []string{"groupB"}},
		},
	}

	r := lookup.Reflect(root)

	log.Printf("child names: %#v", r.Find("Children").Find("Name").Raw())

	log.Printf("groupA children: %#v", r.Find("Children",
		lookup.Filter(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))))).Find("Name").Raw())

	log.Printf("largest child size: %#v",
		r.Find("Children", lookup.Map(lookup.This("Size")), lookup.Index("-1")).Raw())

	log.Printf("has groupB child: %#v",
		r.Find("Children", lookup.Any(lookup.Map(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupB")))))).Raw())
}
