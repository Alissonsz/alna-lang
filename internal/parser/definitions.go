package parser

import (
	"alna-lang/internal/lexer"
	"alna-lang/internal/logger"
)

type Parser struct {
	tokens      []lexer.Token
	position    int
	sourceLines []string
	logger      *logger.Logger
}
