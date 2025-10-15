package parser

import (
	"alna-lang/src/common"
	"alna-lang/src/lexer"
)

type Parser struct {
	tokens      []lexer.Token
	pos         int
	ast         *Node
	sourceLines []string
}

type Node interface {
	NodeType() string
	Pos() common.Position
}

type RootNode struct {
	Children []Node
	position common.Position
}

func (r RootNode) NodeType() string {
	return "RootNode"
}

func (r RootNode) Pos() common.Position {
	return r.position
}

type NumberNode struct {
	Value    interface{}
	position common.Position
}

func (n NumberNode) NodeType() string {
	return "NumberNode"
}

func (n NumberNode) Pos() common.Position {
	return n.position
}

type IdentifierNode struct {
	Name     string
	position common.Position
}

func (i IdentifierNode) NodeType() string {
	return "IdentifierNode"
}

func (i IdentifierNode) Pos() common.Position {
	return i.position
}

type BinaryOpNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position common.Position
}

func (b BinaryOpNode) NodeType() string {
	return "BinaryOpNode"
}

func (b BinaryOpNode) Pos() common.Position {
	return b.position
}

type HighPrecedenceNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position common.Position
}

func (h HighPrecedenceNode) NodeType() string {
	return "HighPrecedenceNode"
}

func (h HighPrecedenceNode) Pos() common.Position {
	return h.position
}

type LowPrecedenceNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	position common.Position
}

func (l LowPrecedenceNode) NodeType() string {
	return "LowPrecedenceNode"
}

func (l LowPrecedenceNode) Pos() common.Position {
	return l.position
}

type ParenthisedNode struct {
	Expression Node
	position   common.Position
}

func (p ParenthisedNode) NodeType() string {
	return "ParenthisedNode"
}

func (p ParenthisedNode) Pos() common.Position {
	return p.position
}

type AssignmentNode struct {
	Left     Node
	Right    Node
	position common.Position
}

func (a AssignmentNode) NodeType() string {
	return "AssignmentNode"
}

func (a AssignmentNode) Pos() common.Position {
	return a.position
}

type VariableDeclarationNode struct {
	Name        string
	Type        string
	Initializer Node
	position    common.Position
}

func (v VariableDeclarationNode) NodeType() string {
	return "VariableDeclarationNode"
}

func (v VariableDeclarationNode) Pos() common.Position {
	return v.position
}
