package parser

import (
	"alna-lang/src/common"
	"alna-lang/src/lexer"
	symboltable "alna-lang/src/symbol_table"
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

type BooleanNode struct {
	Value    bool
	position common.Position
}

func (b BooleanNode) NodeType() string {
	return "BooleanNode"
}

func (b BooleanNode) Pos() common.Position {
	return b.position
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

type BlockNode struct {
	Statements  []Node
	SymbolTable *symboltable.SymbolTable
	position    common.Position
}

func (b BlockNode) NodeType() string {
	return "BlockNode"
}

func (b BlockNode) Pos() common.Position {
	return b.position
}

type IfStatementNode struct {
	Condition  Node
	ThenBranch Node
	ElseBranch Node
	position   common.Position
}

func (i IfStatementNode) NodeType() string {
	return "IfExpressionNode"
}

func (i IfStatementNode) Pos() common.Position {
	return i.position
}
