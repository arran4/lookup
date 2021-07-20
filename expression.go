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
	switch result := result.(type) {
	case *Invalidor:
		return result
	}
	v, err := interfaceToBoolOrParse(result.Raw())
	if !v && err == nil {
		return NewInvalidor(ExtractPath(scope.Position), ErrMatchFail)
	}
	if result.Value().IsZero() {
		return NewInvalidor(ExtractPath(scope.Position), ErrMatchFail)
	}
	return NewConstantor(scope.Path(), v)
}

type equalsFunc struct {
	expression Runner
}

func (ef *equalsFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	if reflect.DeepEqual(result.Raw(), scope.Value().Interface()) {
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

func (ef *isZeroFunc) Run(scope *Scope) Pathor {
	result := ef.expression.Run(scope)
	return NewConstantor(scope.Path(), result.Value().IsZero())
}

func IsZero(e Runner) *isZeroFunc {
	return &isZeroFunc{
		expression: e,
	}
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

// NewDefault used with .Find() as a PathOpt this will will fallback / default to the provided value regardless of future
// nagivations, it suppresses most errors / Invalidators.
func NewDefault(i interface{}) *otherwiseFunc {
	return otherwise(NewConstantor("", i))
}

//TODO
//type ifFunc struct {
//	cond Runner
//	then Runner
//	otherwise Runner
//}
//
//func If(e ...Runner) *ifFunc {
//	return &ifFunc{
//		cond: e,
//		then: e,
//		otherwise: e,
//	}
//}
//
//func (ef *ifFunc) Run(scope *Scope, position Pathor) Pathor {
//}
