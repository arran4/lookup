package lookup

import ()

type evaluatorComparisonFunc struct {
	op  string
	rhs Runner
}

func (ef *evaluatorComparisonFunc) Run(scope *Scope) Pathor {
	if _, ok := scope.Position.(*Invalidor); ok {
		return scope.Position
	}

	rhsResult := ef.rhs.Run(scope)
	if _, ok := rhsResult.(*Invalidor); ok {
		return rhsResult
	}

	result, err := evaluateComparison(ef.op, scope.Position.Raw(), rhsResult.Raw(), scope.Position)
	if err != nil {
		return NewInvalidor(scope.Path(), err)
	}

	if result {
		return True(scope.Path())
	}
	return False(scope.Path())
}

// evaluateComparison uses evaluator.Compare directly. To ensure operand order matches existing evaluator behavior
// where LHS is the scope / position and RHS is the argument (as verified via the original ComparisonExpression config),
// we flip the arguments so evaluator.Compare(pos, rhs) is correct.
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
