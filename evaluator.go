package lookup

import (
	"reflect"
)

type Evaluate interface {
	Evaluate(scope *Scope, position Pathor) (Pathor, error)
}

// Evaluator wraps either a interface or a function. It uses reflection to match type as much as possible
type Evaluator struct {
	fi    interface{}
	group bool
}

func (e *Evaluator) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, e)
}

func NewEvaluator(fi interface{}) *Evaluator {
	return &Evaluator{
		fi: fi,
	}
}

func (e *Evaluator) Evaluate(scope *Scope, position Pathor) (Pathor, error) {
	if e == e.fi || e.fi == nil {
		return position, nil
	}
	switch position.Type().Kind() {
	case reflect.Array, reflect.Slice:
		if !e.group {
			p := ExtractPath(position)
			evaluators := []Evaluate{}
			if e, ok := e.fi.(Evaluate); ok {
				evaluators = append(evaluators, e)
			}
			v := arrayOrSliceForEachPath(p, nil, position.Value(), &PathSettings{}, evaluators, scope)
			if !v.Value().IsValid() || v.Value().IsZero() || v.Value().Len() == 0 {
				return nil, nil
			}
			return v, nil
		}
		fallthrough
	default:
		if e, ok := e.fi.(Evaluate); ok {
			return e.Evaluate(scope, position)
		}
	}
	return nil, NewInvalidor(ExtractPath(position), ErrInvalidEvaluationFunction)
}

func (e *Evaluator) evalArray(scope *Scope, position Pathor) ([]reflect.Value, Pathor, error) {
	result := []reflect.Value{}
	for i := 0; i < position.Value().Len(); i++ {
		if e, ok := e.fi.(Evaluate); ok {
			ee, err := e.Evaluate(scope, position)
			if err != nil {
				return nil, nil, err
			}
			if ee != nil {
				switch ee.(type) {
				case *Invalidor:
				default:
					result = append(result, ee.Value())
				}
			}
		}
	}
	return result, nil, nil
}

type Scope struct {
	Current Pathor
}
