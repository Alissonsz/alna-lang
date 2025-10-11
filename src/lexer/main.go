package lexer

import (
	"bufio"
	"fmt"
	"log"
)

type TokenType string

const (
	BinaryOperador TokenType = "BinaryOperador"
	Number         TokenType = "Number"
	Whitespace     TokenType = "Whitespace"
)

type Token struct {
	Type    TokenType
	Value   string
	LineNum int
	ColNum  int
}

type Lexer struct {
	srcCode bufio.Scanner
	lineNum int
	colNum  int
}

func NewLexer(src bufio.Scanner) *Lexer {
	return &Lexer{
		srcCode: src,
		lineNum: 0,
		colNum:  0,
	}
}

func (l *Lexer) Analyze() ([]Token, error) {
	var tokens []Token

	lineTokens := l.consumeLine()
	for lineTokens != nil {
		tokens = append(tokens, *lineTokens...)
		lineTokens = l.consumeLine()
	}

	return tokens, nil
}

func (l *Lexer) consumeLine() *[]Token {
	if !l.srcCode.Scan() {
		return nil
	}

	l.lineNum++
	l.colNum = 0

	var tokens []Token
	for l.colNum < len(l.srcCode.Text()) {
		token := l.getNextToken()
		if token.Type == Whitespace {
			l.colNum++
			continue
		}
		fmt.Printf("Token: %+v\n", token)
		tokens = append(tokens, token)
		l.colNum++
	}

	return &tokens
}

func (l *Lexer) getNextToken() Token {
	currentChar := string(l.srcCode.Text()[l.colNum])
	var tokenType TokenType

	switch currentChar {
	case "+", "-", "*", "/":
		tokenType = BinaryOperador
	case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
		tokenType = Number
	case " ", "\t":
		tokenType = Whitespace
	default:
		log.Panicf("Unknown symbol: '%s' at line %d, column %d\n", currentChar, l.lineNum, l.colNum)
	}
	return Token{Type: tokenType, Value: currentChar, LineNum: l.lineNum, ColNum: l.colNum}
}
