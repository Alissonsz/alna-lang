package parser

import (
	"alna-lang/src/lexer"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateSnapshots = flag.Bool("update", false, "update snapshot files")

// captureASTPrint captures the output of PrintAST
func captureASTPrint(node Node) string {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Temporarily redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Print the AST
	PrintAST(node, "", true)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	io.Copy(&buf, r)
	return buf.String()
}

// snapshotTest runs a snapshot test for parser/AST output
func snapshotTest(t *testing.T, inputFile string) {
	// Read the input file
	file, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("Failed to open input file %s: %v", inputFile, err)
	}
	defer file.Close()

	// Create lexer and analyze
	scanner := bufio.NewScanner(file)
	lex := lexer.NewLexer(*scanner)
	tokens, sourceLines, err := lex.Analyze(false) // verbose = false for tests
	if err != nil {
		t.Fatalf("Lexical analysis failed: %v", err)
	}

	// Parse (may panic on error)
	defer func() {
		if r := recover(); r != nil {
			// Capture panic as error output
			errorMsg := fmt.Sprintf("Parser panicked: %v", r)
			handleSnapshotComparison(t, inputFile, errorMsg)
		}
	}()

	// Create parser and parse
	p := NewParser(tokens, sourceLines)
	ast := p.Parse()

	// Capture AST output
	output := captureASTPrint(ast)

	handleSnapshotComparison(t, inputFile, output)
}

func handleSnapshotComparison(t *testing.T, inputFile string, output string) {
	// Determine snapshot file path
	baseName := filepath.Base(inputFile)
	snapshotName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".ast.snapshot"
	snapshotFile := filepath.Join(filepath.Dir(inputFile), "snapshots", snapshotName)

	// If update flag is set, always write the snapshot
	if *updateSnapshots {
		err := os.WriteFile(snapshotFile, []byte(output), 0644)
		if err != nil {
			t.Fatalf("Failed to write snapshot file: %v", err)
		}
		t.Logf("Updated snapshot: %s", snapshotFile)
		return
	}

	// Read existing snapshot if it exists
	existingSnapshot, err := os.ReadFile(snapshotFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new snapshot
			err = os.WriteFile(snapshotFile, []byte(output), 0644)
			if err != nil {
				t.Fatalf("Failed to write snapshot file: %v", err)
			}
			t.Logf("Created new snapshot: %s", snapshotFile)
			return
		}
		t.Fatalf("Failed to read snapshot file: %v", err)
	}

	// Compare with existing snapshot
	if output != string(existingSnapshot) {
		t.Errorf("Snapshot mismatch for %s\n\nExpected:\n%s\n\nGot:\n%s\n",
			inputFile, string(existingSnapshot), output)
		t.Log("To update snapshots, run: go test -update")
	}
}

func TestParserSnapshots(t *testing.T) {
	// Get all example files
	examplesDir := "../../examples"
	files, err := filepath.Glob(filepath.Join(examplesDir, "*.alna"))
	if err != nil {
		t.Fatalf("Failed to list example files: %v", err)
	}

	if len(files) == 0 {
		t.Fatal("No example files found")
	}

	for _, file := range files {
		// Get just the filename for the test name
		testName := filepath.Base(file)
		t.Run(testName, func(t *testing.T) {
			snapshotTest(t, file)
		})
	}
}
