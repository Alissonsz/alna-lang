package parser

import (
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
)

// tokenToPosition converts a lexer.Token to a common.Position
// This is a convenience function for the parser to easily convert tokens to positions for error reporting
func tokenToPosition(token lexer.Token) common.Position {
	return common.Position{
		Line:      token.Line,
		Column:    token.StartColumn,
		EndLine:   token.Line,
		EndColumn: token.EndColumn,
	}
}
