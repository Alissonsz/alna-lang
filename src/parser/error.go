package parser

import (
	"alna-lang/src/lexer"
	"fmt"
	"strings"
)

// Error reporting functions for the parser:
//
// - ParserError: Use for single-token errors (most common case)
// - ParserErrorWithPosition: Use for multi-line construct errors (if/for/while blocks, etc.)
// - ParserErrorEOF: Use when unexpectedly reaching end of input
// - ParserErrorSimple: Use when no source location is available

// ParserError creates a formatted error message with source code context
func ParserError(token lexer.Token, message string, sourceLines []string) string {
	var sb strings.Builder

	// Error header
	sb.WriteString(fmt.Sprintf("\n\033[1;31mParser Error:\033[0m %s\n", message))
	sb.WriteString(fmt.Sprintf("\033[36mAt line %d, column %d\033[0m\n\n", token.Line, token.StartColumn))

	// Show the source line if available
	if token.Line > 0 && token.Line <= len(sourceLines) {
		lineNum := token.Line

		// Calculate the maximum line number width for alignment
		contextLines := 2 // lines before and after
		startLine := lineNum - contextLines
		if startLine < 1 {
			startLine = 1
		}
		endLine := lineNum + contextLines
		if endLine > len(sourceLines) {
			endLine = len(sourceLines)
		}

		maxLineNumWidth := len(fmt.Sprintf("%d", endLine))

		// Show context lines before
		for i := startLine; i < lineNum; i++ {
			sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
		}

		// Show the error line
		line := sourceLines[lineNum-1]
		sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, lineNum, line))

		// Show the pointer line
		sb.WriteString(fmt.Sprintf("  %s | ", strings.Repeat(" ", maxLineNumWidth)))

		// Add spacing before the pointer
		if token.StartColumn > 0 {
			sb.WriteString(strings.Repeat(" ", token.StartColumn))
		}

		// Add the pointer
		pointerLength := token.EndColumn - token.StartColumn
		if pointerLength < 1 {
			pointerLength = 1
		}
		sb.WriteString("\033[1;31m")
		sb.WriteString(strings.Repeat("^", pointerLength))
		sb.WriteString("\033[0m\n")

		// Show context lines after
		for i := lineNum + 1; i <= endLine; i++ {
			sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
		}
	}

	return sb.String()
}

// ParserErrorSimple creates a simple error message without source code context
// Used when we don't have a specific token or when EOF is reached
func ParserErrorSimple(message string) string {
	return fmt.Sprintf("\n\033[1;31mParser Error:\033[0m %s\n", message)
}

// ParserErrorWithPosition creates an error message highlighting a Position span
// This is useful for multi-line constructs like control flow statements
func ParserErrorWithPosition(pos Position, message string, sourceLines []string) string {
	var sb strings.Builder

	// Error header
	sb.WriteString(fmt.Sprintf("\n\033[1;31mParser Error:\033[0m %s\n", message))
	sb.WriteString(fmt.Sprintf("\033[36mAt line %d, column %d to line %d, column %d\033[0m\n\n",
		pos.Line, pos.Column, pos.EndLine, pos.EndColumn))

	// Calculate context
	contextLines := 2
	startLine := pos.Line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := pos.EndLine + contextLines
	if endLine > len(sourceLines) {
		endLine = len(sourceLines)
	}

	maxLineNumWidth := len(fmt.Sprintf("%d", endLine))

	// Show context lines before
	for i := startLine; i < pos.Line; i++ {
		sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
	}

	// Handle single-line or multi-line spans
	if pos.Line == pos.EndLine {
		// Single line error
		line := sourceLines[pos.Line-1]
		sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, pos.Line, line))
		sb.WriteString(fmt.Sprintf("  %s | ", strings.Repeat(" ", maxLineNumWidth)))

		// Add spacing and pointer
		if pos.Column > 0 {
			sb.WriteString(strings.Repeat(" ", pos.Column))
		}
		pointerLength := pos.EndColumn - pos.Column
		if pointerLength < 1 {
			pointerLength = 1
		}
		sb.WriteString("\033[1;31m")
		sb.WriteString(strings.Repeat("^", pointerLength))
		sb.WriteString("\033[0m\n")
	} else {
		// Multi-line error - highlight all lines
		for lineNum := pos.Line; lineNum <= pos.EndLine && lineNum <= len(sourceLines); lineNum++ {
			line := sourceLines[lineNum-1]
			sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, lineNum, line))

			// Add pointer for each line
			sb.WriteString(fmt.Sprintf("  %s | ", strings.Repeat(" ", maxLineNumWidth)))

			if lineNum == pos.Line {
				// First line: highlight from start column to end
				if pos.Column > 0 {
					sb.WriteString(strings.Repeat(" ", pos.Column))
				}
				lineLength := len(line) - pos.Column
				sb.WriteString("\033[1;31m")
				sb.WriteString(strings.Repeat("^", lineLength))
				sb.WriteString("\033[0m\n")
			} else if lineNum == pos.EndLine {
				// Last line: highlight from beginning to end column
				sb.WriteString("\033[1;31m")
				sb.WriteString(strings.Repeat("^", pos.EndColumn))
				sb.WriteString("\033[0m\n")
			} else {
				// Middle lines: highlight entire line
				sb.WriteString("\033[1;31m")
				sb.WriteString(strings.Repeat("^", len(line)))
				sb.WriteString("\033[0m\n")
			}
		}
	}

	// Show context lines after
	for i := pos.EndLine + 1; i <= endLine; i++ {
		sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
	}

	return sb.String()
}

// ParserErrorEOF creates an error message for unexpected end of input
// Shows the last token to give context about where more input was expected
func ParserErrorEOF(message string, lastToken lexer.Token, sourceLines []string) string {
	var sb strings.Builder

	// Error header
	sb.WriteString(fmt.Sprintf("\n\033[1;31mParser Error:\033[0m %s\n", message))
	sb.WriteString(fmt.Sprintf("\033[36mAfter line %d, column %d\033[0m\n\n", lastToken.Line, lastToken.EndColumn))

	// Show the last token line if available
	if lastToken.Line > 0 && lastToken.Line <= len(sourceLines) {
		lineNum := lastToken.Line

		// Calculate the maximum line number width for alignment
		contextLines := 2 // lines before and after
		startLine := lineNum - contextLines
		if startLine < 1 {
			startLine = 1
		}
		endLine := lineNum + contextLines
		if endLine > len(sourceLines) {
			endLine = len(sourceLines)
		}

		maxLineNumWidth := len(fmt.Sprintf("%d", endLine))

		// Show context lines before
		for i := startLine; i < lineNum; i++ {
			sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
		}

		// Show the error line
		line := sourceLines[lineNum-1]
		sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, lineNum, line))

		// Show the pointer line
		sb.WriteString(fmt.Sprintf("  %s | ", strings.Repeat(" ", maxLineNumWidth)))

		// Add spacing to point after the last token
		if lastToken.EndColumn > 0 {
			sb.WriteString(strings.Repeat(" ", lastToken.EndColumn))
		}

		// Add the pointer after the last token
		sb.WriteString("\033[1;31m^ expected more input here\033[0m\n")

		// Show context lines after
		for i := lineNum + 1; i <= endLine; i++ {
			sb.WriteString(fmt.Sprintf("  %*d | %s\n", maxLineNumWidth, i, sourceLines[i-1]))
		}
	}

	return sb.String()
}
