package lookup

import (
	"reflect"

	"github.com/arran4/go-evaluator"
)

type matchFunc struct {
	expressions []Runner
}

func Match(e ...Runner) *matchFunc {
	return &matchFunc{
		expressions: e,
	}
}

func (ef *matchFunc) Run(scope *Scope) Pathor {
	for _, e := range ef.expressions {
		result := e.Run(scope)
		switch result := result.(type) {
		case *Invalidor:
			return result
		}

		expr := evaluator.BoolType{Term: evaluator.Constant{Value: result.Raw()}}
		v, err := expr.Evaluate(nil)

		if err != nil {

		} else if b, ok := v.(bool); ok && b {
			continue
		} else {
			return NewInvalidor(ExtractPath(scope.Position), ErrMatchFail)
		}
		if result.Value().IsZero() {
			return NewInvalidor(ExtractPath(scope.Position), ErrMatchFail)
		}
	}
	return scope.Current
}

func ToBool(expression Runner) *toBoolFunc {
	return &toBoolFunc{
		expression: expression,
	}
}

type toBoolFunc struct {
	expression Runner
}

func (s *toBoolFunc) Run(scope *Scope) Pathor {
	var result Pathor
	if s.expression != nil {
		result = s.expression.Run(scope)
	} else {
		result = scope.Position
	}

	expr := evaluator.BoolType{Term: evaluator.Constant{Value: result.Raw()}}
	v, err := expr.Evaluate(nil)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	return NewConstantor(scope.Path(), v)
}

func Truthy(expression Runner) *truthyFunc {
	return &truthyFunc{
		expression: expression,
	}
}

type truthyFunc struct {
	expression Runner
}

func (s *truthyFunc) Run(scope *Scope) Pathor {
	var result Pathor
	if s.expression != nil {
		result = s.expression.Run(scope)
	} else {
		result = scope.Position
	}
	return truthy(scope, result)
}

func truthy(scope *Scope, result Pathor) Pathor {
	switch result := result.(type) {
	case *Invalidor:
		return result
	}

	expr := evaluator.BoolType{Term: evaluator.Constant{Value: result.Raw()}}
	v, err := expr.Evaluate(nil)

	if b, ok := v.(bool); ok && !b && err == nil {
		return NewInvalidor(scope.Path(), ErrFalse)
	}
	if result.Value().IsZero() {
		return NewInvalidor(scope.Path(), ErrMatchFail)
	}
	return NewConstantor(scope.Path(), v)
}

type equalsFunc struct {
	expression Runner
}

func (ef *equalsFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return equals(scope, result)
}

func equals(scope *Scope, result Pathor) Pathor {
	expr := evaluator.ComparisonExpression{
		LHS:       evaluator.Constant{Value: result.Raw()},
		RHS:       evaluator.Constant{Value: scope.Position.Raw()},
		Operation: "eq",
	}
	if v, _ := expr.Evaluate(nil); v {
		return True(scope.Path())
	} else {
		return False(scope.Path())
	}
}

func Equals(e Runner) *equalsFunc {
	return &equalsFunc{
		expression: e,
	}
}

type notFunc struct {
	expression Runner
}

func (ef *notFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)

	expr := evaluator.BoolType{Term: evaluator.Constant{Value: result.Raw()}}
	v, err := expr.Evaluate(nil)

	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if b, ok := v.(bool); ok && !b {
		return True(scope.Path())
	} else {
		return False(scope.Path())
	}
}

func Not(e Runner) *notFunc {
	return &notFunc{
		expression: e,
	}
}

type isZeroFunc struct {
	expression Runner
}

func IsZero(e Runner) *isZeroFunc {
	return &isZeroFunc{
		expression: e,
	}
}

