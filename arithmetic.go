package lookup

import (
	"fmt"
)

type addFunc struct {
	left  Runner
	right Runner
}

func (ef *addFunc) Run(scope *Scope) Pathor {
	leftRes := ef.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := ef.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	v1 := leftRes.Raw()
	v2 := rightRes.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		return NewConstantor(scope.Path(), f1+f2)
	}

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for add: %T, %T", v1, v2))
}

func Add(left, right Runner) *addFunc {
	return &addFunc{
		left:  left,
		right: right,
	}
}

func (ef *addFunc) Name() string {
	return "add"
}

type subtractFunc struct {
	left  Runner
	right Runner
}

func (ef *subtractFunc) Run(scope *Scope) Pathor {
	leftRes := ef.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := ef.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	v1 := leftRes.Raw()
	v2 := rightRes.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		return NewConstantor(scope.Path(), f1-f2)
	}

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for subtract: %T, %T", v1, v2))
}

func Subtract(left, right Runner) *subtractFunc {
	return &subtractFunc{
		left:  left,
		right: right,
	}
}

type multiplyFunc struct {
	left  Runner
	right Runner
}

func (ef *multiplyFunc) Run(scope *Scope) Pathor {
	leftRes := ef.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := ef.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	v1 := leftRes.Raw()
	v2 := rightRes.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		return NewConstantor(scope.Path(), f1*f2)
	}

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for multiply: %T, %T", v1, v2))
}

func Multiply(left, right Runner) *multiplyFunc {
	return &multiplyFunc{
		left:  left,
		right: right,
	}
}

type divideFunc struct {
	left  Runner
	right Runner
}

func (ef *divideFunc) Run(scope *Scope) Pathor {
	leftRes := ef.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := ef.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	v1 := leftRes.Raw()
	v2 := rightRes.Raw()

	f1, err1 := interfaceToFloat(v1)
	f2, err2 := interfaceToFloat(v2)

	if err1 == nil && err2 == nil {
		if f2 == 0 {
			return NewInvalidor(scope.Path(), fmt.Errorf("division by zero"))
		}
		return NewConstantor(scope.Path(), f1/f2)
	}

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for divide: %T, %T", v1, v2))
}

func Divide(left, right Runner) *divideFunc {
	return &divideFunc{
		left:  left,
		right: right,
	}
}

type moduloFunc struct {
	left  Runner
	right Runner
}

func (ef *moduloFunc) Run(scope *Scope) Pathor {
	leftRes := ef.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := ef.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	v1 := leftRes.Raw()
	v2 := rightRes.Raw()

	i1, ok1 := ToInt(v1)
	i2, ok2 := ToInt(v2)

	if ok1 && ok2 {
		if i2 == 0 {
			return NewInvalidor(scope.Path(), fmt.Errorf("modulo by zero"))
		}
		return NewConstantor(scope.Path(), i1%i2)
	}

	// Fallback to float mod if needed, but JSONata usually treats mod as integer operation or math.mod?
	// The specification says JSONata `%` operator.
	// Let's implement generic float mod if ToInt fails but float conversion works?
	// Go's % is only for integers.
	// We can use math.Mod for floats if we really want to support it, but keeping it int-only for now is safer/easier.
	// Actually, let's use math.Mod?
	// No, let's just error if not int for now based on typical usage, or maybe check ToInt implementation.

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for modulo: %T, %T", v1, v2))
}

func Modulo(left, right Runner) *moduloFunc {
	return &moduloFunc{
		left:  left,
		right: right,
	}
}
