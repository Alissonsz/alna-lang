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
	return &Parser{tokens: tokens, position: 0, sourceLines: sourceLines, logger: lgr}
}

func (p *Parser) Parse() ast.RootNode {
	program := ast.RootNode{
		Children:    []ast.Node{},
		Position:    common.Position{Line: 1, Column: 0, EndLine: 1, EndColumn: 0},
		SymbolTable: symboltable.NewSymbolTable(nil, true),
	}

	for p.position < len(p.tokens) {
		program.Children = append(program.Children, p.parseExpression())
	}

	if len(program.Children) > 0 {
		lastChild := program.Children[len(program.Children)-1]
		program.Position.EndLine = lastChild.Pos().EndLine
		program.Position.EndColumn = lastChild.Pos().EndColumn
	}

	return program
}

func (p *Parser) parseBlock() ast.BlockNode {
	token := p.tokens[p.position]
	if token.Type != lexer.OpenBracket {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected '{', got %v", token.Type), p.sourceLines))
	}
	p.position++

	var expressions []ast.Node

	nextToken := p.tokens[p.position]
	for nextToken.Type != lexer.CloseBracket && p.position < len(p.tokens) {
		expressions = append(expressions, p.parseExpression())
		nextToken = p.tokens[p.position]
	}

	p.position++

	return ast.BlockNode{
		Expressions: expressions,
		Position: common.Position{
			Line:      expressions[0].Pos().Line,
			Column:    expressions[0].Pos().Column,
			EndLine:   expressions[len(expressions)-1].Pos().EndLine,
			EndColumn: expressions[len(expressions)-1].Pos().EndColumn,
		},
	}
}

func (p *Parser) peek() lexer.Token {
	if p.position+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
	}
	return p.tokens[p.position+1]
}

func (p *Parser) lastToken() lexer.Token {
	if p.position > 0 && p.position <= len(p.tokens) {
		return p.tokens[p.position-1]
	}
	if len(p.tokens) > 0 {
		return p.tokens[len(p.tokens)-1]
	}
	return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
}

func (p *Parser) parseExpression() ast.Node {
	token := p.tokens[p.position]

	switch token.Type {
	case lexer.IfKeyword:
		return p.parseIfExpression()
	case lexer.DataType:
		nextToken := p.peek()
		if nextToken.Type == lexer.Identifier {
			afterNextToken := p.tokens[p.position+2]
			if afterNextToken.Type == lexer.OpenParenthesis {
				return p.parseFunctionDeclaration()
			}
		}
		return p.parseVariableDeclaration()
	case lexer.Identifier:
		if p.peek().Type == lexer.OpenParenthesis {
			return p.parseFunctionCall()
		}

		if p.peek().Type == lexer.Assignment {
			return p.parseAssignment()
		}
		return p.parseBinaryExpression()
	case lexer.OpenParenthesis, lexer.Number, lexer.BooleanOperator:
		return p.parseBinaryExpression()
	case lexer.ReturnKeyword:
		p.position++
		return ast.ReturnNode{
			Value: nil,
			Position: common.Position{
				Line:      token.Line,
				Column:    token.StartColumn,
				EndLine:   token.Line,
				EndColumn: token.EndColumn,
			},
		}
	default:
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Unexpected token '%v'", token.Value), p.sourceLines))
	}
}

