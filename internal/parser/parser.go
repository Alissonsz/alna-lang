package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"alna-lang/internal/logger"
	symboltable "alna-lang/internal/symbol_table"
)

func NewParser(tokens []lexer.Token, sourceLines []string, lgr *logger.Logger) *Parser {
	return &Parser{tokens: tokens, position: 0, sourceLines: sourceLines, logger: lgr}
}

func (p *Parser) Parse() (ast.RootNode, error) {
	program := ast.RootNode{
		Children:    []ast.Node{},
		Position:    common.Position{Line: 1, Column: 0, EndLine: 1, EndColumn: 0},
		SymbolTable: symboltable.NewSymbolTable(nil, true),
	}

	for p.position < len(p.tokens) {
		expression, err := p.parseExpression()
		program.Children = append(program.Children, expression)

		if err != nil {
			return program, err
		}
	}

	if len(program.Children) > 0 {
		lastPosition := len(program.Children) - 1
		lastChild := program.Children[lastPosition]

		program.Position.EndLine = lastChild.Pos().EndLine
		program.Position.EndColumn = lastChild.Pos().EndColumn
	}

	return program, nil
}

func (p *Parser) parseExpression() (ast.Node, error) {
	token := p.currentToken()

	switch token.Type {
	case lexer.IfKeyword:
		return p.parseIfExpression()
	case lexer.DataType:
		return p.parseDeclaration()
	case lexer.Identifier:
		return p.parseIdentifierUsage()
	case lexer.OpenParenthesis, lexer.Number, lexer.BooleanOperator:
		return p.parseBinaryExpression()
	case lexer.ReturnKeyword:
		return p.parseReturn()
	default:
		return nil, p.unexpectedTokenError(token)
	}
}

func (p *Parser) parseIfExpression() (ast.Node, error) {
	ifToken := p.currentToken()
	if ifToken.Type != lexer.IfKeyword {
		return nil, p.expectedGotError(ifToken, "if")
	}
	p.advance()

	condition, err := p.parseBinaryExpression()
	if err != nil {
		return nil, err
	}

	thenBranch, err := p.parseConditionedBlock()
	if err != nil {
		return nil, err
	}

	var elseBranch *ast.BlockNode
	if p.currentToken().Type == lexer.ElseKeyword {
		p.advance()

		elseBranch, err = p.parseConditionedBlock()
		if err != nil {
			return nil, err
		}
	}

	return ast.IfExpressionNode{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
		Position: common.Position{
			Line:      ifToken.Line,
			Column:    ifToken.StartColumn,
			EndLine:   thenBranch.Pos().EndLine,
			EndColumn: thenBranch.Pos().EndColumn,
		},
	}, nil
}

func (p *Parser) parseConditionedBlock() (*ast.BlockNode, error) {
	var conditionedBlock *ast.BlockNode
	block, err := p.parseBlock()
	if err != nil {
		return &ast.BlockNode{}, err
	}

	conditionedBlock = &block
	if len(conditionedBlock.Expressions) == 0 {
		return &ast.BlockNode{}, p.emptyBlockErrorAt(conditionedBlock.Pos())
	}

	return conditionedBlock, nil
}

func (p *Parser) parseBlock() (ast.BlockNode, error) {
	token := p.currentToken()
	if token.Type != lexer.OpenBracket {
		return ast.BlockNode{}, p.expectedGotError(token, "{")
	}
	p.advance()

	var expressions []ast.Node
	for p.currentToken().Type != lexer.CloseBracket {
		if p.currentToken().Type == lexer.EOF {
			return ast.BlockNode{}, p.unexpectedEOFError()
		}

		expression, err := p.parseExpression()
		if err != nil {
			return ast.BlockNode{}, err
		}

		expressions = append(expressions, expression)
	}
	p.advance()

	if len(expressions) == 0 {
		return ast.BlockNode{}, p.blockCannotBeEmptyError(token)
	}

	return ast.BlockNode{
		Expressions: expressions,
		Position: common.Position{
			Line:      expressions[0].Pos().Line,
			Column:    expressions[0].Pos().Column,
			EndLine:   expressions[len(expressions)-1].Pos().EndLine,
			EndColumn: expressions[len(expressions)-1].Pos().EndColumn,
		},
	}, nil
}

func (p *Parser) parseDeclaration() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.DataType {
		return nil, p.expectedGotError(token, "data type")
	}

	nextToken := p.nextToken()
	if nextToken.Type != lexer.Identifier {
		return nil, p.expectedGotError(nextToken, "identifier")
	}

	afterNextToken := p.tokens[p.position+2]
	if afterNextToken.Type == lexer.OpenParenthesis {
		return p.parseFunctionDeclaration()
	}

	return p.parseVariableDeclaration()
}

func (p *Parser) parseIdentifierUsage() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.Identifier {
		return nil, p.expectedGotError(token, "identifier")
	}

	if p.nextToken().Type == lexer.OpenParenthesis {
		return p.parseFunctionCall()
	}

	if p.nextToken().Type == lexer.Assignment {
		return p.parseAssignment()
	}
	return p.parseBinaryExpression()
}

func (p *Parser) parseBinaryExpression() (ast.Node, error) {
	if p.currentToken().Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	return p.parseLowerPrecedence()
}

func (p *Parser) parseReturn() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.ReturnKeyword {
		return nil, p.expectedGotError(token, "return")
	}
	p.advance()

	return ast.ReturnNode{
		Value: nil,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   token.Line,
			EndColumn: token.EndColumn,
		},
	}, nil
}

func (p *Parser) parseFunctionCall() (ast.Node, error) {
	identifier := p.currentToken()
	if identifier.Type != lexer.Identifier {
		return nil, p.expectedGotError(identifier, "identifier")
	}
	p.advance()

	openParenthesis := p.currentToken()
	if openParenthesis.Type != lexer.OpenParenthesis {
		return nil, p.expectedGotError(openParenthesis, "(")
	}
	p.advance()

	arguments, err := p.parseFunctionArguments()
	if err != nil {
		return nil, err
	}

	closeParenthesis := p.currentToken()
	if closeParenthesis.Type != lexer.CloseParenthesis {
		return nil, p.expectedGotError(closeParenthesis, ")")
	}
	p.advance()

	return ast.FunctionCallNode{
		Name:      identifier.Value,
		Arguments: arguments,
		Position: common.Position{
			Line:      identifier.Line,
			Column:    identifier.StartColumn,
			EndLine:   closeParenthesis.Line,
			EndColumn: closeParenthesis.EndColumn,
		},
	}, nil
}

func (p *Parser) parseFunctionArguments() ([]ast.Node, error) {
	var arguments []ast.Node

	for p.currentToken().Type != lexer.CloseParenthesis {
		if p.currentToken().Type == lexer.EOF {
			return nil, p.unexpectedEOFError()
		}

		arg, err := p.parseBinaryExpression()
		if err != nil {
			return nil, err
		}

		arguments = append(arguments, arg)

		if p.currentToken().Type == lexer.Comma {
			p.advance()
		}
	}

	return arguments, nil
}
