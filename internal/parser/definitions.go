package parser

import (
	"alna-lang/internal/lexer"
	"alna-lang/internal/logger"
)

type Parser struct {
	tokens      []lexer.Token
	position    int
	sourceLines []string
	logger      *logger.Logger
}

func (p *Parser) StoppedAt() lexer.Token {
	return p.currentToken()
}

func (p *Parser) currentToken() lexer.Token {
	if p.position >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
	}
	return p.tokens[p.position]
}

func (p *Parser) nextToken() lexer.Token {
	if p.position+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
	}
	return p.tokens[p.position+1]
}

func (p *Parser) previousToken() lexer.Token {
	if p.position > 0 && p.position <= len(p.tokens) {
		return p.tokens[p.position-1]
	}
	if len(p.tokens) > 0 {
		return p.tokens[len(p.tokens)-1]
	}
	return lexer.Token{Type: lexer.EOF, Value: "", Line: -1, StartColumn: -1, EndColumn: -1}
}

func (p *Parser) advance() lexer.Token {
	if p.position < len(p.tokens) {
		p.position++
	}
	return p.currentToken()
}
