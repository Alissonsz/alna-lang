package parser

import (
	"alna-lang/src/common"
	"alna-lang/src/lexer"
	"fmt"
)

func NewParser(tokens []lexer.Token, sourceLines []string) *Parser {
	return &Parser{tokens: tokens, pos: 0, ast: nil, sourceLines: sourceLines}
}

func (p *Parser) Parse() RootNode {
	program := RootNode{Children: []Node{}, position: common.Position{Line: 1, Column: 0, EndLine: 1, EndColumn: 0}}

	for p.pos < len(p.tokens) {
		program.Children = append(program.Children, p.ParseStatement())
	}

	// Update end position to last child's end position
	if len(program.Children) > 0 {
		lastChild := program.Children[len(program.Children)-1]
		program.position.EndLine = lastChild.Pos().EndLine
		program.position.EndColumn = lastChild.Pos().EndColumn
	}

	// PrintAST(program, "", true) // Debug output disabled for testing
	return program
}

func (p *Parser) peek() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) lastToken() lexer.Token {
	if p.pos > 0 && p.pos <= len(p.tokens) {
		return p.tokens[p.pos-1]
	}
	if len(p.tokens) > 0 {
		return p.tokens[len(p.tokens)-1]
	}
	return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
}

func (p *Parser) ParseStatement() Node {
	token := p.tokens[p.pos]

	switch token.Type {
	case lexer.OpenParenthesis, lexer.Number:
		return p.parseExpression()
	case lexer.DataType:
		return p.parseVariableDeclaration()
	case lexer.Identifier:
		if p.peek().Type == lexer.Assignment {
			return p.parseAssignment()
		}
		return p.parseIdentifier()
	default:
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Unexpected token '%v'", token.Value), p.sourceLines))
	}
}

func (p *Parser) parseParenthised() Node {
	token := p.tokens[p.pos]
	if token.Type != lexer.OpenParenthesis {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected '(', got %v", token.Type), p.sourceLines))
	}

	p.pos++
	expression := p.parseExpression()

	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected closing parenthesis ')'", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	closingToken := p.tokens[p.pos]
	if closingToken.Type != lexer.CloseParenthesis {
		panic(common.CompilerError(tokenToPosition(closingToken), fmt.Sprintf("Expected closing parenthesis ')', got %v", closingToken.Type), p.sourceLines))
	}

	p.pos++
	return ParenthisedNode{
		Expression: expression,
		position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   closingToken.Line,
			EndColumn: closingToken.EndColumn,
		},
	}
}

func (p *Parser) parseExpression() Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	ast := p.parseLowerPrecedence()

	return ast
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
		left = LowPrecedenceNode{
			Left:     left,
			Operator: token,
			Right:    right,
			position: common.Position{
				Line:      left.Pos().Line,
				Column:    left.Pos().Column,
				EndLine:   right.Pos().EndLine,
				EndColumn: right.Pos().EndColumn,
			},
		}
		if p.pos >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.pos]
	}

	return left
}

func (p *Parser) parseHigherPrecedence() Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	var left Node

	switch token.Type {
	case lexer.Number:
		left = p.parseNumber()
	case lexer.Identifier:
		left = p.parseIdentifier()
	case lexer.OpenParenthesis:
		left = p.parseParenthised()
	default:
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Unexpected token '%v'", token.Value), p.sourceLines))
	}

	if p.pos >= len(p.tokens) {
		return left
	}
	token = p.tokens[p.pos]
	for token.Type == lexer.BinaryOperador && isHighPrecedence(token) {
		p.pos++
		right := p.parseExpression()
		left = HighPrecedenceNode{
			Left:     left,
			Operator: token,
			Right:    right,
			position: common.Position{
				Line:      left.Pos().Line,
				Column:    left.Pos().Column,
				EndLine:   right.Pos().EndLine,
				EndColumn: right.Pos().EndColumn,
			},
		}
		if p.pos >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.pos]
	}

	return left
}

func (p *Parser) parseNumber() Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.Number {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected number, got %v", token.Type), p.sourceLines))
	}
	p.pos++
	return NumberNode{
		Value: token.Value,
		position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}
}

func (p *Parser) parseIdentifier() Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected identifier, got %v", token.Type), p.sourceLines))
	}
	p.pos++
	return IdentifierNode{
		Name: token.Value,
		position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}
}

func isHighPrecedence(op lexer.Token) bool {
	return op.Value == "*" || op.Value == "/"
}
