package jsonata

// AST represents a parsed JSONata expression.
type AST struct {
	Node Node
}

type Node interface {
	isNode()
}

type PathNode struct {
	Steps []Step
}

func (n *PathNode) isNode() {}

type BinaryNode struct {
	Operator string
	Left     Node
	Right    Node
}

func (n *BinaryNode) isNode() {}

type LiteralNode struct {
	Value interface{}
}

func (n *LiteralNode) isNode() {}

type FunctionCallNode struct {
	Name string
	Args []Node
}

func (n *FunctionCallNode) isNode() {}

// Step describes a navigation step in the query.
type Step struct {
	Name         string            // field name
	Index        *int              // optional index
	Filter       *Predicate        // optional filter
	SubExpr      Node              // Parenthesized sub-expression in path
	FunctionCall *FunctionCallNode // Function call as a step
}

// Predicate represents a condition.
type Predicate struct {
	Field    string
	Operator string // "=", ">", "<", etc.
	Value    string
}
