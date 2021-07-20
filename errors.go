package lookup

import "errors"

var (
	ErrNoSuchPath                = errors.New("no such path")
	ErrInvalidEvaluationFunction = errors.New("invalid evaluation function")
	ErrEvalFail                  = errors.New("path succeeded but evaluator failed")
	ErrMatchFail                 = errors.New("path succeeded match failed")
	ErrIndexOfNotArray           = errors.New("tried to index a non-array")
	ErrIndexValueNotValid        = errors.New("index value not valid")
	ErrUnknownIndexMode          = errors.New("unknown index mode")
	ErrIndexOutOfRange           = errors.New("index out of range")
	ErrValueNotIn                = errors.New("value not in set")
	ErrNoMatchesForQuery         = errors.New("nothing matched query")
)
