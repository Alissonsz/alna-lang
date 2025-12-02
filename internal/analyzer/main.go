package analyzer

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/common"
	"alna-lang/internal/logger"
	"alna-lang/internal/symbol_table"
	"fmt"
)

type Analyzer struct {
	ast         *ast.RootNode
	SymbolTable *symboltable.SymbolTable
	sourceLines []string
	logger      *logger.Logger
}

func NewAnalyzer(tree *ast.RootNode, srcLines []string, lgr *logger.Logger) *Analyzer {
	tree.SymbolTable = symboltable.NewSymbolTable(nil, true)
	return &Analyzer{ast: tree, SymbolTable: symboltable.NewSymbolTable(nil, true), sourceLines: srcLines, logger: lgr}
}

func (a *Analyzer) Analyze() error {
	for _, stmt := range a.ast.Children {
		if err := a.analyzeStatement(stmt, a.SymbolTable); err != nil {
			panic(common.CompilerError(stmt.Pos(), err.Error(), a.sourceLines))
		}
	}

	return nil
}

func (a *Analyzer) analyzeStatement(stmt ast.Node, st *symboltable.SymbolTable) error {
	switch node := stmt.(type) {
	case ast.IfStatementNode:
		if err := a.analyzeExpression(node.Condition, st); err != nil {
			return err
		}
		if err := a.analyzeStatement(node.ThenBranch, st); err != nil {
			return err
		}
		if node.ElseBranch != nil {
			if err := a.analyzeStatement(node.ElseBranch, st); err != nil {
				return err
			}
		}
	case *ast.BlockNode:
		if node != nil {
			newSt := symboltable.NewSymbolTable(st, false)
			newSt.Parent = st

			a.logger.Debug("Entering new block scope")
			a.logger.Debug("Symbol table: %+v", newSt)
			node.SymbolTable = newSt

			for _, statement := range node.Statements {
				if err := a.analyzeStatement(statement, newSt); err != nil {
					return err
				}
			}
		}
	case ast.BlockNode:
		newSt := symboltable.NewSymbolTable(st, false)
		newSt.Parent = st
		a.logger.Debug("Entering new block scope")
		a.logger.Debug("Symbol table: %+v", newSt)

		node.SymbolTable = newSt

		for _, statement := range node.Statements {
			if err := a.analyzeStatement(statement, newSt); err != nil {
				return err
			}
		}
	case ast.VariableDeclarationNode:
		st.Insert(node.Name, node.Type)
	case ast.AssignmentNode:
		var varName string
		switch node.Left.(type) {
		case ast.IdentifierNode:
			varName = node.Left.(ast.IdentifierNode).Name
		default:
			return fmt.Errorf("invalid assignment target at position %+v", node.Left.Pos())
		}

		if _, exists := st.Lookup(varName); !exists {
			return fmt.Errorf("undefined variable '%s' at position %+v", varName, node.Left.Pos())
		}

		if err := a.analyzeStatement(node.Right, st); err != nil {
			return err
		}
	case ast.BinaryOpNode, ast.NumberNode, ast.BooleanNode, ast.IdentifierNode:
		return a.analyzeExpression(node, st)
	default:
		a.logger.Warn("Unknown statement type: %T at position %+v", node, node.Pos())
	}
	return nil
}

func (a *Analyzer) analyzeExpression(expr ast.Node, st *symboltable.SymbolTable) error {
	a.logger.Debug("Analyzing expression: %T at position %+v", expr, expr.Pos())
	switch node := expr.(type) {
	case ast.NumberNode:
		return nil
	case ast.BooleanNode:
		return nil
	case ast.IdentifierNode:
		if _, exists := st.Lookup(node.Name); !exists {
			return fmt.Errorf("undefined variable '%s' at position %+v", node.Name, node.Pos())
		}
		return nil
	case ast.BinaryOpNode:
		if err := a.analyzeExpression(node.Left, st); err != nil {
			return err
		}
		if err := a.analyzeExpression(node.Right, st); err != nil {
			return err
		}

		_, err := a.inferType(node, st)
		return err
	default:
		return fmt.Errorf("unknown expression type: %T at position %+v", node, node.Pos())
	}
}

func (a *Analyzer) inferType(expr ast.Node, st *symboltable.SymbolTable) (string, error) {
	switch node := expr.(type) {
	case ast.NumberNode:
		return "int", nil
	case ast.BooleanNode:
		return "bool", nil
	case ast.IdentifierNode:
		if varInfo, exists := st.Lookup(node.Name); exists {
			return varInfo.Type, nil
		}
		return "", fmt.Errorf("undefined variable '%s' at position %+v", node.Name, node.Pos())
	case ast.BinaryOpNode:
		leftType, err := a.inferType(node.Left, st)
		if err != nil {
			return "", err
		}
		rightType, err := a.inferType(node.Right, st)
		if err != nil {
			return "", err
		}
		if leftType != rightType {
			return "", fmt.Errorf("type mismatch: %s vs %s at position %+v", leftType, rightType, node.Pos())
		}
		return leftType, nil
	default:
		return "", fmt.Errorf("unknown expression type: %T at position %+v", node, node.Pos())
	}
}

func (a *Analyzer) PrintSymbolTable() {
	fmt.Println("Symbol Table:")
	a.SymbolTable.Print()
}
