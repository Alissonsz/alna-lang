package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
	"fmt"
)

func (p *Parser) parseAssignment() ast.Node {
	p.logger.Debug("Parsing assignment expression")
	token := p.tokens[p.position]
	left := p.parseIdentifier()

	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected '='", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	assignToken := p.tokens[p.position]
	if assignToken.Type != lexer.Assignment {
		panic(common.CompilerError(tokenToPosition(assignToken), fmt.Sprintf("Expected '=', got %v", assignToken.Type), p.sourceLines))
	}

	p.position++
	if p.position >= len(p.tokens) {
		panic(common.CompilerErrorEOF("Unexpected end of input, expected expression after '='", tokenToPosition(p.lastToken()), p.sourceLines))
	}

	right := p.parseExpression()
	return ast.AssignmentNode{
		Left:  left,
		Right: right,
		Position: common.Position{
			Line:      token.Line,
			Column:    token.StartColumn,
			EndLine:   right.Pos().EndLine,
			EndColumn: right.Pos().EndColumn,
		},
	}

}
