# Alna Programming Language - AI Agent Context

## Project Overview

Alna is a custom programming language implementation written in Go. It includes a complete toolchain:
- **Lexer** - Tokenizes source code
- **Parser** - Builds an Abstract Syntax Tree (AST)
- **Analyzer** - Semantic analysis and symbol table management
- **Code Generator** - Produces bytecode
- **Virtual Machine** - Executes bytecode

File extension: `.alna`
Bytecode extension: `.alnbc`

## Build & Run Commands

```bash
# Build
go build -o alna-lang main.go

# Run
./alna-lang <source.alna>

# Run with verbose output (tokens, AST, symbol table, bytecode)
./alna-lang -verbose <source.alna>

# Disassemble bytecode
./alna-lang -disassemble <source.alna>

# Debug with TUI
./alna-lang -debug <source.alna>

# Development
go run main.go <source.alna>
```

## Project Structure

```
alna-lang/
├── main.go                 # Entry point, orchestrates compilation pipeline
├── internal/
│   ├── lexer/              # Lexical analysis (tokenization)
│   ├── parser/             # Parser and AST construction
│   ├── ast/                # AST node definitions and printer
│   ├── analyzer/           # Semantic analysis, symbol table
│   ├── codegen/            # Bytecode generation
│   ├── vm/                 # Virtual machine execution
│   ├── opcode/             # Bytecode opcodes definition
│   ├── disassembler/       # Bytecode to human-readable
│   ├── symbol_table/       # Symbol table implementation
│   ├── logger/             # Logging utilities
│   ├── builtins/           # Built-in functions
│   └── common/             # Shared types (Position, Error)
├── tree-sitter-alna/       # Tree-sitter grammar for editor support
├── .vscode-extension/      # VS Code syntax highlighting
├── examples/               # Example .alna files
└── docs/                   # Grammar documentation
```

## Compilation Pipeline

```
Source Code (.alna)
    ↓
Lexer (internal/lexer)
    ↓
Parser (internal/parser) → AST (internal/ast)
    ↓
Analyzer (internal/analyzer) → Symbol Table
    ↓
Code Generator (internal/codegen) → Bytecode (.alnbc)
    ↓
VM (internal/vm) → Execution
```

## Language Features

### Data Types
- Integers: `int`, `i8`, `i16`, `i32`, `i64`, `uint`, `u8`, `u16`, `u32`, `u64`
- Floats: `float`, `f32`, `f64`
- Other: `bool`, `string`
- Collections: `array<Type>`, `map<KeyableType, ValueType>`

### Syntax Examples

```alna
# Variable declaration
Type verboseVar = value
variable := value

# Conditional (always returns a value)
result = if condition {
  valueIfTrue
} else {
  valueIfFalse
}

# For loops
for item in collection { }
for i := 0; i < 10; i = i + 1 { }

# Match expressions
match expression {
  when value1 { }
  when value2 { }
  default { }
}

# Functions
ReturnType name(arg1: Type, arg2: Type) {
  return value
}
```

## AST Node Types

Located in `internal/ast/ast.go`:
- `RootNode` - Program root
- `BooleanNode`, `NumberNode`, `IdentifierNode` - Literals
- `BinaryOpNode` - Binary operations
- `AssignmentNode` - Variable assignment
- `VariableDeclarationNode` - Variable declaration
- `BlockNode` - Block of expressions
- `IfExpressionNode` - If-else expression
- `FunctionDeclarationNode`, `FunctionCallNode` - Functions
- `ReturnNode` - Return statement

## Bytecode Opcodes

Located in `internal/opcode/main.go`:
- `LOAD_CONST`, `STORE_VAR`, `LOAD_VAR` - Data movement
- `ADD`, `SUB`, `MUL`, `DIV` - Arithmetic
- `EQ`, `LT`, `GT` - Comparison
- `JUMP`, `JUMP_IF_FALSE`, `JUMP_IF_TRUE` - Control flow
- `START_SCOPE`, `END_SCOPE` - Scope management
- `CALL`, `CALL_BUILTIN`, `RETURN` - Function handling

## Testing

```bash
# Run parser tests
go test ./internal/parser/...

# Run lexer tests
go test ./internal/lexer/...
```

## Key Files for Common Tasks

| Task | Files to modify |
|------|-----------------|
| Add new AST node | `internal/ast/ast.go`, `internal/ast/printer.go` |
| Add new opcode | `internal/opcode/main.go`, `internal/codegen/main.go`, `internal/vm/main.go` |
| Modify grammar | `internal/parser/parser.go`, `internal/parser/*.go` |
| Add built-in function | `internal/builtins/main.go`, `internal/vm/builtins.go` |
| Change lexer tokens | `internal/lexer/main.go` |

## Editor Support

- **VS Code**: `.vscode-extension/` - syntax highlighting
- **Neovim**: `tree-sitter-alna/` - tree-sitter grammar
