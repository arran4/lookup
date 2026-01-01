package jsonata

// AST represents a parsed JSONata expression.
type AST struct {
	Steps []Step
}

// Step describes a navigation step in the query.
type Step struct {
	Name      string     // field name
	Index     *int       // optional index
	Filter    *Predicate // optional filter
	Value     string     // if it's a literal value
	IsLiteral bool
	Operator  string     // operator preceding this step (e.g. "+")
}

// Predicate represents a condition.
type Predicate struct {
	Field    string
	Operator string // "=", ">", "<", etc.
	Value    string
}
