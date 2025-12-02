package main

import (
	"alna-lang/internal/analyzer"
	"alna-lang/internal/ast"
	"alna-lang/internal/codegen"
	"alna-lang/internal/disassembler"
	"alna-lang/internal/lexer"
	"alna-lang/internal/logger"
	"alna-lang/internal/parser"
	"alna-lang/internal/vm"
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

	// Create logger based on verbose flag
	var logLevel logger.LogLevel
	if *verbose {
		logLevel = logger.LevelDebug
	} else {
		logLevel = logger.LevelInfo
	}
	lgr := logger.New(logLevel, *verbose)

	srcCode, err := os.Open(args[0])
	if err != nil {
		log.Fatalf("Error reading file: %v", err.Error())
	}

	scanner := bufio.NewScanner(srcCode)
	lex := lexer.NewLexer(*scanner, lgr)

	tokens, sourceLines, err := lex.Analyze()
	if err != nil {
		log.Panicf("Lexical analysis error: %v", err.Error())
	}

	p := parser.NewParser(tokens, sourceLines, lgr)
	tree := p.Parse()

	if *verbose {
		fmt.Println("\n=== AST ===")
		ast.PrintAST(tree, "", true)
	}

	analyzer := analyzer.NewAnalyzer(&tree, sourceLines, lgr)
	analyzer.Analyze()

	if *verbose {
		fmt.Println("\n=== SYMBOL TABLE ===")
		analyzer.PrintSymbolTable()
	}

	codegen := codegen.NewCodeGenerator(tree, sourceLines, analyzer.SymbolTable, lgr)
	codegen.Generate()

	if *disassemble {
		fmt.Println()
		fmt.Print(disassembler.Disassemble(codegen.Bytecode))
	} else if *verbose {
		lgr.Println("\n=== BYTECODE ===")
		for i, b := range codegen.Bytecode {
			lgr.Print("%04d: 0x%02X\n", i, b)
		}
	}

	os.WriteFile("out.alnbc", codegen.Bytecode, 0644)

	// Debug mode is disabled for now (TUI requires interactive terminal)
	vm := vm.NewVM(codegen.Bytecode, sourceLines, false, lgr)
	err = vm.CheckHeader()
	if err != nil {
		log.Panicf("VM header check failed: %v", err.Error())
	}

	err = vm.Run()
	if err != nil {
		log.Panicf("VM runtime error: %v", err.Error())
	}
}
