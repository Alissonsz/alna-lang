package opcode

import "fmt"

type Opcode byte

const (
	LOAD_CONST Opcode = iota
	STORE_VAR
	LOAD_VAR
	ADD
	SUB
	MUL
	DIV
	EQ
	LT
	GT
	JUMP_IF_FALSE
	JUMP_IF_TRUE
	START_SCOPE
	END_SCOPE
	CALL
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
	case START_SCOPE:
		return "START_SCOPE"
	case END_SCOPE:
		return "END_SCOPE"
	case CALL:
		return "CALL"
	default:
		fmt.Printf("Unknown opcode: %d\n", op)
		return "UNKNOWN"
	}
}

// HasOperand returns true if the opcode takes a 1-byte operand
func (op Opcode) HasOperand() bool {
	switch op {
	case LOAD_CONST, LOAD_VAR, STORE_VAR, JUMP_IF_FALSE, JUMP_IF_TRUE, START_SCOPE, CALL:
		return true
	default:
		return false
	}
}
