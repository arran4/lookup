package lookup

import (
	"reflect"
	"testing"

	"github.com/arran4/go-evaluator"
	"github.com/stretchr/testify/assert"
)

func TestScope_Copy_NilReceiver(t *testing.T) {
	var s *Scope
	copied := s.Copy()
	assert.Nil(t, copied)
}

func TestScope_Copy_PreservesInvariants(t *testing.T) {
	parent := &Scope{
		Current: Simple("parent_current"),
	}

	pathStr := "some.path"
	ctx := &evaluator.Context{Variables: map[string]interface{}{"key": "val"}}

	s := &Scope{
		Current:  Simple("current_val"),
		Parent:   parent,
		v:        reflect.ValueOf(42),
		path:     &pathStr,
		Position: Simple("position_val"),
		Context:  ctx,
	}

	copied := s.Copy()

	// Assertions for fields that should be preserved exactly
	assert.NotNil(t, copied)
	assert.Same(t, s.Current, copied.Current, "Current should be the same")
	assert.Same(t, s.Parent, copied.Parent, "Parent should be the same")
	assert.Equal(t, s.v, copied.v, "reflect.Value should be equal")
	assert.Same(t, s.path, copied.path, "path pointer should be the same")
	assert.Same(t, s.Position, copied.Position, "Position should be the same")
	assert.Same(t, s.Context, copied.Context, "Context should be the same")

	// Behavior assertions
	assert.Equal(t, s.Path(), copied.Path(), "Path() output should match")
	assert.Equal(t, s.Value().Int(), copied.Value().Int(), "Value() output should match")

	// Ensure changes in behavior from copied don't mutate context maps or similar unexpectedly
	if copied.Context != nil && copied.Context.Variables != nil {
		assert.Equal(t, "val", copied.Context.Variables["key"])
	}
}
