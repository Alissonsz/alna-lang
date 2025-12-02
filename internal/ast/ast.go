package ast

import (
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	symboltable "alna-lang/internal/symbol_table"
)

// Node is the base interface for all AST nodes
type Node interface {
	NodeType() string
	Pos() common.Position
}

// RootNode represents the root of the AST (the entire program)
type RootNode struct {
	Children    []Node
	Position    common.Position
	SymbolTable *symboltable.SymbolTable
}

func (r RootNode) NodeType() string {
	return "RootNode"
}

func (r RootNode) Pos() common.Position {
	return r.Position
}

// BooleanNode represents a boolean literal (true/false)
type BooleanNode struct {
	Value    bool
	Position common.Position
}

func (b BooleanNode) NodeType() string {
	return "BooleanNode"
}

func (b BooleanNode) Pos() common.Position {
	return b.Position
}

// NumberNode represents a numeric literal
type NumberNode struct {
	Value    interface{}
	Position common.Position
}

func (n NumberNode) NodeType() string {
	return "NumberNode"
}

func (n NumberNode) Pos() common.Position {
	return n.Position
}

// IdentifierNode represents a variable or function reference
type IdentifierNode struct {
	Name     string
	Position common.Position
}

func (i IdentifierNode) NodeType() string {
	return "IdentifierNode"
}

func (i IdentifierNode) Pos() common.Position {
	return i.Position
}

// BinaryOpNode represents a binary operation (e.g., +, -, *, /, ==, <, >)
type BinaryOpNode struct {
	Left     Node
	Operator lexer.Token
	Right    Node
	Position common.Position
}

func (b BinaryOpNode) NodeType() string {
	return "BinaryOpNode"
}

func (b BinaryOpNode) Pos() common.Position {
	return b.Position
}

// AssignmentNode represents variable assignment
type AssignmentNode struct {
	Left     Node
	Right    Node
	Position common.Position
}

func (a AssignmentNode) NodeType() string {
	return "AssignmentNode"
}

func (a AssignmentNode) Pos() common.Position {
	return a.Position
}

// VariableDeclarationNode represents a variable declaration
type VariableDeclarationNode struct {
	Name        string
	Type        string
	Initializer Node
	Position    common.Position
}

func (v VariableDeclarationNode) NodeType() string {
	return "VariableDeclarationNode"
}

func (v VariableDeclarationNode) Pos() common.Position {
	return v.Position
}

// BlockNode represents a block of statements
type BlockNode struct {
	Statements  []Node
	SymbolTable *symboltable.SymbolTable
	Position    common.Position
}

func (b BlockNode) NodeType() string {
	return "BlockNode"
}

func (b BlockNode) Pos() common.Position {
	return b.Position
}

// IfStatementNode represents an if-else statement
type IfStatementNode struct {
	Condition  Node
	ThenBranch Node
	ElseBranch Node
	Position   common.Position
}

func (i IfStatementNode) NodeType() string {
	return "IfExpressionNode"
}

func (i IfStatementNode) Pos() common.Position {
	return i.Position
}

// FunctionCallNode represents a function call
type FunctionCallNode struct {
	Name      string
	Arguments []Node
	Position  common.Position
}

func (f FunctionCallNode) NodeType() string {
	return "FunctionCallNode"
}

func (f FunctionCallNode) Pos() common.Position {
	return f.Position
}
