package lookup

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapModifier(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	result := r.Find("", Map(Constant(5))).Raw()
	assert.Equal(t, []int{5, 5, 5}, result)
}

func TestMapModifierIdentity(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	result := r.Find("", Map(This())).Raw()
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestUnionModifier(t *testing.T) {
	r := Reflect([]int{1, 2})
	result := r.Find("", Union(Array(2, 3))).Raw()
	assert.Equal(t, []interface{}{1, 2, 3}, result)
}

func TestUnionModifierDeduplicate(t *testing.T) {
	r := Reflect([]int{1, 2})
	result := r.Find("", Union(Array(2, 2, 3, 1))).Raw()
	assert.Equal(t, []interface{}{1, 2, 3}, result)
}

func TestAppendModifier(t *testing.T) {
	r := Reflect([]int{1, 2})
	result := r.Find("", Append(Array(3, 4))).Raw()
	assert.Equal(t, []interface{}{1, 2, 3, 4}, result)
}

func TestAppendModifierWithDuplicates(t *testing.T) {
	r := Reflect([]int{1, 2})
	result := r.Find("", Append(Array(2, 3))).Raw()
	assert.Equal(t, []interface{}{1, 2, 2, 3}, result)
}

func TestIntersectionModifier(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	result := r.Find("", Intersection(Array(2, 4))).Raw()
	assert.Equal(t, []interface{}{2}, result)
}

func TestIntersectionModifierNoMatch(t *testing.T) {
	r := Reflect([]int{1, 2})
	result := r.Find("", Intersection(Array(3, 4))).Raw()
	assert.Equal(t, []interface{}{}, result)
}

func TestFirstModifier(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	res := r.Find("", First(nil)).Raw()
	assert.Equal(t, 1, res)
	res2 := r.Find("", First(Equals(Constant(2)))).Raw()
	assert.Equal(t, 2, res2)
}

func TestFirstModifierNoMatch(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	res := r.Find("", First(Equals(Constant(5))))
	_, ok := res.(*Invalidor)
	assert.True(t, ok)
}

func TestLastModifier(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	res := r.Find("", Last(nil)).Raw()
	assert.Equal(t, 3, res)
	r2 := Reflect([]int{1, 2, 2, 3})
	res2 := r2.Find("", Last(Equals(Constant(2)))).Raw()
	assert.Equal(t, 2, res2)
}

func TestLastModifierNoMatch(t *testing.T) {
	r := Reflect([]int{1, 2, 3})
	res := r.Find("", Last(Equals(Constant(5))))
	_, ok := res.(*Invalidor)
	assert.True(t, ok)
}

func TestRangeModifier(t *testing.T) {
	r := Reflect([]int{1, 2, 3, 4})
	res := r.Find("", Range(1, 3)).Raw()
	assert.Equal(t, []int{2, 3}, res)
}

func TestRangeModifierAll(t *testing.T) {
	r := Reflect([]int{1, 2, 3, 4})
	res := r.Find("", Range(nil, nil)).Raw()
	assert.Equal(t, []int{1, 2, 3, 4}, res)
}

func TestRangeModifierNegativeIndexes(t *testing.T) {
	r := Reflect([]int{1, 2, 3, 4})
	res := r.Find("", Range(-3, -1)).Raw()
	assert.Equal(t, []int{2, 3}, res)
}
