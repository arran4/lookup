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
			{Name: "child3", Size: 3, Tags: []string{"groupA", "groupC"}},
		},
	}

	r := lookup.Reflect(root)

	log.Printf("child names: %#v", r.Find("Children").Find("Name").Raw())

	log.Printf("groupA children: %#v", r.Find("Children",
		lookup.Filter(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))))).Find("Name").Raw())

	log.Printf("largest child size: %#v",
		r.Find("Children", lookup.Map(lookup.This("Size")), lookup.Last(nil)).Raw())

	log.Printf("has groupB child: %#v",
		r.Find("Children", lookup.Any(lookup.Map(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupB")))))).Raw())

	log.Printf("union tags: %#v",
		r.Find("Tags", lookup.Union(lookup.Array("groupB", "groupC"))).Raw())

	log.Printf("append tag: %#v",
		r.Find("Tags", lookup.Append(lookup.Array("groupA"))).Raw())

	log.Printf("intersect tags with groupB: %#v",
		r.Find("Tags", lookup.Intersection(lookup.Array("groupA", "groupB"))).Raw())

	log.Printf("first child: %s",
		r.Find("Children", lookup.First(nil)).Find("Name").Raw())

	log.Printf("last child: %s",
		r.Find("Children", lookup.Last(nil)).Find("Name").Raw())

	log.Printf("children from index 1: %#v",
		r.Find("Children", lookup.Range(1, nil)).Find("Name").Raw())
}
