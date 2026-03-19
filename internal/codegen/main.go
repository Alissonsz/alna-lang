package codegen

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/builtins"
	"alna-lang/internal/logger"
	"alna-lang/internal/opcode"
	symboltable "alna-lang/internal/symbol_table"
	"encoding/json"
	"os"
	"strconv"
)

const IntTypeId = 1
const FunctionTypeId = 2

type ConstantDefinition struct {
	Value  any
	TypeId int
}

type VariableDefinition struct {
	Name  string
	Index int
}

type SourceMapEntry struct {
	Pc        int    `json:"pc"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndColumn int    `json:"endColumn"`
	VarName   string `json:"varName,omitempty"`
}

type VariableInfo struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

type FunctionInfo struct {
	Address int    `json:"address"`
	Name    string `json:"name"`
}

type DebugInfo struct {
	Version     int              `json:"version"`
	SourceFile  string           `json:"sourceFile"`
	SourceLines []string         `json:"sourceLines"`
	Variables   []VariableInfo   `json:"variables"`
	Functions   []FunctionInfo   `json:"functions"`
	SourceMap   []SourceMapEntry `json:"sourceMap"`
}

type CodeGenerator struct {
	ast                ast.RootNode
	sourceLines        []string
	sourceFile         string
	symbolTable        *symboltable.SymbolTable
	constants          []ConstantDefinition
	constantMap        map[interface{}]int
	variables          []interface{}
	variablesMap       map[string]int
	functions          []builtins.Function
	functionsMap       map[string]int
	mainBytecode       []byte
	Bytecode           []byte
	logger             *logger.Logger
	debugMode          bool
	debugInfo          *DebugInfo
	currentSourcePos   ast.Node
	compiledFuncMap    map[string]int
	scopeDepth         int
	functionScopeDepth int
}

func NewCodeGenerator(tree ast.RootNode, srcLines []string, st *symboltable.SymbolTable, lgr *logger.Logger) *CodeGenerator {
	return &CodeGenerator{ast: tree, sourceLines: srcLines, symbolTable: st, constantMap: make(map[interface{}]int), logger: lgr}
}

func (cg *CodeGenerator) SetDebugMode(sourceFile string) {
	cg.debugMode = true
	cg.sourceFile = sourceFile
	cg.debugInfo = &DebugInfo{
		Version:     1,
		SourceFile:  sourceFile,
		SourceLines: cg.sourceLines,
		Variables:   []VariableInfo{},
		Functions:   []FunctionInfo{},
		SourceMap:   []SourceMapEntry{},
	}
}

func (cg *CodeGenerator) AddConstant(typeId int, value interface{}) int {
	if typeId != FunctionTypeId {
		if idx, exists := cg.constantMap[value]; exists {
			return idx
		}
	}

	cg.constants = append(cg.constants, ConstantDefinition{Value: value, TypeId: typeId})

	if typeId != FunctionTypeId {
		cg.constantMap[value] = len(cg.constants) - 1
	}
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

	if cg.debugMode && cg.scopeDepth == 0 {
		cg.debugInfo.Variables = append(cg.debugInfo.Variables, VariableInfo{
			Index: idx,
			Name:  name,
		})
	}
	return idx
}

func (cg *CodeGenerator) Generate() string {
	cg.Bytecode = append(cg.Bytecode, 0x7F, 'A', 'L', 'N')
	cg.Bytecode = append(cg.Bytecode, 0x01, 0x00, 0x00, 0x00)
	cg.Bytecode = append(cg.Bytecode, 0x00, 0x00, 0x00, 0x00)

	builtins := builtins.GetBuiltins()
	cg.functionsMap = make(map[string]int)
	cg.compiledFuncMap = make(map[string]int)
	for name, fn := range builtins {
		cg.functions = append(cg.functions, fn)
		cg.functionsMap[name] = len(cg.functions) - 1
	}

	st := cg.ast.SymbolTable
	cg.logger.Debug("Starting code generation")
	cg.logger.Debug("Symbol Table at root:")
	cg.logger.Debug("%+v", st)

	for _, expr := range cg.ast.Children {
		cg.generateExpression(expr, st)
	}

	cg.writeConstantsPool()
	cg.Bytecode = append(cg.Bytecode, cg.mainBytecode...)

	mainAddress := cg.compiledFuncMap["main"]
	cg.Bytecode[8] = byte(mainAddress & 0xFF)
	cg.Bytecode[9] = byte((mainAddress >> 8) & 0xFF)
	cg.Bytecode[10] = byte((mainAddress >> 16) & 0xFF)
	cg.Bytecode[11] = byte((mainAddress >> 24) & 0xFF)

	return ""
}

func (cg *CodeGenerator) writeConstantsPool() {
	cg.Bytecode = append(cg.Bytecode, byte(len(cg.constants)))
	for _, constant := range cg.constants {
		cg.logger.Debug("Writing constant to bytecode: %#v", constant)
		cg.Bytecode = append(cg.Bytecode, byte(constant.TypeId))
		switch constant.TypeId {
		case IntTypeId:
			cg.Bytecode = append(cg.Bytecode, byte(constant.Value.(int8)))
		case FunctionTypeId:
			instructions := constant.Value.([]byte)
			cg.Bytecode = append(cg.Bytecode, byte(len(instructions)))
			cg.Bytecode = append(cg.Bytecode, instructions...)
		}
	}
}

func (cg *CodeGenerator) generateExpression(node ast.Node, st *symboltable.SymbolTable) string {
	if cg.debugMode && node != nil {
		if block, ok := node.(*ast.BlockNode); ok && block == nil {
		} else {
			cg.setCurrentSourcePos(node)
		}
	}

	switch n := node.(type) {
	case ast.NumberNode:
		intValue, err := strconv.ParseInt(n.Value.(string), 10, 8)
		if err != nil {
			cg.logger.Error("Error parsing number '%s' at position %+v: %v", n.Value, n.Pos(), err)
			return ""
		}
		constIdx := cg.AddConstant(IntTypeId, int8(intValue))
		cg.logger.Debug("Generating LOAD_CONST for number %d at index %d", intValue, constIdx)
		cg.emit(opcode.LOAD_CONST, constIdx)
	case ast.IdentifierNode:
		if varIdx, exists := cg.variablesMap[n.Name]; exists {
			if cg.debugMode {
				cg.emitWithVarName(opcode.LOAD_VAR, n.Name, varIdx)
			} else {
				cg.emit(opcode.LOAD_VAR, varIdx)
			}
		} else {
			cg.logger.Error("Undefined variable '%s' at position %+v", n.Name, n.Pos())
		}
	case ast.VariableDeclarationNode:
		return cg.generateVariableDeclaration(n, st)
	case ast.AssignmentNode:
		cg.generateExpression(n.Right, st)
		var varName string
		switch n.Left.(type) {
		case ast.IdentifierNode:
			varName = n.Left.(ast.IdentifierNode).Name
			if cg.debugMode {
				cg.setCurrentSourcePos(node)
				cg.emitWithVarName(opcode.STORE_VAR, varName, cg.variablesMap[varName])
			} else {
				cg.emit(opcode.STORE_VAR, cg.variablesMap[varName])
			}
		default:
			cg.logger.Error("Invalid assignment target at position %+v", n.Left.Pos())
		}
	case ast.BinaryOpNode:
		cg.generateBinaryExpression(n, st)
	case ast.IfExpressionNode:
		cg.generateBinaryExpression(n.Condition, st)
		if cg.debugMode {
			cg.setCurrentSourcePos(node)
		}
		cg.emit(opcode.JUMP_IF_FALSE, 0)
		thenStart := len(cg.mainBytecode)
		cg.generateExpression(n.ThenBranch, st)
		cg.emit(opcode.JUMP, 0)
		elseJump := len(cg.mainBytecode)

		if n.ElseBranch != nil {
			elseStart := len(cg.mainBytecode)
			cg.mainBytecode[thenStart-1] = byte(elseStart)
			cg.generateExpression(n.ElseBranch, st)
			cg.mainBytecode[elseJump-1] = byte(len(cg.mainBytecode))
		} else {
			cg.mainBytecode[thenStart-1] = byte(len(cg.mainBytecode))
		}
	case *ast.BlockNode:
		if n == nil {
			return ""
		}

		varsBeforeScope := len(cg.variables)
		cg.scopeDepth++
		cg.emit(opcode.START_SCOPE, varsBeforeScope)
		for _, expr := range n.Expressions {
			cg.generateExpression(expr, n.SymbolTable)
		}
		cg.emit(opcode.END_SCOPE)
		cg.variables = cg.variables[:varsBeforeScope]
		cg.scopeDepth--

	case ast.BlockNode:
		cg.logger.Debug("Entering new block scope in codegen")
		varsBeforeScope := len(cg.variables)
		cg.scopeDepth++
		cg.emit(opcode.START_SCOPE, varsBeforeScope)
		for _, expr := range n.Expressions {
			cg.generateExpression(expr, n.SymbolTable)
		}
		cg.emit(opcode.END_SCOPE)
		cg.variables = cg.variables[:varsBeforeScope]
		cg.scopeDepth--
	case ast.FunctionCallNode:
		cg.logger.Debug("Generating function call to '%s'", n.Name)
		cg.logger.Debug("Function is at address: %d", cg.compiledFuncMap[n.Name])
		for _, arg := range n.Arguments {
			cg.generateExpression(arg, st)
		}
		if cg.debugMode {
			cg.setCurrentSourcePos(node)
		}
		if fnIdx, exists := cg.functionsMap[n.Name]; exists {
			cg.emit(opcode.CALL_BUILTIN, fnIdx)
		} else {

			if fnIdx, exists := cg.compiledFuncMap[n.Name]; exists {
				cg.emit(opcode.CALL, fnIdx)
			} else {
				cg.logger.Error("Undefined function '%s' at position %+v", n.Name, n.Pos())
			}

		}
	case ast.FunctionDeclarationNode:
		return cg.generateFunctionDeclaration(n, st)
	case ast.ReturnNode:
		scopesToClose := cg.scopeDepth - cg.functionScopeDepth
		for i := 0; i < scopesToClose; i++ {
			cg.emit(opcode.END_SCOPE)
		}
		cg.emit(opcode.RETURN)
	default:
		cg.logger.Warn("Unknown expression type: %T at position %+v", node, node.Pos())
	}

	return ""
}

func (cg *CodeGenerator) generateFunctionDeclaration(node ast.FunctionDeclarationNode, st *symboltable.SymbolTable) string {
	cg.logger.Debug("Generating function declaration for '%s'", node.Name)

	functionStartPos := len(cg.mainBytecode)
	cg.functionScopeDepth = cg.scopeDepth

	savedVariablesMap := cg.variablesMap
	savedVariables := cg.variables
	cg.variablesMap = make(map[string]int)
	cg.variables = nil

	cg.scopeDepth++
	cg.emit(opcode.START_SCOPE, 0)

	cg.logger.Debug("Adding function's %d parameters to variable map", len(node.Parameters))
	for _, param := range node.Parameters {
		varIdx := cg.AddVariable(param.Name)
		cg.emitWithVarName(opcode.STORE_VAR, param.Name, varIdx)
	}

	for _, expr := range node.Body.Expressions {
		cg.logger.Debug("Generating function body expression")
		cg.generateExpression(expr, node.Body.SymbolTable)
	}
	cg.emit(opcode.END_SCOPE)
	cg.emit(opcode.RETURN)
	cg.scopeDepth--

	cg.variablesMap = savedVariablesMap
	cg.variables = savedVariables

	cg.logger.Debug("Saving function '%s' start position at bytecode index %d", node.Name, functionStartPos)
	cg.compiledFuncMap[node.Name] = functionStartPos

	return ""
}

func (cg *CodeGenerator) generateVariableDeclaration(node ast.VariableDeclarationNode, st *symboltable.SymbolTable) string {
	varIdx := cg.AddVariable(node.Name)

	if node.Initializer != nil {
		cg.generateExpression(node.Initializer, st)
	}

	if cg.debugMode {
		cg.setCurrentSourcePos(node)
		cg.emitWithVarName(opcode.STORE_VAR, node.Name, varIdx)
	} else {
		cg.emit(opcode.STORE_VAR, varIdx)
	}
	return ""
}

func (cg *CodeGenerator) generateBinaryExpression(expr ast.Node, st *symboltable.SymbolTable) string {
	if cg.debugMode && expr != nil {
		cg.setCurrentSourcePos(expr)
	}

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
			if cg.debugMode {
				cg.emitWithVarName(opcode.LOAD_VAR, node.Name, varIdx)
			} else {
				cg.emit(opcode.LOAD_VAR, varIdx)
			}
		} else {
			cg.logger.Error("Undefined variable '%s' at position %+v", node.Name, node.Pos())
		}
	case ast.BinaryOpNode:
		cg.generateBinaryExpression(node.Left, st)
		cg.generateBinaryExpression(node.Right, st)

		if cg.debugMode {
			cg.setCurrentSourcePos(node)
		}

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
	case ast.FunctionCallNode:
		cg.logger.Debug("Generating function call to '%s'", node.Name)
		for _, arg := range node.Arguments {
			cg.generateExpression(arg, st)
		}
		if cg.debugMode {
			cg.setCurrentSourcePos(node)
		}
		if fnIdx, exists := cg.functionsMap[node.Name]; exists {
			cg.emit(opcode.CALL, fnIdx)
		} else {
			if fnIdx, exists := cg.compiledFuncMap[node.Name]; exists {
				cg.emit(opcode.CALL, fnIdx)
			} else {
				cg.logger.Error("Undefined function '%s' at position %+v", node.Name, node.Pos())
			}
		}

	default:
		cg.logger.Warn("Unknown binary expression type: %T at position %+v", node, node.Pos())
	}
	return ""
}

func (cg *CodeGenerator) emit(op opcode.Opcode, operands ...int) {
	cg.emitWithVarName(op, "", operands...)
}

func (cg *CodeGenerator) emitWithVarName(op opcode.Opcode, varName string, operands ...int) {
	cg.logger.Debug("Emitting opcode: %s with operands %v", op, operands)

	pc := len(cg.mainBytecode)

	switch op {
	case opcode.LOAD_CONST, opcode.STORE_VAR, opcode.LOAD_VAR:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.ADD, opcode.SUB, opcode.MUL, opcode.DIV, opcode.EQ, opcode.LT, opcode.GT:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	case opcode.JUMP_IF_FALSE, opcode.JUMP, opcode.JUMP_IF_TRUE, opcode.CALL, opcode.CALL_BUILTIN:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.START_SCOPE:
		cg.mainBytecode = append(cg.mainBytecode, byte(op), byte(operands[0]))
	case opcode.END_SCOPE:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	case opcode.RETURN:
		cg.mainBytecode = append(cg.mainBytecode, byte(op))
	default:
		cg.logger.Error("Unknown opcode to emit: %d", op)
	}

	if cg.debugMode && cg.currentSourcePos != nil {
		pos := cg.currentSourcePos.Pos()
		if pos.Line > 0 && pos.Column >= 0 {
			entry := SourceMapEntry{
				Pc:        pc,
				Line:      pos.Line - 1,
				Column:    pos.Column,
				EndColumn: pos.EndColumn,
			}
			if varName != "" {
				entry.VarName = varName
			}
			cg.debugInfo.SourceMap = append(cg.debugInfo.SourceMap, entry)
		}
	}
}

func (cg *CodeGenerator) setCurrentSourcePos(node ast.Node) {
	cg.currentSourcePos = node
}

func (cg *CodeGenerator) WriteDebugFile(outputPath string) error {
	if !cg.debugMode || cg.debugInfo == nil {
		return nil
	}

	compiledFunctions := make([]FunctionInfo, 0, len(cg.compiledFuncMap))
	for name, addr := range cg.compiledFuncMap {
		compiledFunctions = append(compiledFunctions, FunctionInfo{
			Name:    name,
			Address: addr,
		})
	}
	cg.debugInfo.Functions = compiledFunctions

	data, err := json.MarshalIndent(cg.debugInfo, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}
