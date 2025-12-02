package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"alna-lang/internal/logger"
	symboltable "alna-lang/internal/symbol_table"
	"fmt"
)

func NewParser(tokens []lexer.Token, sourceLines []string, lgr *logger.Logger) *Parser {
	return &Parser{tokens: tokens, pos: 0, sourceLines: sourceLines, logger: lgr}
}

func (p *Parser) Parse() ast.RootNode {
	program := ast.RootNode{
		Children:    []ast.Node{},
		Position:    common.Position{Line: 1, Column: 0, EndLine: 1, EndColumn: 0},
		SymbolTable: symboltable.NewSymbolTable(nil, true),
	}

	for p.pos < len(p.tokens) {
		program.Children = append(program.Children, p.ParseStatement())
	}

	// Update end position to last child's end position
	if len(program.Children) > 0 {
		lastChild := program.Children[len(program.Children)-1]
		program.Position.EndLine = lastChild.Pos().EndLine
		program.Position.EndColumn = lastChild.Pos().EndColumn
	}

	return program
}

func (p *Parser) parseBlock() ast.BlockNode {
	token := p.tokens[p.pos]
	if token.Type != lexer.OpenBracket {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected '{', got %v", token.Type), p.sourceLines))
	}
	p.pos++

	var statements []ast.Node

	nextToken := p.tokens[p.pos]
	for nextToken.Type != lexer.CloseBracket && p.pos < len(p.tokens) {
		statements = append(statements, p.ParseStatement())
		nextToken = p.tokens[p.pos]
	}

	p.pos++

	return ast.BlockNode{
		Statements: statements,
		Position: common.Position{
			Line:      statements[0].Pos().Line,
			Column:    statements[0].Pos().Column,
			EndLine:   statements[len(statements)-1].Pos().EndLine,
			EndColumn: statements[len(statements)-1].Pos().EndColumn,
		},
	}
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

func (p *Parser) ParseStatement() ast.Node {
	token := p.tokens[p.pos]

	switch token.Type {
	case lexer.IfKeyword:
		return p.parseIfStatement()
	case lexer.DataType:
		return p.parseVariableDeclaration()
	case lexer.Identifier:
		if p.peek().Type == lexer.OpenParenthesis {
			return p.parseFunctionCall()
		}

		if p.peek().Type == lexer.Assignment {
			return p.parseAssignment()
		}
		return p.parseExpression()
	case lexer.OpenParenthesis, lexer.Number, lexer.BooleanOperator:
		return p.parseExpression()
	default:
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Unexpected token '%v'", token.Value), p.sourceLines))
	}
}

func (p *Parser) parseIfStatement() ast.Node {
	token := p.tokens[p.pos]
	if token.Type != lexer.IfKeyword {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected 'if' keyword, got %v", token.Type), p.sourceLines))
	}

	p.pos++
	expression := p.parseExpression()
	var thenBranch *ast.BlockNode

	nextToken := p.tokens[p.pos]
	if nextToken.Type == lexer.OpenBracket {
		block := p.parseBlock()
		thenBranch = &block
	}

	if thenBranch == nil {
		panic(common.CompilerError(tokenToPosition(nextToken), "Expected '{' to start 'then' block", p.sourceLines))
	}

	if len(thenBranch.Statements) == 0 {
		panic(common.CompilerError(thenBranch.Pos(), "'then' block cannot be empty", p.sourceLines))
	}

	var elseBranch *ast.BlockNode

	if p.pos < len(p.tokens) {
		nextToken = p.tokens[p.pos]
		if nextToken.Type == lexer.ElseKeyword {
			p.pos++
			elseBlock := p.parseBlock()
			elseBranch = &elseBlock

			if len(elseBlock.Statements) == 0 {
				panic(common.CompilerError(elseBlock.Pos(), "'else' block cannot be empty", p.sourceLines))
			}
		}
	}

	return ast.IfStatementNode{
		Condition:  expression,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   thenBranch.Pos().EndLine,
			EndColumn: thenBranch.Pos().EndColumn,
		},
	}
}

func (p *Parser) parseParenthised() ast.Node {
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
	return expression
}

func (p *Parser) parseExpression() ast.Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	ast := p.parseLowerPrecedence()

	return ast
}

func (p *Parser) parseLowerPrecedence() ast.Node {
	left := p.parseHigherPrecedence()
	if p.pos >= len(p.tokens) {
		return left
	}

	token := p.tokens[p.pos]
	for token.Type == lexer.BinaryOperador && !isHighPrecedence(token) {
		p.pos++
		right := p.parseHigherPrecedence()
		left = ast.BinaryOpNode{
			Left:     left,
			Operator: token,
			Right:    right,
			Position: common.Position{
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

func (p *Parser) parseHigherPrecedence() ast.Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	var left ast.Node

	switch token.Type {
	case lexer.Number:
		left = p.parseNumber()
	case lexer.BooleanOperator:
		left = p.parseBoolean()
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
		left = ast.BinaryOpNode{
			Left:     left,
			Operator: token,
			Right:    right,
			Position: common.Position{
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

func (p *Parser) parseNumber() ast.Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.Number {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected number, got %v", token.Type), p.sourceLines))
	}
	p.pos++
	return ast.NumberNode{
		Value: token.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}
}

func (p *Parser) parseBoolean() ast.Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.BooleanOperator {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected boolean operator, got %v", token.Type), p.sourceLines))
	}
	p.pos++
	return ast.BooleanNode{
		Value: token.Value == "true",
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}
}

func (p *Parser) parseIdentifier() ast.Node {
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.pos]
	if token.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected identifier, got %v", token.Type), p.sourceLines))
	}
	p.pos++
	return ast.IdentifierNode{
		Name: token.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}
}

func (p *Parser) parseFunctionCall() ast.Node {
	identifierToken := p.tokens[p.pos]
	if identifierToken.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(identifierToken), fmt.Sprintf("Expected identifier, got %v", identifierToken.Type), p.sourceLines))
	}
	p.pos++

	openParenToken := p.tokens[p.pos]
	if openParenToken.Type != lexer.OpenParenthesis {
		panic(common.CompilerError(tokenToPosition(openParenToken), fmt.Sprintf("Expected '(', got %v", openParenToken.Type), p.sourceLines))
	}
	p.pos++

	var args []ast.Node
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type != lexer.CloseParenthesis {
		arg := p.parseExpression()
		args = append(args, arg)

		if p.pos < len(p.tokens) && p.tokens[p.pos].Type == lexer.Comma {
			p.pos++
		}
	}

	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected closing parenthesis ')'", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	closeParenToken := p.tokens[p.pos]
	if closeParenToken.Type != lexer.CloseParenthesis {
		panic(common.CompilerError(tokenToPosition(closeParenToken), fmt.Sprintf("Expected closing parenthesis ')', got %v", closeParenToken.Type), p.sourceLines))
	}
	p.pos++

	return ast.FunctionCallNode{
		Name:      identifierToken.Value,
		Arguments: args,
		Position: common.Position{
			Line:      identifierToken.Line,
			Column:    identifierToken.StartColumn,
			EndLine:   closeParenToken.Line,
			EndColumn: closeParenToken.EndColumn,
		},
	}
}

func isHighPrecedence(op lexer.Token) bool {
	return op.Value == "*" || op.Value == "/"
}
