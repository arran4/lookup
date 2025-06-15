package main

import (
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdvanced(t *testing.T) {
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
	assert.Equal(t, []string{"child1", "child2"}, r.Find("Children").Find("Name").Raw())
	assert.Equal(t, []string{"child1"}, r.Find("Children", lookup.Filter(
		lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))))).Find("Name").Raw())
	assert.Equal(t, 2, r.Find("Children", lookup.Map(lookup.This("Size")), lookup.Index("-1")).Raw())
	assert.Equal(t, true, r.Find("Children", lookup.Any(lookup.Map(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupB")))))).Raw())
}
