package codegen

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/builtins"
	"alna-lang/internal/logger"
	"alna-lang/internal/opcode"
	symboltable "alna-lang/internal/symbol_table"
	"strconv"
)

const IntTypeId = 1

type ConstantDefinition struct {
	Value  any
	TypeId int
}

type VariableDefinition struct {
	Name  string
	Index int
}

type CodeGenerator struct {
	ast          ast.RootNode
	sourceLines  []string
	symbolTable  *symboltable.SymbolTable
	constants    []ConstantDefinition
	constantMap  map[interface{}]int
	variables    []interface{}
	variablesMap map[string]int
	functions    []builtins.Function
	functionsMap map[string]int
	mainBytecode []byte
	Bytecode     []byte
	logger       *logger.Logger
}

func NewCodeGenerator(tree ast.RootNode, srcLines []string, st *symboltable.SymbolTable, lgr *logger.Logger) *CodeGenerator {
	return &CodeGenerator{ast: tree, sourceLines: srcLines, symbolTable: st, constantMap: make(map[interface{}]int), logger: lgr}
}

func (cg *CodeGenerator) AddConstant(typeId int, value interface{}) int {
	if idx, exists := cg.constantMap[value]; exists {
		return idx
	}

	cg.constants = append(cg.constants, ConstantDefinition{Value: value, TypeId: typeId})
	cg.constantMap[value] = len(cg.constants) - 1
	return len(cg.constants) - 1
}

func (cg *CodeGenerator) AddVariable(name string) int {
	if idx, exists := cg.variablesMap[name]; exists {
		return idx
	}
	if cg.variablesMap == nil {
		cg.variablesMap = make(map[string]int)
	}
	cg.variablesMap[name] = len(cg.variables)
	idx := len(cg.variables)
	cg.variables = append(cg.variables, nil)
	return idx
}

func (cg *CodeGenerator) Generate() string {
	// write magic number and version
	cg.Bytecode = append(cg.Bytecode, 0x7F, 'A', 'L', 'N')    // Magic number
	cg.Bytecode = append(cg.Bytecode, 0x01, 0x00, 0x00, 0x00) // Version

	builtins := builtins.GetBuiltins()
	cg.functionsMap = make(map[string]int)
	for name, fn := range builtins {
		cg.functions = append(cg.functions, fn)
		cg.functionsMap[name] = len(cg.functions) - 1
	}

	st := cg.ast.SymbolTable
	cg.logger.Debug("Starting code generation")
	cg.logger.Debug("Symbol Table at root:")
	cg.logger.Debug("%+v", st)
	for _, stmt := range cg.ast.Children {
		cg.generateStatement(stmt, st)
	}

	cg.writeConstantsPool()
	cg.Bytecode = append(cg.Bytecode, cg.mainBytecode...)

	return ""
}

func (cg *CodeGenerator) writeConstantsPool() {
	// write constants pool
	cg.Bytecode = append(cg.Bytecode, byte(len(cg.constants)))
	for _, constant := range cg.constants {
		cg.logger.Debug("Writing constant to bytecode: %#v", constant)
		cg.Bytecode = append(cg.Bytecode, byte(constant.TypeId))
		switch constant.TypeId {
		case IntTypeId:
			cg.Bytecode = append(cg.Bytecode, byte(constant.Value.(int8)))
		}
	}
}

