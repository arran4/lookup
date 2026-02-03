package lookup

import (
	"github.com/arran4/go-evaluator"
)

type binaryComparisonFunc struct {
	op    string
	left  Runner
	right Runner
}

func (b *binaryComparisonFunc) Run(scope *Scope) Pathor {
	leftRes := b.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := b.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	expr := evaluator.ComparisonExpression{
		LHS:       evaluator.Constant{Value: leftRes.Raw()},
		RHS:       evaluator.Constant{Value: rightRes.Raw()},
		Operation: b.op,
	}

	result, err := expr.Evaluate(nil)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}

	if result {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func BinaryEquals(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "eq", left: left, right: right}
}

func BinaryNotEquals(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "neq", left: left, right: right}
}

func BinaryGreaterThan(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "gt", left: left, right: right}
}

func BinaryLessThan(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "lt", left: left, right: right}
}

func BinaryGreaterThanOrEqual(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "gte", left: left, right: right}
}

func BinaryLessThanOrEqual(left, right Runner) Runner {
	return &binaryComparisonFunc{op: "lte", left: left, right: right}
}

type binaryInFunc struct {
	left  Runner
	right Runner
}

func BinaryIn(left, right Runner) Runner {
	return &binaryInFunc{left: left, right: right}
}

func (b *binaryInFunc) Run(scope *Scope) Pathor {
	leftRes := b.left.Run(scope)
	if _, ok := leftRes.(*Invalidor); ok {
		return leftRes
	}

	rightRes := b.right.Run(scope)
	if _, ok := rightRes.(*Invalidor); ok {
		return rightRes
	}

	found := elementOf(leftRes.Value(), rightRes.Value(), nil)
	return NewConstantor(scope.Path(), found)
}
