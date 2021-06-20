package lookup

import "errors"

var (
	ErrNoSuchPath                = errors.New("no such path")
	ErrInvalidEvaluationFunction = errors.New("invalid evaluation function")
	ErrEvalFail                  = errors.New("path succeeded but evaluator failed")
)
