package main

import (
	"log"

	"github.com/arran4/lookup"
)

func main() {
	numbers := []int{1, 2, 3, 3}
	r := lookup.Reflect(numbers)

	union := r.Find("", lookup.Union(lookup.Array(3, 4))).Raw()
	log.Printf("union: %#v", union)

	inter := r.Find("", lookup.Intersection(lookup.Array(2, 3, 4))).Raw()
	log.Printf("intersection: %#v", inter)

	first := r.Find("", lookup.First(lookup.Equals(lookup.Constant(3)))).Raw()
	log.Printf("first 3: %#v", first)

	last := r.Find("", lookup.Last(lookup.Equals(lookup.Constant(3)))).Raw()
	log.Printf("last 3: %#v", last)

	rng := r.Find("", lookup.Range(1, 3)).Raw()
	log.Printf("range [1:3]: %#v", rng)
}
