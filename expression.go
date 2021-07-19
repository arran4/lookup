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

func (ef *matchFunc) Run(scope *Scope, position Pathor) Pathor {
	var outcome Pathor = nil
	for _, e := range ef.expressions {
		result := e.Run(scope, position)
		if v, err := interfaceToBoolOrParse(result.Raw()); err != nil && v {
			return position
		}
		switch result := result.(type) {
		case *Invalidor:
			{
				outcome = result
			}
		}
	}
	if outcome != nil {
		return outcome
	}
	return position
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

type equalsFunc struct {
	expression Runner
}

func (ef *equalsFunc) Run(scope *Scope, position Pathor) Pathor {
	result := ef.expression.Run(scope, position)
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

func (ef *notFunc) Run(scope *Scope, position Pathor) Pathor {
	result := ef.expression.Run(scope, position)
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

func (ef *isZeroFunc) Run(scope *Scope, position Pathor) Pathor {
	result := ef.expression.Run(scope, position)
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

func (ef *otherwiseFunc) Run(scope *Scope, position Pathor) Pathor {
	result := ef.expression.Run(scope, position)
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
