package parser

import (
	"alna-lang/internal/logger"
	"alna-lang/internal/lexer"
)

type Parser struct {
	tokens      []lexer.Token
	pos         int
	sourceLines []string
	logger      *logger.Logger
}
