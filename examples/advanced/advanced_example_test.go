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
			{Name: "child3", Size: 3, Tags: []string{"groupA", "groupC"}},
		},
	}
	r := lookup.Reflect(root)
	assert.Equal(t, []string{"child1", "child2", "child3"}, r.Find("Children").Find("Name").Raw())
	assert.Equal(t, []string{"child1", "child3"}, r.Find("Children", lookup.Filter(
		lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))))).Find("Name").Raw())
	assert.Equal(t, 3, r.Find("Children", lookup.Map(lookup.This("Size")), lookup.Last(nil)).Raw())
	assert.Equal(t, true, r.Find("Children", lookup.Any(lookup.Map(lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupB")))))).Raw())
	assert.Equal(t, []interface{}{"root", "groupA", "groupB", "groupC"}, r.Find("Tags", lookup.Union(lookup.Array("groupB", "groupC"))).Raw())
	assert.Equal(t, []interface{}{"root", "groupA", "groupA"}, r.Find("Tags", lookup.Append(lookup.Array("groupA"))).Raw())
	assert.Equal(t, []interface{}{"groupA"}, r.Find("Tags", lookup.Intersection(lookup.Array("groupA", "groupB"))).Raw())
	assert.Equal(t, "child1", r.Find("Children", lookup.First(nil)).Find("Name").Raw())
	assert.Equal(t, "child3", r.Find("Children", lookup.Last(nil)).Find("Name").Raw())
	assert.Equal(t, []string{"child2", "child3"}, r.Find("Children", lookup.Range(1, nil)).Find("Name").Raw())
}
