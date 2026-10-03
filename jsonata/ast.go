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

type CompositionNode struct {
	Base  Node
	Steps []Step
}

func (n *CompositionNode) isNode() {}

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
	Variable     string            // variable name (without $)
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

type ArrayNode struct {
	Elements []Node
}

func (n *ArrayNode) isNode() {}

type ObjectProperty struct {
	Key   Node
	Value Node
}

type ObjectNode struct {
	Properties []ObjectProperty
}

func (n *ObjectNode) isNode() {}

type BlockNode struct {
	Expressions []Node
}

func (n *BlockNode) isNode() {}
