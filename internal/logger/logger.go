package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Logger provides structured logging with different levels
type Logger struct {
	level      LogLevel
	debugLog   *log.Logger
	infoLog    *log.Logger
	warnLog    *log.Logger
	errorLog   *log.Logger
	verboseOut io.Writer
}

// New creates a new Logger with the specified level
func New(level LogLevel, verbose bool) *Logger {
	flags := log.Ltime

	var verboseOut io.Writer = io.Discard
	if verbose {
		verboseOut = os.Stdout
	}

	return &Logger{
		level:      level,
		debugLog:   log.New(verboseOut, "[DEBUG] ", flags),
		infoLog:    log.New(verboseOut, "[INFO]  ", flags),
		warnLog:    log.New(os.Stderr, "[WARN]  ", flags),
		errorLog:   log.New(os.Stderr, "[ERROR] ", flags),
		verboseOut: verboseOut,
	}
}

// NewDefault creates a logger with Info level and no verbose output
func NewDefault() *Logger {
	return New(LevelInfo, false)
}

// Debug logs a debug message (only when verbose is enabled)
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.debugLog.Printf(format, args...)
	}
}

// Info logs an info message (only when verbose is enabled)
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.infoLog.Printf(format, args...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.warnLog.Printf(format, args...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.errorLog.Printf(format, args...)
	}
}

// Print writes directly to the verbose output (for user-facing output)
// This is for output that should always be shown, like AST dumps or bytecode
func (l *Logger) Print(format string, args ...interface{}) {
	if l.verboseOut != io.Discard {
		fmt.Fprintf(l.verboseOut, format, args...)
	}
}

// Println writes directly to the verbose output with a newline
func (l *Logger) Println(args ...interface{}) {
	if l.verboseOut != io.Discard {
		fmt.Fprintln(l.verboseOut, args...)
	}
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// IsVerbose returns whether verbose output is enabled
func (l *Logger) IsVerbose() bool {
	return l.verboseOut != io.Discard
}
