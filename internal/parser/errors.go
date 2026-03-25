package parser

import (
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"fmt"
)

func (p *Parser) blockCannotBeEmptyError(token lexer.Token) error {
	position := tokenToPosition(token)
	return p.emptyBlockErrorAt(position)
}

func (p *Parser) emptyBlockErrorAt(position common.Position) error {
	return common.CompilerError(position, "Block cannot be empty", p.sourceLines)
}

func (p *Parser) expectedGotError(token lexer.Token, expected string) error {
	message := fmt.Sprintf("Expected token '%v', got '%v'", expected, token.Type)
	position := tokenToPosition(token)
	return common.CompilerError(position, message, p.sourceLines)
}

func (p *Parser) unexpectedEOFError() error {
	message := "Unexpected end of input"
	position := tokenToPosition(p.previousToken())
	return common.CompilerErrorEOF(position, message, p.sourceLines)
}

func (p *Parser) unexpectedTokenError(token lexer.Token) error {
	message := fmt.Sprintf("Unexpected token '%v'", token.Value)
	position := tokenToPosition(token)
	return common.CompilerError(position, message, p.sourceLines)
}
