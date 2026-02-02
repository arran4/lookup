package lookup

import (
	"github.com/arran4/go-evaluator"
)

type evaluatorComparisonFunc struct {
	op  string
	rhs Runner
}

func (ef *evaluatorComparisonFunc) Run(scope *Scope) Pathor {
	rhsResult := ef.rhs.Run(scope)

	expr := evaluator.ComparisonExpression{
		LHS:       evaluator.Self{},
		RHS:       evaluator.Constant{Value: rhsResult.Raw()},
		Operation: ef.op,
	}

	result, _ := expr.Evaluate(scope.Position.Raw())
	if result {
		return True(scope.Path())
	}
	return False(scope.Path())
}

func GreaterThan(e Runner) *evaluatorComparisonFunc {
	return &evaluatorComparisonFunc{op: "gt", rhs: e}
}

func LessThan(e Runner) *evaluatorComparisonFunc {
	return &evaluatorComparisonFunc{op: "lt", rhs: e}
}

func GreaterThanOrEqual(e Runner) *evaluatorComparisonFunc {
	return &evaluatorComparisonFunc{op: "gte", rhs: e}
}

func LessThanOrEqual(e Runner) *evaluatorComparisonFunc {
	return &evaluatorComparisonFunc{op: "lte", rhs: e}
}

func NotEquals(e Runner) *evaluatorComparisonFunc {
	return &evaluatorComparisonFunc{op: "neq", rhs: e}
}

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

func BinaryEquals(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "eq", left: left, right: right}
}

func BinaryNotEquals(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "neq", left: left, right: right}
}

func BinaryGreaterThan(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "gt", left: left, right: right}
}

func BinaryLessThan(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "lt", left: left, right: right}
}

func BinaryGreaterThanOrEqual(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "gte", left: left, right: right}
}

func BinaryLessThanOrEqual(left, right Runner) *binaryComparisonFunc {
	return &binaryComparisonFunc{op: "lte", left: left, right: right}
}

type binaryInFunc struct {
	left  Runner
	right Runner
}

func BinaryIn(left, right Runner) *binaryInFunc {
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
