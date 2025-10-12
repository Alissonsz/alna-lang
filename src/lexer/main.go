package lexer

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
)

type TokenType string

const (
	BinaryOperador   TokenType = "BinaryOperador"
	Number           TokenType = "Number"
	Whitespace       TokenType = "Whitespace"
	OpenParenthesis  TokenType = "OpenParenthesis"
	CloseParenthesis TokenType = "CloseParenthesis"
	Identifier       TokenType = "Identifier"
	Assignment       TokenType = "Assignment"
	DataType         TokenType = "DataType"
)

type Token struct {
	Type        TokenType
	Value       string
	Line        int
	StartColumn int
	EndColumn   int
}

type Lexer struct {
	srcCode             bufio.Scanner
	lineNum             int
	colNum              int
	binaryOperatorChars *regexp.Regexp
	numberChars         *regexp.Regexp
	whitespaceChars     *regexp.Regexp
	openParenthesis     *regexp.Regexp
	closeParenthesis    *regexp.Regexp
	identifierChars     *regexp.Regexp
	assignmentChars     *regexp.Regexp
	dataType            *regexp.Regexp
}

func NewLexer(src bufio.Scanner) *Lexer {
	return &Lexer{
		srcCode:             src,
		lineNum:             0,
		colNum:              0,
		binaryOperatorChars: regexp.MustCompile(`^[+\-*/]`),
		numberChars:         regexp.MustCompile(`^[0-9]+`),
		whitespaceChars:     regexp.MustCompile(`^[ \t]+`),
		openParenthesis:     regexp.MustCompile(`^\(`),
		closeParenthesis:    regexp.MustCompile(`^\)`),
		identifierChars:     regexp.MustCompile(`^[_A-Za-z][_A-Za-z0-9]*`),
		assignmentChars:     regexp.MustCompile(`^=`),
		dataType:            regexp.MustCompile(`^int|i8|i16|i32|i64`),
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
			continue
		}

		fmt.Printf("%+v\n", token)
		tokens = append(tokens, token)
	}

	fmt.Printf("\n")
	return &tokens
}

func (l *Lexer) getNextToken() Token {
	currentLine := l.srcCode.Text()
	nextSubstr := currentLine[l.colNum:]

	var tokenType TokenType
	var value string

	switch {
	case l.binaryOperatorChars.MatchString(nextSubstr):
		value = l.binaryOperatorChars.FindString(nextSubstr)
		tokenType = BinaryOperador
	case l.numberChars.MatchString(nextSubstr):
		value = l.numberChars.FindString(nextSubstr)
		tokenType = Number
	case l.openParenthesis.MatchString(nextSubstr):
		value = l.openParenthesis.FindString(nextSubstr)
		tokenType = OpenParenthesis
	case l.closeParenthesis.MatchString(nextSubstr):
		value = l.closeParenthesis.FindString(nextSubstr)
		tokenType = CloseParenthesis
	case l.assignmentChars.MatchString(nextSubstr):
		value = l.assignmentChars.FindString(nextSubstr)
		tokenType = Assignment
	case l.dataType.MatchString(nextSubstr):
		value = l.dataType.FindString(nextSubstr)
		tokenType = DataType
	case l.identifierChars.MatchString(nextSubstr):
		value = l.identifierChars.FindString(nextSubstr)
		tokenType = Identifier
	case l.whitespaceChars.MatchString(nextSubstr):
		value = l.whitespaceChars.FindString(nextSubstr)
		tokenType = Whitespace
	default:
		log.Panicf("Unknown symbol: '%s' at line %d, column %d\n", nextSubstr, l.lineNum, l.colNum)
	}

	tokenSize := len(value)
	token := Token{Type: tokenType, Value: value, Line: l.lineNum, StartColumn: l.colNum, EndColumn: l.colNum + tokenSize}

	l.colNum += tokenSize

	return token
}
