package vm

import (
	"alna-lang/internal/builtins"
	"alna-lang/internal/codegen"
	"alna-lang/internal/logger"
	"alna-lang/internal/opcode"
	"encoding/json"
	"fmt"
	"os"
)

type DebugInfo struct {
	Version     int              `json:"version"`
	SourceFile  string           `json:"sourceFile"`
	SourceLines []string         `json:"sourceLines"`
	Variables   []VariableInfo   `json:"variables"`
	Functions   []FunctionInfo   `json:"functions"`
	SourceMap   []SourceMapEntry `json:"sourceMap"`
}

type VariableInfo struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

type FunctionInfo struct {
	ConstantIndex int    `json:"constantIndex"`
	Name          string `json:"name"`
}

type SourceMapEntry struct {
	Pc        int    `json:"pc"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndColumn int    `json:"endColumn"`
	VarName   string `json:"varName,omitempty"`
}

type SourcePosition struct {
	Line      int
	Column    int
	EndColumn int
	VarName   string
}

type VM struct {
	program       []byte
	rawCode       []string
	Pc            int
	PcOffset      int
	stack         []any
	constants     []any
	Variables     []any
	Functions     []FunctionDefinition
	debugMode     bool
	logger        *logger.Logger
	debugInfo     *DebugInfo
	VariableNames map[int]string
	SourceMap     map[int]SourcePosition
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
	Instructions   []byte
}

func NewVM(program []byte, code []string, debugMode bool, lgr *logger.Logger) *VM {
	vm := &VM{
		program:       program,
		rawCode:       code,
		Pc:            0,
		stack:         []any{},
		debugMode:     debugMode,
		logger:        lgr,
		VariableNames: make(map[int]string),
		SourceMap:     make(map[int]SourcePosition),
	}
	return vm
}

func (vm *VM) LoadDebugFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read debug file: %w", err)
	}

	var debugInfo DebugInfo
	if err := json.Unmarshal(data, &debugInfo); err != nil {
		return fmt.Errorf("failed to parse debug file: %w", err)
	}

	vm.debugInfo = &debugInfo
	vm.rawCode = debugInfo.SourceLines

	for _, v := range debugInfo.Variables {
		vm.VariableNames[v.Index] = v.Name
	}

	for _, entry := range debugInfo.SourceMap {
		vm.SourceMap[entry.Pc] = SourcePosition{
			Line:      entry.Line,
			Column:    entry.Column,
			EndColumn: entry.EndColumn,
			VarName:   entry.VarName,
		}
	}

	return nil
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

	compiledFuncNames := make(map[int]string)
	if vm.debugInfo != nil {
		for _, fn := range vm.debugInfo.Functions {
			compiledFuncNames[fn.ConstantIndex] = fn.Name
		}
	}

	for i := 0; i < int(costantsCount); i++ {
		typeId := vm.readByte()
		switch typeId {
		case codegen.IntTypeId:
			intValue := vm.readByte()

			vm.logger.Debug("Constant %d: INT %d", i, intValue)
			vm.constants[i] = int(intValue)

		case codegen.FunctionTypeId:
			instructionsCount := vm.readByte()
			functionInstructions := vm.readBytes(int(instructionsCount))

			funcName := fmt.Sprintf("func_%d", i)
			if name, ok := compiledFuncNames[i]; ok {
				funcName = name
			}

			function := FunctionDefinition{
				Name:           funcName,
				Implementation: nil,
				Type:           FunctionTypeCompiled,
				Instructions:   functionInstructions,
			}

			vm.Functions = append(vm.Functions, function)
			vm.logger.Debug("Constant %d: COMPILED FUNCTION with %d bytes", i, instructionsCount)
		default:
			return fmt.Errorf("unknown constant type id: %d", typeId)
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
