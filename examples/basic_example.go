package main

import (
	"log"

	"github.com/arran4/lookup"
)

type Node struct {
	Name string
	Size int
}

func main() {
	root := &Node{Name: "root", Size: 10}
	r := lookup.Reflect(root)
	log.Printf("name = %s", r.Find("Name").Raw())
	log.Printf("size = %d", r.Find("Size").Raw())
}
