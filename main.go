package main

import (
	"alna-lang/src/lexer"
	"alna-lang/src/parser"
	"bufio"
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalf("Please provide the source code file path as an argument.")
	}

	srcCode, err := os.Open(args[1])
	if err != nil {
		log.Fatalf("Error reading file: %v", err.Error())
	}

	scanner := bufio.NewScanner(srcCode)
	lex := lexer.NewLexer(*scanner)

	tokens, err := lex.Analyze()
	if err != nil {
		log.Panicf("Lexical analysis error: %v", err.Error())
	}

	parser := parser.NewParser(tokens)
	_ = parser.Parse()
}
