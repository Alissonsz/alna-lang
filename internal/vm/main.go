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
	callStack     []int
	scopeStack    []int
	basePointer   int
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
		callStack:     []int{},
		scopeStack:    []int{},
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
	startingPc := vm.readBytes(4)
	mainAddress := int(startingPc[0]) | int(startingPc[1])<<8 | int(startingPc[2])<<16 | int(startingPc[3])<<24
	vm.logger.Debug("Starting PC: %d", mainAddress)
	costantsCount := vm.readByte()
	vm.logger.Debug("Constants count: %d", costantsCount)
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
		default:
			return fmt.Errorf("unknown constant type id: %d", typeId)
		}
	}
	vm.PcOffset = vm.Pc
	vm.Pc = mainAddress + vm.PcOffset
	vm.logger.Debug("Initial PC set to: %d", vm.Pc)

	if vm.debugMode {
		vm.logger.Info("=== STARTING DEBUG MODE ===")
		err := vm.StartTuiDebugger()
		if err != nil {
			return err
		}
		return nil
	}

	for vm.Pc < len(vm.program) {
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
		varValue := vm.getVariable(vm.basePointer + int(varIndex))
		vm.pushStack(varValue)
		vm.logger.Debug("LOAD_VAR %d (abs %d) -> %v", varIndex, vm.basePointer+int(varIndex), varValue)
	case byte(opcode.STORE_VAR):
		varIndex := vm.readByte()
		value := vm.popStack()
		absIndex := vm.basePointer + int(varIndex)
		for len(vm.Variables) <= absIndex {
			vm.Variables = append(vm.Variables, nil)
		}
		vm.Variables[absIndex] = value

		vm.logger.Debug("STORE_VAR %d (abs %d) <- %v", varIndex, absIndex, value)

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
		absIndex := vm.basePointer + int(localsIndex)
		vm.pushScopeStack(absIndex)
		vm.logger.Debug("START_SCOPE %d (abs %d)", localsIndex, absIndex)
	case byte(opcode.END_SCOPE):
		scopeVarIndex := vm.popScopeStack()
		vm.logger.Debug("END_SCOPE, clearing variables to abs index %d", scopeVarIndex)
		vm.Variables = vm.Variables[:scopeVarIndex]
	case byte(opcode.CALL_BUILTIN):
		funcIndex := vm.readByte()
		function := vm.Functions[int(funcIndex)]
		vm.logger.Debug("CALL_BUILTIN function %s", function.Name)
		arg := vm.popStack()
		result := function.Implementation(arg)
		if result != nil {
			vm.pushStack(result)
			vm.logger.Debug("Function %s returned %v", function.Name, result)
		}
	case byte(opcode.CALL):
		funcIndex := int(vm.readByte()) + vm.PcOffset
		vm.logger.Debug("CALL function at: %d", funcIndex)

		returnAddress := vm.Pc
		vm.pushCallStack(returnAddress)
		vm.pushCallStack(vm.basePointer)
		vm.basePointer = len(vm.Variables)
		vm.Pc = funcIndex

	case byte(opcode.RETURN):
		if len(vm.callStack) == 0 {
			vm.Pc = len(vm.program)
			vm.logger.Debug("RETURN from main - program ended")
			return nil
		}
		vm.Variables = vm.Variables[:vm.basePointer]
		vm.basePointer = vm.popCallStack()
		returnAddress := vm.popCallStack()
		vm.Pc = returnAddress
		vm.logger.Debug("RETURN to %d, basePointer restored to %d", returnAddress, vm.basePointer)

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

func (vm *VM) pushCallStack(returnAddress int) {
	vm.callStack = append(vm.callStack, returnAddress)
}

func (vm *VM) popCallStack() int {
	if len(vm.callStack) == 0 {
		return 0
	}
	address := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]
	return address
}

func (vm *VM) pushScopeStack(localsIndex int) {
	vm.scopeStack = append(vm.scopeStack, localsIndex)
}

func (vm *VM) popScopeStack() int {
	if len(vm.scopeStack) == 0 {
		return 0
	}
	index := vm.scopeStack[len(vm.scopeStack)-1]
	vm.scopeStack = vm.scopeStack[:len(vm.scopeStack)-1]
	return index
}

func (vm *VM) getVariable(index int) any {
	if index < 0 || index >= len(vm.Variables) {
		return nil
	}
	return vm.Variables[index]
}
