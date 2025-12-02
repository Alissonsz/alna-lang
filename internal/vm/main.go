package vm

import (
	"alna-lang/internal/builtins"
	"alna-lang/internal/codegen"
	"alna-lang/internal/logger"
	"alna-lang/internal/opcode"
	"fmt"
)

type VM struct {
	program   []byte
	rawCode   []string
	Pc        int
	PcOffset  int
	stack     []any
	constants []any
	Variables []any
	Functions []FunctionDefinition
	debugMode bool
	logger    *logger.Logger
}

type FunctionType int

const (
	FunctionTypeBuiltin FunctionType = iota
	FunctionTypeCompiled
)

type FunctionDefinition struct {
	Name           string
	Implementation builtins.Function
	Type           FunctionType
}

func NewVM(program []byte, code []string, debugMode bool, lgr *logger.Logger) *VM {
	return &VM{
		program:   program,
		rawCode:   code,
		Pc:        0,
		stack:     []any{},
		debugMode: debugMode,
		logger:    lgr,
	}
}

func (vm *VM) CheckHeader() error {
	header := vm.readBytes(4)
	expectedHeader := []byte{0x7F, 'A', 'L', 'N'}
	for i := 0; i < 4; i++ {
		if header[i] != expectedHeader[i] {
			return fmt.Errorf("invalid magic number")
		}
	}

	version := vm.readBytes(4)
	vm.logger.Debug("Version: %v", version)
	return nil
}

func (vm *VM) Run() error {
	costantsCount := vm.readByte()
	vm.constants = make([]any, int(costantsCount))
	vm.registerBuiltins()

	for i := 0; i < int(costantsCount); i++ {
		typeId := vm.readByte()
		switch typeId {
		case codegen.IntTypeId:
			intValue := vm.readByte()

			vm.logger.Debug("Constant %d: INT %d", i, intValue)
			vm.constants[i] = int(intValue)
		}
	}
	vm.PcOffset = vm.Pc
	if vm.debugMode {
		vm.logger.Info("=== STARTING DEBUG MODE ===")
		err := vm.StartTuiDebugger()
		if err != nil {
			return err
		}
		return nil
	}

	for i := vm.Pc; i < len(vm.program)-1; i++ {
		err := vm.Step()
		if err != nil {
			return err
		}
	}

	return nil
}

func (vm *VM) Step() error {
	if vm.Pc >= len(vm.program) {
		return nil
	}
	op := vm.readByte()

	switch op {
	case byte(opcode.LOAD_CONST):
		constIndex := vm.readByte()
		constValue := vm.constants[int(constIndex)]

		vm.pushStack(constValue)
		vm.logger.Debug("LOAD_CONST %d -> %v", constIndex, constValue)
	case byte(opcode.LOAD_VAR):
		varIndex := vm.readByte()
		varValue := vm.getVariable(int(varIndex))
		vm.pushStack(varValue)
		vm.logger.Debug("LOAD_VAR %d -> %v", varIndex, varValue)
	case byte(opcode.STORE_VAR):
		varIndex := vm.readByte()
		value := vm.popStack()
		vm.pushVariable(value)

		vm.logger.Debug("STORE_VAR %d <- %v", varIndex, value)

	case byte(opcode.ADD):
		right := vm.popStack()
		left := vm.popStack()
		result := left.(int) + right.(int)
		vm.pushStack(result)
		vm.logger.Debug("ADD %v + %v -> %v", left, right, result)

	case byte(opcode.GT):
		right := vm.popStack()
		left := vm.popStack()
		result := left.(int) > right.(int)
		vm.pushStack(result)
		vm.logger.Debug("GT %v > %v -> %v", left, right, result)

	case byte(opcode.JUMP_IF_FALSE):
		target := vm.readByte()
		condition := vm.popStack()
		if condition == false {
			vm.Pc = int(target) + vm.PcOffset
			vm.logger.Debug("JUMP_IF_FALSE to %d", target)
		} else {
			vm.logger.Debug("JUMP_IF_FALSE skipped")
		}
	case byte(opcode.START_SCOPE):
		localsIndex := vm.readByte()
		vm.pushStack(int(localsIndex))
		vm.logger.Debug("START_SCOPE %d", localsIndex)
	case byte(opcode.END_SCOPE):
		vm.logger.Debug("END_SCOPE")
		scopeVarIndex := vm.popStack().(int)
		vm.logger.Debug("Clearing variables till index %d", scopeVarIndex)
		vm.clearVariablesTill(scopeVarIndex)
	case byte(opcode.CALL):
		funcIndex := vm.readByte()
		function := vm.Functions[int(funcIndex)]
		vm.logger.Debug("CALL function %s", function.Name)

		switch function.Type {
		case FunctionTypeBuiltin:
			arg := vm.popStack()
			result := function.Implementation(arg)

			if result != nil {
				vm.pushStack(result)
				vm.logger.Debug("Function %s returned %v", function.Name, result)

			}
		case FunctionTypeCompiled:
			// For compiled functions, we would need to implement a call stack and manage execution context
			return fmt.Errorf("compiled function calls not implemented yet")
		}

	default:
		return fmt.Errorf("unknown opcode: 0x%02X at pc %d", op, vm.Pc-1)
	}
	return nil

}

func (vm *VM) readByte() byte {
	if vm.Pc >= len(vm.program) {
		return 0
	}

	b := vm.program[vm.Pc]
	vm.Pc++
	return b
}

func (vm *VM) readBytes(n int) []byte {
	bytes := vm.program[vm.Pc : vm.Pc+n]
	vm.Pc += n
	return bytes
}

func (vm *VM) pushStack(value any) {
	vm.stack = append(vm.stack, value)
}

func (vm *VM) popStack() any {
	if len(vm.stack) == 0 {
		return nil
	}
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return value
}

func (vm *VM) pushVariable(value any) int {
	vm.Variables = append(vm.Variables, value)
	return len(vm.Variables) - 1
}

func (vm *VM) getVariable(index int) any {
	if index < 0 || index >= len(vm.Variables) {
		return nil
	}
	return vm.Variables[index]
}

func (vm *VM) clearVariablesTill(index int) {
	if index < 0 || index >= len(vm.Variables) {
		return
	}
	vm.Variables = vm.Variables[:index+1]
}
