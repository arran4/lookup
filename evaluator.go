package lookup

// Evaluator wraps either a interface or a function. It uses reflection to match type as much as possible
type Evaluator struct {
	fi interface{}
}

func (e *Evaluator) PathOptSet(settings *PathSettings) {
	settings.Evaluators = append(settings.Evaluators, e)
}

func NewEvaluator(fi interface{}) *Evaluator {
	return &Evaluator{
		fi: fi,
	}
}

func (e *Evaluator) Evaluate(position Pathor) (bool, error) {
	if e == e.fi || e.fi == nil {
		return true, nil
	}
	if e, ok := e.fi.(EvaluateFromPosition); ok {
		return e.Evaluate(position), nil
	}
	if e, ok := e.fi.(EvaluateFromPositionError); ok {
		return e.Evaluate(position)
	}
	if e, ok := e.fi.(EvaluateNoArg); ok {
		return e.Evaluate(), nil
	}
	if e, ok := e.fi.(EvaluateFromPositionFunc); ok {
		return e(position), nil
	}
	if e, ok := e.fi.(EvaluateFromPositionErrorFunc); ok {
		return e(position)
	}
	if e, ok := e.fi.(EvaluateNoArgFunc); ok {
		return e(), nil
	}
	return false, NewInvalidor(ExtractPath(position), ErrInvalidEvaluationFunction)
}

func Where(fi interface{}) PathOpt {
	return NewEvaluator(fi)
}

type EvaluateFromPosition interface {
	Evaluate(position Pathor) bool
}
type EvaluateFromPositionError interface {
	Evaluate(position Pathor) (bool, error)
}
type EvaluateNoArg interface {
	Evaluate() bool
}

type EvaluateFromPositionFunc func(position Pathor) bool
type EvaluateFromPositionErrorFunc func(position Pathor) (bool, error)
type EvaluateNoArgFunc func() bool
