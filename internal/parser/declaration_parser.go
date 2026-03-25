package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
)

func (p *Parser) parseVariableDeclaration() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.DataType {
		return nil, p.expectedGotError(token, "data type")
	}

	dataType := token.Value
	identifier := p.advance()

	if identifier.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if identifier.Type != lexer.Identifier {
		return nil, p.expectedGotError(identifier, "identifier")
	}

	token = p.advance()
	if variableInitialization(token) {
		p.advance()

		initializer, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		return ast.VariableDeclarationNode{
			Type:        dataType,
			Name:        identifier.Value,
			Initializer: initializer,
			Position: common.Position{
				Line:      token.Line,
				Column:    token.StartColumn,
				EndLine:   initializer.Pos().EndLine,
				EndColumn: initializer.Pos().EndColumn,
			},
		}, nil
	}

	return ast.VariableDeclarationNode{
		Type: dataType,
		Name: identifier.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   identifier.Line,
			EndColumn: identifier.EndColumn,
		},
	}, nil
}

func variableInitialization(token lexer.Token) bool {
	return token.Type == lexer.Assignment
}

func (p *Parser) parseFunctionDeclaration() (ast.Node, error) {
	token := p.currentToken()
	if token.Type != lexer.DataType {
		return nil, p.expectedGotError(token, "return data type")
	}

	returnType := token.Value
	identifier := p.advance()

	if identifier.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if identifier.Type != lexer.Identifier {
		return nil, p.expectedGotError(identifier, "function name identifier")
	}

	token = p.advance()
	if token.Type != lexer.OpenParenthesis {
		return nil, p.expectedGotError(token, "opening parenthesis")
	}

	token = p.advance()
	parameters, err := p.parseFunctionParametersDeclaration()
	if err != nil {
		return nil, err
	}

	if p.currentToken().Type != lexer.CloseParenthesis {
		return nil, p.expectedGotError(token, "closing parenthesis")
	}

	p.advance()
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}

	return ast.FunctionDeclarationNode{
		Name:       identifier.Value,
		ReturnType: returnType,
		Parameters: parameters,
		Body:       body,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   body.Pos().EndLine,
			EndColumn: body.Pos().EndColumn,
		},
	}, nil
}

func (p *Parser) parseFunctionParametersDeclaration() ([]ast.FunctionParam, error) {
	parameters := []ast.FunctionParam{}
	token := p.currentToken()

	for token.Type != lexer.CloseParenthesis {
		if token.Type == lexer.EOF {
			return nil, p.unexpectedEOFError()
		}

		parameterType := token
		if parameterType.Type != lexer.DataType {
			return nil, p.expectedGotError(parameterType, "parameter data type")
		}

		parameterName := p.advance()
		if parameterName.Type == lexer.EOF {
			return nil, p.unexpectedEOFError()
		}

		if parameterName.Type != lexer.Identifier {
			return nil, p.expectedGotError(parameterName, "parameter name")
		}

		parameters = append(parameters, ast.FunctionParam{
			Type: parameterType.Value,
			Name: parameterName.Value,
		})

		token = p.advance()
		if token.Type == lexer.Comma {
			token = p.advance()
		}
	}

	return parameters, nil
}
