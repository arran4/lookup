package main

import (
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

func TestIfExample(t *testing.T) {
	root := &Node{Name: "child1", Tags: []string{"groupA"}}
	r := lookup.Reflect(root)
	v := r.Find("", lookup.If(
		lookup.This("Tags").Find("", lookup.Contains(lookup.Constant("groupA"))),
		lookup.This("Name"),
		lookup.Constant("other"),
	)).Raw()
	assert.Equal(t, "child1", v)
}