func (cg *CodeGenerator) generateStatement(stmt ast.Node, st *symboltable.SymbolTable) string {
	switch node := stmt.(type) {
	case ast.VariableDeclarationNode:
		return cg.generateVariableDeclaration(node, st)
	case ast.AssignmentNode:
		cg.generateExpression(node.Right, st)
		var varName string
		switch node.Left.(type) {
		case ast.IdentifierNode:
			varName = node.Left.(ast.IdentifierNode).Name

			cg.emit(opcode.STORE_VAR, cg.variablesMap[varName])
		default:
			cg.logger.Error("Invalid assignment target at position %+v", node.Left.Pos())
		}
	case ast.BinaryOpNode:
		cg.generateExpression(node, st)
	case ast.IfStatementNode:
		cg.generateExpression(node.Condition, st)
		cg.emit(opcode.JUMP_IF_FALSE, 0)
		thenStart := len(cg.mainBytecode)
		cg.generateStatement(node.ThenBranch, st)

		if node.ElseBranch != nil {
			elseStart := len(cg.mainBytecode)
			cg.mainBytecode[thenStart-1] = byte(elseStart)
			cg.generateStatement(node.ElseBranch, st)
		} else {
			cg.mainBytecode[thenStart-1] = byte(len(cg.mainBytecode))
		}
	case *ast.BlockNode:
		if node == nil {
			return ""
		}

		scopeVarIndex := len(cg.variables) - 1
		cg.emit(opcode.START_SCOPE, scopeVarIndex)
		for _, statement := range node.Statements {
			cg.generateStatement(statement, node.SymbolTable)
		}
		cg.emit(opcode.END_SCOPE)
		cg.variables = cg.variables[:scopeVarIndex+1]

	case ast.BlockNode:
		cg.logger.Debug("Entering new block scope in codegen")
		cg.emit(opcode.START_SCOPE, len(cg.variables)-1)
		for _, statement := range node.Statements {
			cg.generateStatement(statement, node.SymbolTable)
		}
		cg.emit(opcode.END_SCOPE)
	case ast.FunctionCallNode:
		for _, arg := range node.Arguments {
			cg.generateExpression(arg, st)
		}
		if fnIdx, exists := cg.functionsMap[node.Name]; exists {
			cg.emit(opcode.CALL, fnIdx)
		} else {
			cg.logger.Error("Undefined function '%s' at position %+v", node.Name, node.Pos())
		}
	default:
		cg.logger.Warn("Unknown statement type: %T at position %+v", node, node.Pos())
	}

	return ""
}

func (cg *CodeGenerator) generateVariableDeclaration(node ast.VariableDeclarationNode, st *symboltable.SymbolTable) string {
	varIdx := cg.AddVariable(node.Name)

	if node.Initializer != nil {
		cg.generateExpression(node.Initializer, st)
	}

	cg.emit(opcode.STORE_VAR, varIdx)
	return ""
}

func (cg *CodeGenerator) generateExpression(expr ast.Node, st *symboltable.SymbolTable) string {
	switch node := expr.(type) {
	case ast.NumberNode:
		intValue, err := strconv.ParseInt(node.Value.(string), 10, 8)
		if err != nil {
			cg.logger.Error("Error parsing number '%s' at position %+v: %v", node.Value, node.Pos(), err)
			return ""
		}

		constIdx := cg.AddConstant(IntTypeId, int8(intValue))
		cg.logger.Debug("Generating LOAD_CONST for number %d at index %d", intValue, constIdx)

		cg.emit(opcode.LOAD_CONST, constIdx)
	case ast.IdentifierNode:
		if varIdx, exists := cg.variablesMap[node.Name]; exists {
			cg.emit(opcode.LOAD_VAR, varIdx)
		} else {
			cg.logger.Error("Undefined variable '%s' at position %+v", node.Name, node.Pos())
		}
	case ast.BinaryOpNode:
		cg.generateExpression(node.Left, st)
		cg.generateExpression(node.Right, st)
		op := node.Operator.Value
		switch op {
		case "*":
			cg.emit(opcode.MUL)
		case "/":
			cg.emit(opcode.DIV)
		case "+":
			cg.emit(opcode.ADD)
		case "-":
			cg.emit(opcode.SUB)
		case "==":
			cg.emit(opcode.EQ)
		case "<":
			cg.emit(opcode.LT)
		case ">":
			cg.emit(opcode.GT)
		default:
			cg.logger.Error("Unknown binary operator '%s' at position %+v", op, node.Pos())
		}

	default:
		cg.logger.Warn("Unknown expression type: %T at position %+v", node, node.Pos())
	}
	return ""
}

func (cg *CodeGenerator) emit(op opcode.Opcode, operands ...int) {
	cg.logger.Debug("Emitting opcode: %d with operands %v", op, operands)

	switch op {
	case opcode.LOAD_CONST, opcode.STORE_VAR, opcode.LOAD_VAR:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.ADD, opcode.SUB, opcode.MUL, opcode.DIV, opcode.EQ, opcode.LT, opcode.GT:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	case opcode.JUMP_IF_FALSE, opcode.JUMP_IF_TRUE, opcode.CALL:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.START_SCOPE:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.END_SCOPE:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	default:
		cg.logger.Error("Unknown opcode to emit: %d", op)
	}
}
