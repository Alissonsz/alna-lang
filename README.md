# Alna Language

A simple programming language with its own lexer, parser, analyzer, code generator, and virtual machine.

## Build

```bash
go build -o alna-lang main.go
```

or

```bash
go build .
```

## Run

The compiler and runner are unified into a single command. It compiles the source code and immediately executes it:

```bash
./alna-lang <source.alna>
```

## Flags

| Flag | Description |
|------|-------------|
| `-verbose` | Print tokens, AST, symbol table, and bytecode during compilation |
| `-disassemble` | Show human-readable bytecode disassembly |
| `-debug` | Run with interactive TUI debugger |

## Examples

```bash
# Basic run
./alna-lang examples/valid_program.alna

# Verbose output (shows compilation details)
./alna-lang -verbose examples/conditionals.alna

# Disassemble bytecode
./alna-lang -disassemble examples/basic_arith.alna

# Debug with TUI
./alna-lang -debug examples/valid_program.alna
```

## TUI Debugger Controls

| Key | Action |
|-----|--------|
| `n` | Step to next instruction |
| `q` | Quit debugger |

## Output

Compiled bytecode is saved to `out.alnbc` after each run.

## Development

```bash
# Run without building
go run main.go <source.alna>

# Run with flags
go run main.go -verbose examples/valid_program.alna
```

## Testing

```bash
# Run the lexer snapshot tests
go test ./internal/lexer
```

```bash
# Run and update the snapshot tests
go test ./internal/lexer -update
```