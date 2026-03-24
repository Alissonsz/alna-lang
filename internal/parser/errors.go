package parser

import (
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"fmt"
)

func (p *Parser) blockCannotBeEmptyError(token lexer.Token) error {
	message := "Block cannot be empty"
	position := tokenToPosition(token)
	return common.CompilerError(position, message, p.sourceLines)
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
