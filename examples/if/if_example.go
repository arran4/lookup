package main

import (
	"github.com/arran4/lookup"
	"log"
)

type Node struct {
	Name string
	Tags []string
}

func main() {
	root := &Node{Name: "child1", Tags: []string{"groupA"}}
	r := lookup.Reflect(root)
	desc := r.Find("", lookup.If(
		lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))),
		lookup.This("Name"),
		lookup.Constant("other"),
	)).Raw()
	log.Printf("result=%s", desc)
}
