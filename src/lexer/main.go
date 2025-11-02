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
	ElseKeyword      TokenType = "ElseKeyword"
	ReturnKeyword    TokenType = "ReturnKeyword"
	BooleanOperator  TokenType = "BooleanOperator"
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
	elseKeyword         *regexp.Regexp
	returnKeyword       *regexp.Regexp
	booleanOperator     *regexp.Regexp
	openBracket         *regexp.Regexp
	closeBracket        *regexp.Regexp
}

func NewLexer(src bufio.Scanner) *Lexer {
	return &Lexer{
		srcCode:             src,
		lineNum:             0,
		colNum:              0,
		sourceLines:         []string{},
		binaryOperatorChars: regexp.MustCompile(`^(==|&&|\|\||<=|>=|!=|[+\-*/><])[^=&\|]`),
		numberChars:         regexp.MustCompile(`^[0-9]+`),
		whitespaceChars:     regexp.MustCompile(`^[ \t]+`),
		openParenthesis:     regexp.MustCompile(`^\(`),
		closeParenthesis:    regexp.MustCompile(`^\)`),
		identifierChars:     regexp.MustCompile(`^([_A-Za-z][_A-Za-z0-9]*)`),
		assignmentChars:     regexp.MustCompile(`^(=)[^=]`),
		dataType:            regexp.MustCompile(`^(int|i8|i16|i32|i64)[^_A-Za-z0-9]`),
		comma:               regexp.MustCompile(`^,`),
		ifKeyword:           regexp.MustCompile(`^(if)[^_A-Za-z0-9]`),
		elseKeyword:         regexp.MustCompile(`^(else)[^_A-Za-z0-9]`),
		returnKeyword:       regexp.MustCompile(`^(return)[^_A-Za-z0-9]`),
		booleanOperator:     regexp.MustCompile(`^(true|false)[^_A-Za-z0-9]`),
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

func getStringMatch(re *regexp.Regexp, str string) string {
	match := re.FindStringSubmatch(str)
	if len(match) > 1 {
		return match[1]
	}
	return match[0]
}

func (l *Lexer) getNextToken() Token {
	currentLine := l.srcCode.Text()
	nextSubstr := currentLine[l.colNum:]

	var tokenType TokenType
	var value string

	switch {
	case l.binaryOperatorChars.MatchString(nextSubstr):
		value = getStringMatch(l.binaryOperatorChars, nextSubstr)
		tokenType = BinaryOperador
	case l.numberChars.MatchString(nextSubstr):
		value = getStringMatch(l.numberChars, nextSubstr)
		tokenType = Number
	case l.openParenthesis.MatchString(nextSubstr):
		value = getStringMatch(l.openParenthesis, nextSubstr)
		tokenType = OpenParenthesis
	case l.closeParenthesis.MatchString(nextSubstr):
		value = getStringMatch(l.closeParenthesis, nextSubstr)
		tokenType = CloseParenthesis
	case l.openBracket.MatchString(nextSubstr):
		value = getStringMatch(l.openBracket, nextSubstr)
		tokenType = OpenBracket
	case l.closeBracket.MatchString(nextSubstr):
		value = getStringMatch(l.closeBracket, nextSubstr)
		tokenType = CloseBracket
	case l.comma.MatchString(nextSubstr):
		value = getStringMatch(l.comma, nextSubstr)
		tokenType = Comma
	case l.assignmentChars.MatchString(nextSubstr):
		value = getStringMatch(l.assignmentChars, nextSubstr)
		tokenType = Assignment
	case l.ifKeyword.MatchString(nextSubstr):
		value = getStringMatch(l.ifKeyword, nextSubstr)
		tokenType = IfKeyword
	case l.elseKeyword.MatchString(nextSubstr):
		value = getStringMatch(l.elseKeyword, nextSubstr)
		tokenType = ElseKeyword
	case l.returnKeyword.MatchString(nextSubstr):
		value = getStringMatch(l.returnKeyword, nextSubstr)
		tokenType = ReturnKeyword
	case l.booleanOperator.MatchString(nextSubstr):
		value = getStringMatch(l.booleanOperator, nextSubstr)
		tokenType = BooleanOperator
	case l.dataType.MatchString(nextSubstr):
		value = getStringMatch(l.dataType, nextSubstr)
		tokenType = DataType
	case l.identifierChars.MatchString(nextSubstr):
		value = getStringMatch(l.identifierChars, nextSubstr)
		tokenType = Identifier
	case l.whitespaceChars.MatchString(nextSubstr):
		value = getStringMatch(l.whitespaceChars, nextSubstr)
		tokenType = Whitespace
	default:
		log.Panicf("Unknown symbol: '%s' at line %d, column %d\n", nextSubstr, l.lineNum, l.colNum)
	}

	tokenSize := len(value)
	token := Token{Type: tokenType, Value: value, Line: l.lineNum, StartColumn: l.colNum, EndColumn: l.colNum + tokenSize}

	l.colNum += tokenSize

	return token
}
