package codegen

import (
	"alna-lang/src/opcode"
	"alna-lang/src/parser"
	symboltable "alna-lang/src/symbol_table"
	"fmt"
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
	ast          parser.RootNode
	sourceLines  []string
	symbolTable  *symboltable.SymbolTable
	constants    []ConstantDefinition
	constantMap  map[interface{}]int
	variables    map[string]int
	mainBytecode []byte
	Bytecode     []byte
}

func NewCodeGenerator(ast parser.RootNode, srcLines []string, st *symboltable.SymbolTable) *CodeGenerator {
	return &CodeGenerator{ast: ast, sourceLines: srcLines, symbolTable: st, constantMap: make(map[interface{}]int)}
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
	if cg.variables == nil {
		cg.variables = make(map[string]int)
	}
	if idx, exists := cg.variables[name]; exists {
		return idx
	}
	idx := len(cg.variables)
	cg.variables[name] = idx
	return idx
}

func (cg *CodeGenerator) Generate() string {
	// write magic number and version
	cg.Bytecode = append(cg.Bytecode, 0x7F, 'A', 'L', 'N')    // Magic number
	cg.Bytecode = append(cg.Bytecode, 0x01, 0x00, 0x00, 0x00) // Version
	for _, stmt := range cg.ast.Children {
		cg.generateStatement(stmt)
	}

	cg.writeConstantsPool()
	cg.Bytecode = append(cg.Bytecode, cg.mainBytecode...)

	return ""
}

func (cg *CodeGenerator) writeConstantsPool() {
	// write constants pool
	cg.Bytecode = append(cg.Bytecode, byte(len(cg.constants)))
	for _, constant := range cg.constants {
		fmt.Printf("Writing constant to bytecode: %#v \n", constant)
		cg.Bytecode = append(cg.Bytecode, byte(constant.TypeId))
		switch constant.TypeId {
		case IntTypeId:
			cg.Bytecode = append(cg.Bytecode, byte(constant.Value.(int64)))
		}
	}
}

func (cg *CodeGenerator) generateStatement(stmt parser.Node) string {
	switch node := stmt.(type) {
	case parser.VariableDeclarationNode:
		return cg.generateVariableDeclaration(node)
	case parser.AssignmentNode:
		cg.generateExpression(node.Right)
		var varName string
		switch node.Left.(type) {
		case parser.IdentifierNode:
			varName = node.Left.(parser.IdentifierNode).Name
			cg.emit(opcode.STORE_VAR, cg.variables[varName])
		default:
			fmt.Printf("Invalid assignment target at position %+v\n", node.Left.Pos())
		}
	case parser.BinaryOpNode:
		cg.generateExpression(node)
	case parser.IfStatementNode:
		cg.generateExpression(node.Condition)
		cg.emit(opcode.JUMP_IF_FALSE, 0)
		thenStart := len(cg.mainBytecode)
		cg.generateStatement(node.ThenBranch)

		if node.ElseBranch != nil {
			elseStart := len(cg.mainBytecode)
			cg.mainBytecode[thenStart-1] = byte(elseStart)
			cg.generateStatement(node.ElseBranch)

			fmt.Printf("If statement with else: thenStart=%d, elseStart=%d, end=%d\n", thenStart, elseStart, len(cg.mainBytecode))
		} else {
			cg.mainBytecode[thenStart-1] = byte(len(cg.mainBytecode))
		}
	case *parser.BlockNode:
		if node != nil {
			for _, statement := range node.Statements {
				cg.generateStatement(statement)
			}
		}
	case parser.BlockNode:
		for _, statement := range node.Statements {
			cg.generateStatement(statement)
		}
	default:
		fmt.Printf("codegen: Unknown statement type: %T at position %+v\n", node, node.Pos())
	}

	return ""
}

func (cg *CodeGenerator) generateVariableDeclaration(node parser.VariableDeclarationNode) string {
	varIdx := cg.AddVariable(node.Name)

	if node.Initializer != nil {
		cg.generateExpression(node.Initializer)
	}

	cg.emit(opcode.STORE_VAR, varIdx)
	return ""
}

func (cg *CodeGenerator) generateExpression(expr parser.Node) string {
	switch node := expr.(type) {
	case parser.NumberNode:
		intValue, err := strconv.ParseInt(node.Value.(string), 10, 64)
		if err != nil {
			fmt.Printf("Error parsing number '%s' at position %+v: %v\n", node.Value, node.Pos(), err)
			return ""
		}

		constIdx := cg.AddConstant(IntTypeId, intValue)

		cg.emit(opcode.LOAD_CONST, constIdx)
	case parser.IdentifierNode:
		if varIdx, exists := cg.variables[node.Name]; exists {
			cg.emit(opcode.LOAD_VAR, varIdx)
		} else {
			fmt.Printf("Undefined variable '%s' at position %+v\n", node.Name, node.Pos())
		}
	case parser.BinaryOpNode:
		cg.generateExpression(node.Left)
		cg.generateExpression(node.Right)
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
			fmt.Printf("Unknown binary operator '%s' at position %+v\n", op, node.Pos())
		}

	default:
		fmt.Printf("Unknown expression type: %T at position %+v\n", node, node.Pos())
	}
	return ""
}

func (cg *CodeGenerator) emit(op opcode.Opcode, operands ...int) {
	fmt.Printf("Emitting opcode: %d with operands %v\n", op, operands)

	switch op {
	case opcode.LOAD_CONST, opcode.STORE_VAR, opcode.LOAD_VAR:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.ADD, opcode.SUB, opcode.MUL, opcode.DIV, opcode.EQ, opcode.LT, opcode.GT:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	case opcode.JUMP_IF_FALSE, opcode.JUMP_IF_TRUE:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	}
}
