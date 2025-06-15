package main

import (
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	root := &Node{Name: "root", Size: 10}
	r := lookup.Reflect(root)
	assert.Equal(t, "root", r.Find("Name").Raw())
	assert.Equal(t, 10, r.Find("Size").Raw())
}
