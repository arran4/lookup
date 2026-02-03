package lookup

import (
	"errors"
)

type binaryLogicFunc struct {
	op    string
	left  Runner
	right Runner
}

func BinaryAnd(left, right Runner) Runner {
	return &binaryLogicFunc{op: "and", left: left, right: right}
}

func BinaryOr(left, right Runner) Runner {
	return &binaryLogicFunc{op: "or", left: left, right: right}
}

func (b *binaryLogicFunc) Run(scope *Scope) Pathor {
	leftRes := b.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	tRes := truthy(scope, leftRes)
	isTruthy := true
	if inv, ok := tRes.(*Invalidor); ok {
		if errors.Is(inv, ErrFalse) {
			isTruthy = false
		} else {
			return tRes
		}
	}

	if b.op == "and" {
		if !isTruthy {
			return NewConstantor(scope.Path(), false)
		}

		rightRes := b.right.Run(scope)
		if _, ok := rightRes.(*Invalidor); ok {
			return rightRes
		}

		tResRight := truthy(scope, rightRes)
		isTruthyRight := true
		if inv, ok := tResRight.(*Invalidor); ok {
			if errors.Is(inv, ErrFalse) {
				isTruthyRight = false
			} else {
				return tResRight
			}
		}
		return NewConstantor(scope.Path(), isTruthyRight)

	} else { // or
		if isTruthy {
			return NewConstantor(scope.Path(), true)
		}

		rightRes := b.right.Run(scope)
		if _, ok := rightRes.(*Invalidor); ok {
			return rightRes
		}

		tResRight := truthy(scope, rightRes)
		isTruthyRight := true
		if inv, ok := tResRight.(*Invalidor); ok {
			if errors.Is(inv, ErrFalse) {
				isTruthyRight = false
			} else {
				return tResRight
			}
		}
		return NewConstantor(scope.Path(), isTruthyRight)
	}
}
