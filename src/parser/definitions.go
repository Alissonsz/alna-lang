package parser

import "alna-lang/src/lexer"

type Parser struct {
	tokens []lexer.Token
	pos    int
	ast    *RootNode
}

type Position struct {
	Line   int
	Column int
}

type Node interface {
	NodeType() string
	Pos() Position
}

type RootNode struct {
	Children []Node
	position Position
}

func (r RootNode) NodeType() string {
	return "RootNode"
}

func (r RootNode) Pos() Position {
	return r.position
}

type NumberNode struct {
	Value    interface{}
	position Position
}

func (n NumberNode) NodeType() string {
	return "NumberNode"
}

func (n NumberNode) Pos() Position {
	return n.position
}

type BinaryOpNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position Position
}

func (b BinaryOpNode) NodeType() string {
	return "BinaryOpNode"
}

func (b BinaryOpNode) Pos() Position {
	return b.position
}

type HighPrecedenceNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position Position
}

func (h HighPrecedenceNode) NodeType() string {
	return "HighPrecedenceNode"
}

func (h HighPrecedenceNode) Pos() Position {
	return h.position
}

type LowPrecedenceNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position Position
}

func (l LowPrecedenceNode) NodeType() string {
	return "LowPrecedenceNode"
}

func (l LowPrecedenceNode) Pos() Position {
	return l.position
}
