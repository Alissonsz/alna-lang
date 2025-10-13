package main

import (
	"alna-lang/src/lexer"
	"alna-lang/src/parser"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

var verbose = flag.Bool("verbose", false, "print tokens and AST during compilation")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatalf("Please provide the source code file path as an argument.")
	}

	srcCode, err := os.Open(args[0])
	if err != nil {
		log.Fatalf("Error reading file: %v", err.Error())
	}

	scanner := bufio.NewScanner(srcCode)
	lex := lexer.NewLexer(*scanner)

	tokens, sourceLines, err := lex.Analyze(*verbose)
	if err != nil {
		log.Panicf("Lexical analysis error: %v", err.Error())
	}

	p := parser.NewParser(tokens, sourceLines)
	ast := p.Parse()

	if *verbose {
		fmt.Println("\n=== AST ===")
		parser.PrintAST(ast, "", true)
	}
}
