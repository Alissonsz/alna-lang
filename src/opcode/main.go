package opcode

import "fmt"

type Opcode byte

const (
	LOAD_CONST    Opcode = 0x00
	STORE_VAR     Opcode = 0x01
	LOAD_VAR      Opcode = 0x02
	ADD           Opcode = 0x03
	SUB           Opcode = 0x04
	MUL           Opcode = 0x05
	DIV           Opcode = 0x06
	EQ            Opcode = 0x07
	LT            Opcode = 0x08
	GT            Opcode = 0x09
	JUMP_IF_FALSE Opcode = 0x0A
	JUMP_IF_TRUE  Opcode = 0x0B
)

// String returns the mnemonic name of the opcode
func (op Opcode) String() string {
	switch op {
	case LOAD_CONST:
		return "LOAD_CONST"
	case STORE_VAR:
		return "STORE_VAR"
	case LOAD_VAR:
		return "LOAD_VAR"
	case ADD:
		return "ADD"
	case SUB:
		return "SUB"
	case MUL:
		return "MUL"
	case DIV:
		return "DIV"
	case EQ:
		return "EQ"
	case LT:
		return "LT"
	case GT:
		return "GT"
	case JUMP_IF_FALSE:
		return "JUMP_IF_FALSE"
	case JUMP_IF_TRUE:
		return "JUMP_IF_TRUE"
	default:
		fmt.Printf("Unknown opcode: %d\n", op)
		return "UNKNOWN"
	}
}

// HasOperand returns true if the opcode takes a 1-byte operand
func (op Opcode) HasOperand() bool {
	switch op {
	case LOAD_CONST, STORE_VAR, LOAD_VAR, JUMP_IF_FALSE, JUMP_IF_TRUE:
		return true
	default:
		return false
	}
}
