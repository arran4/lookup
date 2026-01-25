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
