package parser

import (
	"alna-lang/src/lexer"
	"fmt"
)

func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0, ast: &RootNode{Children: []Node{}}}
}

func (p *Parser) Parse() *Node {
	ast := p.parseExpression()
	PrintAST(ast, "", true)
	return &ast
}

func (p *Parser) parseExpression() Node {
	if p.pos >= len(p.tokens) {
		panic("Unexpected end of input")
	}

	ast := p.parseLowerPrecedence()

	return ast
}

func (p *Parser) parseHigherPrecedence() Node {
	left := p.parseNumber()
	if p.pos >= len(p.tokens) {
		return left
	}

	token := p.tokens[p.pos]
	for token.Type == lexer.BinaryOperador && isHighPrecedence(token) {
		p.pos++
		right := p.parseNumber()
		left = HighPrecedenceNode{Left: left, Operator: token, Right: right, position: Position{Line: token.LineNum, Column: token.ColNum}}
		if p.pos >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.pos]
	}

	return left
}

func (p *Parser) parseLowerPrecedence() Node {
	left := p.parseHigherPrecedence()
	if p.pos >= len(p.tokens) {
		return left
	}

	token := p.tokens[p.pos]
	for token.Type == lexer.BinaryOperador && !isHighPrecedence(token) {
		p.pos++
		right := p.parseHigherPrecedence()
		left = LowPrecedenceNode{Left: left, Operator: token, Right: right, position: Position{Line: token.LineNum, Column: token.ColNum}}
		if p.pos >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.pos]
	}

	return left
}

func (p *Parser) parseNumber() Node {
	if p.pos >= len(p.tokens) {
		panic("Unexpected end of input")
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.Number {
		panic(fmt.Sprintf("Expected number, got %v \n at %+v", token.Type, Position{Line: token.LineNum, Column: token.ColNum}))
	}
	p.pos++
	return NumberNode{Value: token.Value, position: Position{Line: token.LineNum, Column: token.ColNum}}
}

func isHighPrecedence(op lexer.Token) bool {
	return op.Value == "*" || op.Value == "/"
}

// PrintAST prints the AST in a tree-like visual format
func PrintAST(node Node, indent string, isLast bool) {
	if node == nil {
		return
	}

	// Determine the connector for this node
	connector := ""
	if indent != "" {
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}
	}

	// Print the node based on its type
	switch n := node.(type) {
	case RootNode:
		fmt.Printf("%s%sRoot\n", indent, connector)
		for i, child := range n.Children {
			childIndent := indent
			if indent != "" {
				if isLast {
					childIndent += "    "
				} else {
					childIndent += "│   "
				}
			}
			PrintAST(child, childIndent, i == len(n.Children)-1)
		}
	case NumberNode:
		fmt.Printf("%s%sNumber: %v\n", indent, connector, n.Value)
	case HighPrecedenceNode:
		fmt.Printf("%s%sHighPrecedence (%v)\n", indent, connector, n.Operator.Value)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Left, childIndent, false)
		PrintAST(n.Right, childIndent, true)
	case LowPrecedenceNode:
		fmt.Printf("%s%sLowPrecedence (%v)\n", indent, connector, n.Operator.Value)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Left, childIndent, false)
		PrintAST(n.Right, childIndent, true)
	default:
		fmt.Printf("%s%sUnknown Node Type\n", indent, connector)
	}
}
