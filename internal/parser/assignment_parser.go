package parser

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/lexer"
)

func (p *Parser) parseAssignment() (ast.Node, error) {
	identifier, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	assignment := p.currentToken()
	if assignment.Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	if assignment.Type != lexer.Assignment {
		return nil, p.expectedGotError(assignment, "=")
	}

	p.advance()
	if p.currentToken().Type == lexer.EOF {
		return nil, p.unexpectedEOFError()
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return ast.AssignmentNode{
		Left:  identifier,
		Right: value,
		Position: common.Position{
			Line:      identifier.Pos().Line,
			Column:    identifier.Pos().Column,
			EndLine:   value.Pos().EndLine,
			EndColumn: value.Pos().EndColumn,
		},
	}, nil

}