func (p *Parser) parseIfExpression() ast.Node {
	token := p.tokens[p.position]
	if token.Type != lexer.IfKeyword {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected 'if' keyword, got %v", token.Type), p.sourceLines))
	}

	p.position++
	condition := p.parseBinaryExpression()
	var thenBranch *ast.BlockNode

	nextToken := p.tokens[p.position]
	if nextToken.Type == lexer.OpenBracket {
		block := p.parseBlock()
		thenBranch = &block
	}

	if thenBranch == nil {
		panic(common.CompilerError(tokenToPosition(nextToken), "Expected '{' to start 'then' block", p.sourceLines))
	}

	if len(thenBranch.Expressions) == 0 {
		panic(common.CompilerError(thenBranch.Pos(), "'then' block cannot be empty", p.sourceLines))
	}

	var elseBranch *ast.BlockNode

	if p.position < len(p.tokens) {
		nextToken = p.tokens[p.position]
		if nextToken.Type == lexer.ElseKeyword {
			p.position++
			elseBlock := p.parseBlock()
			elseBranch = &elseBlock

			if len(elseBlock.Expressions) == 0 {
				panic(common.CompilerError(elseBlock.Pos(), "'else' block cannot be empty", p.sourceLines))
			}
		}
	}

	return ast.IfExpressionNode{
		Condition:  condition,
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
	token := p.tokens[p.position]
	if token.Type != lexer.OpenParenthesis {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected '(', got %v", token.Type), p.sourceLines))
	}

	p.position++
	expression := p.parseBinaryExpression()

	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected closing parenthesis ')'", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	closingToken := p.tokens[p.position]
	if closingToken.Type != lexer.CloseParenthesis {
		panic(common.CompilerError(tokenToPosition(closingToken), fmt.Sprintf("Expected closing parenthesis ')', got %v", closingToken.Type), p.sourceLines))
	}

	p.position++
	return expression
}

func (p *Parser) parseBinaryExpression() ast.Node {
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	return p.parseLowerPrecedence()
}

func (p *Parser) parseLowerPrecedence() ast.Node {
	left := p.parseHigherPrecedence()
	if p.position >= len(p.tokens) {
		return left
	}

	token := p.tokens[p.position]
	for token.Type == lexer.BinaryOperador && !isHighPrecedence(token) {
		p.position++
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
		if p.position >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.position]
	}

	return left
}

func (p *Parser) parseHigherPrecedence() ast.Node {
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.position]
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

	if p.position >= len(p.tokens) {
		return left
	}
	token = p.tokens[p.position]
	for token.Type == lexer.BinaryOperador && isHighPrecedence(token) {
		p.position++
		right := p.parseBinaryExpression()
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
		if p.position >= len(p.tokens) {
			return left
		}
		token = p.tokens[p.position]
	}

	return left
}

func (p *Parser) parseNumber() ast.Node {
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.position]
	if token.Type != lexer.Number {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected number, got %v", token.Type), p.sourceLines))
	}
	p.position++
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
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.position]
	if token.Type != lexer.BooleanOperator {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected boolean operator, got %v", token.Type), p.sourceLines))
	}
	p.position++
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
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	token := p.tokens[p.position]
	if token.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected identifier, got %v", token.Type), p.sourceLines))
	}
	p.position++
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
	p.logger.Debug("Parsing function call at token: %v", p.tokens[p.position])
	identifierToken := p.tokens[p.position]
	if identifierToken.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(identifierToken), fmt.Sprintf("Expected identifier, got %v", identifierToken.Type), p.sourceLines))
	}
	p.position++

	openParenToken := p.tokens[p.position]
	if openParenToken.Type != lexer.OpenParenthesis {
		panic(common.CompilerError(tokenToPosition(openParenToken), fmt.Sprintf("Expected '(', got %v", openParenToken.Type), p.sourceLines))
	}
	p.position++

	var args []ast.Node
	for p.position < len(p.tokens) && p.tokens[p.position].Type != lexer.CloseParenthesis {
		arg := p.parseBinaryExpression()
		args = append(args, arg)

		if p.position < len(p.tokens) && p.tokens[p.position].Type == lexer.Comma {
			p.position++
		}
	}

	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected closing parenthesis ')'", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	closeParenToken := p.tokens[p.position]
	if closeParenToken.Type != lexer.CloseParenthesis {
		panic(common.CompilerError(tokenToPosition(closeParenToken), fmt.Sprintf("Expected closing parenthesis ')', got %v", closeParenToken.Type), p.sourceLines))
	}
	p.position++

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
