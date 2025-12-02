package common

// Position represents a location span in source code
// Used across the compiler for tracking locations in AST nodes and error reporting
type Position struct {
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}
