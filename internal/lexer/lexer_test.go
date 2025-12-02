package lexer

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateSnapshots = flag.Bool("update", false, "update snapshot files")

// snapshotTest runs a snapshot test for lexer output
func snapshotTest(t *testing.T, inputFile string) {
	// Read the input file
	file, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("Failed to open input file %s: %v", inputFile, err)
	}
	defer file.Close()

	// Create lexer and analyze
	scanner := bufio.NewScanner(file)
	lex := NewLexer(*scanner)
	tokens, _, err := lex.Analyze(false) // verbose = false for tests
	if err != nil {
		t.Fatalf("Lexical analysis failed: %v", err)
	}

	// Generate output
	var output strings.Builder
	for _, token := range tokens {
		output.WriteString(fmt.Sprintf("%+v\n", token))
	}

	// Determine snapshot file path
	baseName := filepath.Base(inputFile)
	snapshotName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".lexer.snapshot"
	snapshotFile := filepath.Join(filepath.Dir(inputFile), "snapshots", snapshotName)

	// If update flag is set, always write the snapshot
	if *updateSnapshots {
		err = os.WriteFile(snapshotFile, []byte(output.String()), 0644)
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
			err = os.WriteFile(snapshotFile, []byte(output.String()), 0644)
			if err != nil {
				t.Fatalf("Failed to write snapshot file: %v", err)
			}
			t.Logf("Created new snapshot: %s", snapshotFile)
			return
		}
		t.Fatalf("Failed to read snapshot file: %v", err)
	}

	// Compare with existing snapshot
	if output.String() != string(existingSnapshot) {
		t.Errorf("Snapshot mismatch for %s\n\nExpected:\n%s\n\nGot:\n%s\n",
			inputFile, string(existingSnapshot), output.String())
		t.Log("To update snapshots, run: go test -update")
	}
}

func TestLexerSnapshots(t *testing.T) {
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
