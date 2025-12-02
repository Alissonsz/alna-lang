package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"fmt"
)

func (p *Parser) parseAssignment() ast.Node {
	token := p.tokens[p.pos]
	left := p.parseIdentifier()

	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected '='", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	assignToken := p.tokens[p.pos]
	if assignToken.Type != lexer.Assignment {
		panic(common.CompilerError(tokenToPosition(assignToken), fmt.Sprintf("Expected '=', got %v", assignToken.Type), p.sourceLines))
	}

	p.pos++
	if p.pos >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected expression after '='", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	right := p.parseExpression()
	return ast.AssignmentNode{
		Left:     left,
		Right:    right,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   right.Pos().EndLine,
			EndColumn: right.Pos().EndColumn,
		},
	}

}
