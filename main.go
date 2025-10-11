package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type TokenType string

const (
	BinaryOperador TokenType = "BinaryOperador"
	Number         TokenType = "Number"
)

type Token struct {
	Type    string
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
		lineNum: 1,
		colNum:  0,
	}
}

func (l *Lexer) getNextToken() string {
	// check if we already started the scanner
	if l.colNum == 0 {
		if !l.srcCode.Scan() {
			return ""
		}
	}

	// check if we got to the end of the line
	if l.colNum > len(l.srcCode.Text())-1 {
		l.lineNum++
		l.colNum = 0

		if !l.srcCode.Scan() {
			return ""
		}
	}

	currentChar := l.getCurrentChar()
	if currentChar == 0 {
		return ""
	}

	l.colNum++

	return string(currentChar)
}

func (l *Lexer) getCurrentChar() byte {
	text := l.srcCode.Text()
	if l.colNum > len(text)-1 {
		return 0
	}
	return text[l.colNum]
}

func main() {
	srcCode, err := os.Open("exp.as")
	if err != nil {
		log.Fatalf("Error reading file: %v", err.Error())
	}

	scanner := bufio.NewScanner(srcCode)
	lexer := NewLexer(*scanner)

	token := lexer.getNextToken()
	for token != "" {
		fmt.Printf("Token: %+v\n", token)
		token = lexer.getNextToken()
	}
}
