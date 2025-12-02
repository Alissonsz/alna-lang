package disassembler

import (
	"alna-lang/internal/opcode"
	"fmt"
	"strings"
)

type Constant struct {
	TypeID int
	Value  interface{}
}

// Disassemble takes bytecode and returns a human-readable disassembly
func Disassemble(bytecode []byte) string {
	var output strings.Builder

	output.WriteString("=== BYTECODE DISASSEMBLY ===\n")

	if len(bytecode) < 8 {
		output.WriteString("Error: Bytecode too short (expected at least 8 bytes for header)\n")
		return output.String()
	}

	// Parse and validate header
	magic := bytecode[0:4]
	version := bytecode[4:8]

	if magic[0] != 0x7F || magic[1] != 'A' || magic[2] != 'L' || magic[3] != 'N' {
		output.WriteString(fmt.Sprintf("Error: Invalid magic number: %02X %02X %02X %02X\n", magic[0], magic[1], magic[2], magic[3]))
		return output.String()
	}

	output.WriteString(fmt.Sprintf("Header: Valid (ALNA v%d.%d.%d.%d)\n\n", version[0], version[1], version[2], version[3]))

	// Parse constants pool
	pos := 8
	if pos >= len(bytecode) {
		output.WriteString("Error: No constants pool found\n")
		return output.String()
	}

	constantCount := int(bytecode[pos])
	pos++

	output.WriteString(fmt.Sprintf("Constants Pool (%d entries):\n", constantCount))

	constants := make([]Constant, 0, constantCount)
	for i := 0; i < constantCount; i++ {
		if pos >= len(bytecode) {
			output.WriteString(fmt.Sprintf("Error: Unexpected end while reading constant %d\n", i))
			return output.String()
		}

		typeID := int(bytecode[pos])
		pos++

		if pos >= len(bytecode) {
			output.WriteString(fmt.Sprintf("Error: Unexpected end while reading constant %d value\n", i))
			return output.String()
		}

		value := int(bytecode[pos])
		pos++

		constants = append(constants, Constant{TypeID: typeID, Value: value})

		typeName := "unknown"
		if typeID == 1 {
			typeName = "int"
		}

		output.WriteString(fmt.Sprintf("  [%d] %s: %v\n", i, typeName, value))
	}

	output.WriteString("\nInstructions:\n")

	// Remember where instructions start for proper IP addressing
	instructionsStart := pos

	// Parse instructions
	for pos < len(bytecode) {
		// Show instruction position relative to start of instructions (matches IP)
		instructionPos := pos - instructionsStart
		op := opcode.Opcode(bytecode[pos])
		pos++

		instruction := fmt.Sprintf("  %04d: %s", instructionPos, op.String())

		if op.HasOperand() {
			if pos >= len(bytecode) {
				output.WriteString(fmt.Sprintf("%s <missing operand>\n", instruction))
				break
			}

			operand := int(bytecode[pos])
			pos++

			instruction += fmt.Sprintf(" %d", operand)

			// Add comment for constant references
			if op == opcode.LOAD_CONST && operand < len(constants) {
				instruction += fmt.Sprintf("    ; load %v", constants[operand].Value)
			}
		}

		output.WriteString(instruction + "\n")
	}

	return output.String()
}
