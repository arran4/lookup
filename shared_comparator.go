package lookup

import (
	"fmt"
	"github.com/arran4/go-evaluator"
)

func evaluateComparison(op string, lhs, rhs interface{}, pos Pathor) (bool, error) {
	comp, err := evaluator.Compare(lhs, rhs)

	// Workaround for go-evaluator v0.0.2 behavior:
	// Compare(A, B) does not correctly propagate the error when B implements Comparator but A doesn't,
	// instead it returns -1, <nil>. We detect if RHS is the source of truth and retry properly if needed.
	if err == nil {
		if _, ok := lhs.(evaluator.Comparator); !ok {
			if _, ok := rhs.(evaluator.Comparator); ok {
				// To get the actual error from RHS, we need to call Compare on it directly
				_, err = evaluator.Compare(rhs, lhs)
			}
		}
	}

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
