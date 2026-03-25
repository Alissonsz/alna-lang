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
var debug = flag.Bool("tui", false, "run with TUI debugger (generates .alnbc.debug file)")

func main() {
	flag.Parse()
	args := flag.Args()
	sourceFile := args[0]

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
	lexical := lexer.NewLexer(*scanner)
	tokens, sourceLines, err := lexical.Analyze()

	if *verbose {
		ll := lgr.WithStep("lexer")

		ll.Println("\n=== TOKENS ===")
		for _, token := range tokens {
			ll.Debug("%+v", token)
		}
		ll.Println()
	}

	if err != nil {
		log.Panicf("Lexical analysis error: %v", err.Error())
	}

	syntax := parser.NewParser(tokens, sourceLines, lgr.WithStep("parser"))
	tree, err := syntax.Parse()

	if *verbose {
		lp := lgr.WithStep("parser")

		lp.Println("\n=== AST ===")
		ast.PrintAST(tree, "", true)
	}

	if err != nil {
		log.Panicf("Syntax analysis error: %v", err.Error())
	}

	semantic := analyzer.NewAnalyzer(&tree, sourceLines, lgr.WithStep("analyzer"))
	semantic.Analyze()

	if *verbose {
		fmt.Println("\n=== SYMBOL TABLE ===")
		semantic.PrintSymbolTable()
	}

	codegen := codegen.NewCodeGenerator(tree, sourceLines, semantic.SymbolTable, lgr.WithStep("codegen"))

	if *debug {
		codegen.SetDebugMode(sourceFile)
	}
	codegen.Generate()

	if *disassemble {
		fmt.Println()
		fmt.Print(disassembler.Disassemble(codegen.Bytecode))
	}

	if *verbose {
		lgr.Println("\n=== BYTECODE ===")
		for i, b := range codegen.Bytecode {
			lgr.Print("%04d: 0x%02X\n", i, b)
		}
	}

	os.WriteFile("out.alnac", codegen.Bytecode, 0644)

	if *debug {
		if err := codegen.WriteDebugFile("out.alnac.debug"); err != nil {
			log.Fatalf("Failed to write debug file: %v", err)
		}
	}

	vm := vm.NewVM(codegen.Bytecode, sourceLines, *debug, lgr.WithStep("vm"))

	if *debug {
		if err := vm.LoadDebugFile("out.alnac.debug"); err != nil {
			log.Fatalf("Failed to load debug file: %v", err)
		}
	}

	err = vm.CheckHeader()
	if err != nil {
		log.Panicf("VM header check failed: %v", err.Error())
	}

	err = vm.Run()
	if err != nil {
		log.Panicf("VM runtime error: %v", err.Error())
	}
}