func (ef *isZeroFunc) Run(scope *Scope) Pathor {
	switch scope.Position.Value().Kind() {
	case reflect.Slice, reflect.Array:
		return arrayOrSliceForEachPath(scope.Path(), nil, scope.Position.Value(), []Runner{ef}, scope)
	}
	result := ef.expression.Run(scope)
	return NewConstantor(scope.Path(), result.Value().IsValid() && result.Value().IsZero())
}

type otherwiseFunc struct {
	expression Runner
}

func (ef *otherwiseFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	position := scope.Position
	switch position := position.(type) {
	case *Constantor:
		i, err := interfaceToBool(position.c)
		if err == nil && !i {
			return result
		}
	case *Invalidor:
		return result
	}
	return position
}

func otherwise(e Runner) *otherwiseFunc {
	return &otherwiseFunc{
		expression: e,
	}
}

// Default used with .Find() as a PathOpt this will will fallback / default to the provided value regardless of future
// nagivations, it suppresses most errors / Invalidators.
func Default(i interface{}) *otherwiseFunc {
	return otherwise(NewConstantor("", i))
}

type ifFunc struct {
	cond      Runner
	then      Runner
	otherwise Runner
}

// If evaluates `cond` and runs `then` when true otherwise `otherwise`.
func If(cond Runner, then Runner, otherwise Runner) *ifFunc {
	return &ifFunc{
		cond:      cond,
		then:      then,
		otherwise: otherwise,
	}
}

func (ef *ifFunc) Run(scope *Scope) Pathor {
	c := ef.cond.Run(scope)
	if invalid, ok := c.(*Invalidor); ok {
		return invalid
	}

	expr := evaluator.BoolType{Term: evaluator.Constant{Value: c.Raw()}}
	v, err := expr.Evaluate(nil)

	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if b, ok := v.(bool); ok && b {
		if ef.then != nil {
			return ef.then.Run(scope)
		}
		return scope.Position
	}
	if ef.otherwise != nil {
		return ef.otherwise.Run(scope)
	}
	return scope.Position
}

type fallbackPathsFunc struct {
	paths []string
}

func FallbackPaths(paths ...string) *fallbackPathsFunc {
	return &fallbackPathsFunc{
		paths: paths,
	}
}

func (ef *fallbackPathsFunc) Run(scope *Scope) Pathor {
	if _, ok := scope.Position.(*Invalidor); !ok {
		return scope.Position
	}

	if scope.Parent == nil || scope.Parent.Current == nil {
		return scope.Position
	}

	parent := scope.Parent.Current

	for _, path := range ef.paths {
		res := parent.Find(path)
		if _, ok := res.(*Invalidor); !ok {
			return res
		}
	}

	return scope.Position
}

type andFunc struct {
	left  Runner
	right Runner
}

func And(left, right Runner) *andFunc {
	return &andFunc{left: left, right: right}
}

func (a *andFunc) Run(scope *Scope) Pathor {
	leftRes := a.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	isTrue := false
	expr := evaluator.BoolType{Term: evaluator.Constant{Value: leftRes.Raw()}}
	v, err := expr.Evaluate(nil)
	if err == nil {
		if b, ok := v.(bool); ok && b {
			isTrue = true
		}
	}
	if leftRes.Value().IsValid() && leftRes.Value().IsZero() {
		isTrue = false
	}

	if isTrue {
		return a.right.Run(scope)
	}
	return leftRes
}

type orFunc struct {
	left  Runner
	right Runner
}

func Or(left, right Runner) *orFunc {
	return &orFunc{left: left, right: right}
}

func (o *orFunc) Run(scope *Scope) Pathor {
	leftRes := o.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	isTrue := false
	expr := evaluator.BoolType{Term: evaluator.Constant{Value: leftRes.Raw()}}
	v, err := expr.Evaluate(nil)
	if err == nil {
		if b, ok := v.(bool); ok && b {
			isTrue = true
		}
	}
	if leftRes.Value().IsValid() && leftRes.Value().IsZero() {
		isTrue = false
	}

	if isTrue {
		return leftRes
	}
	return o.right.Run(scope)
}
