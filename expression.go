package lookup

import "reflect"

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
		if v, err := interfaceToBoolOrParse(result.Raw()); err != nil {

		} else if v {
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
	return &toBoolFunc{}
}

type toBoolFunc struct{}

func (s *toBoolFunc) Run(scope *Scope) Pathor {
	b, err := interfaceToBoolOrParse(scope.Position.Raw())
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	return NewConstantor(scope.Path(), b)
}

func Truthy(expression Runner) *truthyFunc {
	return &truthyFunc{}
}

type truthyFunc struct{}

func (s *truthyFunc) Run(scope *Scope) Pathor {
	result := scope.Position
	return truthy(scope, result)
}

func truthy(scope *Scope, result Pathor) Pathor {
	switch result := result.(type) {
	case *Invalidor:
		return result
	}
	v, err := interfaceToBoolOrParse(result.Raw())
	if !v && err == nil {
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
	if reflect.DeepEqual(result.Raw(), scope.Position.Raw()) {
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
	v, err := interfaceToBoolOrParse(result.Raw())
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if !v {
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
	b, err := interfaceToBoolOrParse(c.Raw())
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}
	if b {
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
