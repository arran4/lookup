package lookup

import (
	"reflect"

	"github.com/arran4/go-evaluator"
)

type greaterThanFunc struct {
	expression Runner
}

func (ef *greaterThanFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return greaterThan(scope, result)
}

func greaterThan(scope *Scope, result Pathor) Pathor {
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	if c, _ := evaluator.Compare(v1, v2); c > 0 {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func GreaterThan(e Runner) *greaterThanFunc {
	return &greaterThanFunc{
		expression: e,
	}
}

type lessThanFunc struct {
	expression Runner
}

func (ef *lessThanFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return lessThan(scope, result)
}

func lessThan(scope *Scope, result Pathor) Pathor {
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	if c, _ := evaluator.Compare(v1, v2); c < 0 {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func LessThan(e Runner) *lessThanFunc {
	return &lessThanFunc{
		expression: e,
	}
}

type greaterThanOrEqualFunc struct {
	expression Runner
}

func (ef *greaterThanOrEqualFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	if c, _ := evaluator.Compare(v1, v2); c >= 0 {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func GreaterThanOrEqual(e Runner) *greaterThanOrEqualFunc {
	return &greaterThanOrEqualFunc{expression: e}
}

type lessThanOrEqualFunc struct {
	expression Runner
}

func (ef *lessThanOrEqualFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	v1 := scope.Position.Raw()
	v2 := result.Raw()

	if c, _ := evaluator.Compare(v1, v2); c <= 0 {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func LessThanOrEqual(e Runner) *lessThanOrEqualFunc {
	return &lessThanOrEqualFunc{expression: e}
}

type notEqualsFunc struct {
	expression Runner
}

func (ef *notEqualsFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	if !reflect.DeepEqual(result.Raw(), scope.Position.Raw()) {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func NotEquals(e Runner) *notEqualsFunc {
	return &notEqualsFunc{expression: e}
}
