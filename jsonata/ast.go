package jsonata

// AST represents a parsed JSONata expression.
type AST struct {
	Steps []Step
}

// Step describes a navigation step in the query.
type Step struct {
	Name   string     // field name
	Index  *int       // optional index
	Filter *Predicate // optional equality filter
}

// Predicate represents a simple equality condition.
type Predicate struct {
	Field string
	Value string
}
