package parser

import (
	"alna-lang/src/common"
	"alna-lang/src/lexer"
	"fmt"
)

func (p *Parser) parseVariableDeclaration() Node {
	token := p.tokens[p.pos]
	if token.Type != lexer.DataType {
		panic(common.CompilerError(tokenToPosition(token), fmt.Sprintf("Expected data type, got %v", token.Type), p.sourceLines))
	}

	dataType := token.Value
	p.pos++

	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected identifier after data type", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	identifierToken := p.tokens[p.pos]
	if identifierToken.Type != lexer.Identifier {
		panic(common.CompilerError(tokenToPosition(identifierToken), fmt.Sprintf("Expected identifier, got %v", identifierToken.Type), p.sourceLines))
	}

	p.pos++
	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == lexer.Assignment {
		p.pos++
		initializer := p.parseExpression()
		return VariableDeclarationNode{
			Type:        dataType,
			Name:        identifierToken.Value,
			Initializer: initializer,
			position: common.Position{
				Line:      token.Line,
				Column:    token.StartColumn,
				EndLine:   initializer.Pos().EndLine,
				EndColumn: initializer.Pos().EndColumn,
			},
		}
	}

	return VariableDeclarationNode{
		Type: dataType,
		Name: identifierToken.Value,
		position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   identifierToken.Line,
			EndColumn: identifierToken.EndColumn,
		},
	}
}
