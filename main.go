package main

import (
	"alna-lang/src/analyzer"
	"alna-lang/src/codegen"
	"alna-lang/src/disassembler"
	"alna-lang/src/lexer"
	"alna-lang/src/parser"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

var verbose = flag.Bool("verbose", false, "print tokens and AST during compilation")
var disassemble = flag.Bool("disassemble", false, "disassemble bytecode into human-readable format")

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

	analyzer := analyzer.NewAnalyzer(ast, sourceLines)
	analyzer.Analyze()

	if *verbose {
		fmt.Println("\n=== SYMBOL TABLE ===")
		analyzer.PrintSymbolTable()
	}

	codegen := codegen.NewCodeGenerator(ast, sourceLines, analyzer.SymbolTable)
	codegen.Generate()

	if *disassemble {
		fmt.Println()
		fmt.Print(disassembler.Disassemble(codegen.Bytecode))
	} else {
		fmt.Println("\n=== BYTECODE ===")
		for i, b := range codegen.Bytecode {
			fmt.Printf("%04d: 0x%02X\n", i, b)
		}
	}

	os.WriteFile("out.alnbc", codegen.Bytecode, 0644)
}
