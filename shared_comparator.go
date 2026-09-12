package lookup

import (
	"fmt"
	"github.com/arran4/go-evaluator"
)

func evaluateComparison(op string, lhs, rhs interface{}) (bool, error) {
	comp, err := evaluator.Compare(lhs, rhs)

	if err != nil {
		return false, err
	}

	switch op {
	case "eq":
		return comp == 0, nil
	case "neq":
		return comp != 0, nil
	case "gt":
		return comp > 0, nil
	case "gte":
		return comp >= 0, nil
	case "lt":
		return comp < 0, nil
	case "lte":
		return comp <= 0, nil
	default:
		return false, fmt.Errorf("unknown operator %s", op)
	}
}
