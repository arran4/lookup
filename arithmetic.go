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
		return NewConstantor(scope.Path(), f1 + f2)
	}

	return NewInvalidor(scope.Path(), fmt.Errorf("type mismatch or invalid types for add: %T, %T", v1, v2))
}

func Add(left, right Runner) *addFunc {
	return &addFunc{
		left:  left,
		right: right,
	}
}
