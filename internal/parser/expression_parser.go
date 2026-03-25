package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
)

func (p *Parser) parseLowerPrecedence() (ast.Node, error) {
	left, err := p.parseHigherPrecedence()
	if err != nil {
		return nil, err
	}

	operator := p.currentToken()
	if operator.Type == lexer.EOF {
		return left, nil
	}

	for isLowPrecedenceOperator(operator) {
		p.advance()

		right, err := p.parseHigherPrecedence()
		if err != nil {
			return nil, err
		}

		left = ast.BinaryOpNode{
			Left:     left,
			Operator: operator,
			Right:    right,
			Position: common.Position{
				Line:      left.Pos().Line,
				Column:    left.Pos().Column,
				EndLine:   right.Pos().EndLine,
				EndColumn: right.Pos().EndColumn,
			},
		}

		operator = p.currentToken()
	}

	return left, nil
}

func (p *Parser) parseHigherPrecedence() (ast.Node, error) {
	token := p.currentToken()
	if token.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	var left ast.Node
	var err error

	switch token.Type {
	case lexer.OpenParenthesis:
		left, err = p.parseParenthised()
	case lexer.Number:
		left, err = p.parseNumber()
	case lexer.BooleanOperator:
		left, err = p.parseBoolean()
	case lexer.Identifier:
		left, err = p.parseIdentifier()
	default:
		return nil, p.unexpectedTokenError(token)
	}

	if err != nil {
		return nil, err
	}

	operator := p.currentToken()
	for isHighPrecedenceOperator(operator) {
		p.advance()

		right, err := p.parseHigherPrecedence()
		if err != nil {
			return nil, err
		}

		left = ast.BinaryOpNode{
			Left:     left,
			Operator: operator,
			Right:    right,
			Position: common.Position{
				Line:      left.Pos().Line,
				Column:    left.Pos().Column,
				EndLine:   right.Pos().EndLine,
				EndColumn: right.Pos().EndColumn,
			},
		}

		operator = p.currentToken()
	}

	return left, nil
}

func (p *Parser) parseParenthised() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.OpenParenthesis {
		return nil, p.expectedGotError(token, "(")
	}

	p.advance()

	expression, err := p.parseBinaryExpression()
	if err != nil {
		return nil, err
	}

	token = p.currentToken()
	if token.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if token.Type != lexer.CloseParenthesis {
		return nil, p.expectedGotError(token, ")")
	}

	p.advance()

	return expression, nil
}

func (p *Parser) parseNumber() (ast.Node, error) {
	token := p.currentToken()
	if token.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if token.Type != lexer.Number {
		return nil, p.expectedGotError(token, "number")
	}

	p.advance()

	return ast.NumberNode{
		Value: token.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}, nil
}

func (p *Parser) parseBoolean() (ast.Node, error) {
	token := p.currentToken()
	if token.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if token.Type != lexer.BooleanOperator {
		return nil, p.expectedGotError(token, "boolean operator")
	}

	p.advance()

	return ast.BooleanNode{
		Value: token.Value == "true",
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}, nil
}

func (p *Parser) parseIdentifier() (ast.Node, error) {
	token := p.currentToken()
	if token.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if token.Type != lexer.Identifier {
		return nil, p.expectedGotError(token, "identifier")
	}

	p.advance()

	return ast.IdentifierNode{
		Name: token.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}, nil
}

func isHighPrecedenceOperator(op lexer.Token) bool {
	return op.Value == "*" || op.Value == "/"
}

func isLowPrecedenceOperator(op lexer.Token) bool {
	return !isHighPrecedenceOperator(op) && op.Type == lexer.BinaryOperador
}
