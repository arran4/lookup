package main

import (
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCollections(t *testing.T) {
	numbers := []int{1, 2, 3, 3}
	r := lookup.Reflect(numbers)
	assert.Equal(t, []interface{}{1, 2, 3, 4}, r.Find("", lookup.Union(lookup.Array(3, 4))).Raw())
	assert.Equal(t, []interface{}{2, 3}, r.Find("", lookup.Intersection(lookup.Array(2, 3, 4))).Raw())
	assert.Equal(t, 3, r.Find("", lookup.First(lookup.Equals(lookup.Constant(3)))).Raw())
	assert.Equal(t, 3, r.Find("", lookup.Last(lookup.Equals(lookup.Constant(3)))).Raw())
	assert.Equal(t, []int{2, 3}, r.Find("", lookup.Range(1, 3)).Raw())
}
