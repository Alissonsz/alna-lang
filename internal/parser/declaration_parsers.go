package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"fmt"
)

func (p *Parser) parseVariableDeclaration() ast.Node {
	p.logger.Debug("Parsing variable declaration")
	token := p.tokens[p.position]
	if token.Type != lexer.DataType {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected data type, got %v", token.Type), p.sourceLines))
	}

	dataType := token.Value
	p.position++

	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected identifier after data type", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	identifierToken := p.tokens[p.position]
	if identifierToken.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(identifierToken), fmt.Sprintf("Expected identifier, got %v", identifierToken.Type), p.sourceLines))
	}

	p.position++
	if p.position < len(p.tokens) && p.tokens[p.position].Type == lexer.Assignment {
		p.position++
		initializer := p.parseExpression()
		return ast.VariableDeclarationNode{
			Type:        dataType,
			Name:        identifierToken.Value,
			Initializer: initializer,
			Position: common.Position{
				Line:      token.Line,
				Column:    token.StartColumn,
				EndLine:   initializer.Pos().EndLine,
				EndColumn: initializer.Pos().EndColumn,
			},
		}
	}

	return ast.VariableDeclarationNode{
		Type: dataType,
		Name: identifierToken.Value,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   identifierToken.Line,
			EndColumn: identifierToken.EndColumn,
		},
	}
}

func (p *Parser) parseFunctionDeclaration() ast.Node {
	p.logger.Debug("Parsing function declaration")
	token := p.tokens[p.position]
	if token.Type != lexer.DataType {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected return data type, got %v", token.Type), p.sourceLines))
	}

	returnType := token.Value
	p.position++

	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected function name after return data type", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	identifierToken := p.tokens[p.position]
	if identifierToken.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(identifierToken), fmt.Sprintf("Expected function name identifier, got %v", identifierToken.Type), p.sourceLines))
	}

	p.position++
	if p.position >= len(p.tokens) || p.tokens[p.position].Type != lexer.OpenParenthesis {
		panic(common.CompilerError(tokenToPosition(p.tokens[p.position]), "Expected '(' after function name", p.sourceLines))
	}

	p.position++ // Skip '('
	parameters := []ast.FunctionParam{}
	for p.position < len(p.tokens) && p.tokens[p.position].Type != lexer.CloseParenthesis {
		paramTypeToken := p.tokens[p.position]
		if paramTypeToken.Type != lexer.DataType {
			panic(common.CompilerError(tokenToPosition(paramTypeToken), fmt.Sprintf("Expected parameter data type, got %v", paramTypeToken.Type), p.sourceLines))
		}
		paramType := paramTypeToken.Value
		p.position++

		if p.position >= len(p.tokens) {
			panic(common.CompilerErrorEOF("Unexpected end of input, expected parameter name after data type", tokenToPosition(p.lastToken()), p.sourceLines))
		}

		paramNameToken := p.tokens[p.position]
		if paramNameToken.Type != lexer.Identifier {
			panic(common.CompilerError(tokenToPosition(paramNameToken), fmt.Sprintf("Expected parameter name identifier, got %v", paramNameToken.Type), p.sourceLines))
		}
		paramName := paramNameToken.Value
		p.position++

		parameters = append(parameters, ast.FunctionParam{
			Type: paramType,
			Name: paramName,
		})

		if p.position < len(p.tokens) && p.tokens[p.position].Type == lexer.Comma {
			p.position++ // Skip ','
		}
	}

	if p.position >= len(p.tokens) || p.tokens[p.position].Type != lexer.CloseParenthesis {
		panic(common.CompilerError(tokenToPosition(p.tokens[p.position]), "Expected ')' after function parameters", p.sourceLines))
	}

	p.position++ // Skip ')'
	body := p.parseBlock()
	return ast.FunctionDeclarationNode{
		Name:       identifierToken.Value,
		ReturnType: returnType,
		Parameters: parameters,
		Body:       body,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   body.Pos().EndLine,
			EndColumn: body.Pos().EndColumn,
		},
	}
}
