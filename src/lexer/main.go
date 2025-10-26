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
	Comma            TokenType = "Comma"
	IfKeyword        TokenType = "IfKeyword"
	ReturnKeyword    TokenType = "ReturnKeyword"
	OpenBracket      TokenType = "OpenBracket"
	CloseBracket     TokenType = "CloseBracket"
	EOF              TokenType = "EOF"
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
	sourceLines         []string
	binaryOperatorChars *regexp.Regexp
	numberChars         *regexp.Regexp
	whitespaceChars     *regexp.Regexp
	openParenthesis     *regexp.Regexp
	closeParenthesis    *regexp.Regexp
	identifierChars     *regexp.Regexp
	assignmentChars     *regexp.Regexp
	dataType            *regexp.Regexp
	comma               *regexp.Regexp
	ifKeyword           *regexp.Regexp
	returnKeyword       *regexp.Regexp
	openBracket         *regexp.Regexp
	closeBracket        *regexp.Regexp
}

func NewLexer(src bufio.Scanner) *Lexer {
	return &Lexer{
		srcCode:             src,
		lineNum:             0,
		colNum:              0,
		sourceLines:         []string{},
		binaryOperatorChars: regexp.MustCompile(`^[+\-*/]`),
		numberChars:         regexp.MustCompile(`^[0-9]+`),
		whitespaceChars:     regexp.MustCompile(`^[ \t]+`),
		openParenthesis:     regexp.MustCompile(`^\(`),
		closeParenthesis:    regexp.MustCompile(`^\)`),
		identifierChars:     regexp.MustCompile(`^[_A-Za-z][_A-Za-z0-9]*`),
		assignmentChars:     regexp.MustCompile(`^=`),
		dataType:            regexp.MustCompile(`^(int|i8|i16|i32|i64)[^_A-Za-z0-9,]`),
		comma:               regexp.MustCompile(`^,`),
		ifKeyword:           regexp.MustCompile(`^(if) `),
		returnKeyword:       regexp.MustCompile(`^return`),
		openBracket:         regexp.MustCompile(`^{`),
		closeBracket:        regexp.MustCompile(`^}`),
	}
}

func (l *Lexer) Analyze(verbose bool) ([]Token, []string, error) {
	var tokens []Token

	if verbose {
		fmt.Println("=== TOKENS ===")
	}

	lineTokens := l.consumeLine(verbose)
	for lineTokens != nil {
		tokens = append(tokens, *lineTokens...)
		lineTokens = l.consumeLine(verbose)
	}

	return tokens, l.sourceLines, nil
}

func (l *Lexer) consumeLine(verbose bool) *[]Token {
	if !l.srcCode.Scan() {
		return nil
	}

	l.lineNum++
	l.colNum = 0

	// Store the source line
	currentLine := l.srcCode.Text()
	l.sourceLines = append(l.sourceLines, currentLine)

	var tokens []Token
	for l.colNum < len(currentLine) {
		token := l.getNextToken()
		if token.Type == Whitespace {
			continue
		}

		if verbose {
			fmt.Printf("%+v\n", token)
		}
		tokens = append(tokens, token)
	}

	if verbose && len(tokens) > 0 {
		fmt.Printf("\n")
	}
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
	case l.openBracket.MatchString(nextSubstr):
		value = l.openBracket.FindString(nextSubstr)
		tokenType = OpenBracket
	case l.closeBracket.MatchString(nextSubstr):
		value = l.closeBracket.FindString(nextSubstr)
		tokenType = CloseBracket
	case l.comma.MatchString(nextSubstr):
		value = l.comma.FindString(nextSubstr)
		tokenType = Comma
	case l.assignmentChars.MatchString(nextSubstr):
		value = l.assignmentChars.FindString(nextSubstr)
		tokenType = Assignment
	case l.ifKeyword.MatchString(nextSubstr):
		value = l.ifKeyword.FindString(nextSubstr)
		tokenType = IfKeyword
	case l.returnKeyword.MatchString(nextSubstr):
		value = l.returnKeyword.FindString(nextSubstr)
		tokenType = ReturnKeyword
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
