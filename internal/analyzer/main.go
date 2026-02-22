package analyzer

import (
	"alna-lang/internal/ast"
	"alna-lang/internal/builtins"
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
	for _, expr := range a.ast.Children {
		if err := a.analyzeExpression(expr, a.SymbolTable); err != nil {
			panic(common.CompilerError(expr.Pos(), err.Error(), a.sourceLines))
		}
	}

	return nil
}

func (a *Analyzer) analyzeExpression(node ast.Node, st *symboltable.SymbolTable) error {
	switch n := node.(type) {
	case ast.IfExpressionNode:
		if err := a.analyzeBinaryExpression(n.Condition, st); err != nil {
			return err
		}
		if err := a.analyzeExpression(n.ThenBranch, st); err != nil {
			return err
		}
		if n.ElseBranch != nil {
			if err := a.analyzeExpression(n.ElseBranch, st); err != nil {
				return err
			}
		}
	case *ast.BlockNode:
		if n != nil {
			newSt := symboltable.NewSymbolTable(st, false)
			newSt.Parent = st

			a.logger.Debug("Entering new block scope")
			a.logger.Debug("Symbol table: %+v", newSt)
			n.SymbolTable = newSt

			for _, expr := range n.Expressions {
				if err := a.analyzeExpression(expr, newSt); err != nil {
					return err
				}
			}
		}
	case ast.BlockNode:
		newSt := symboltable.NewSymbolTable(st, false)
		newSt.Parent = st
		a.logger.Debug("Entering new block scope")
		a.logger.Debug("Symbol table: %+v", newSt)

		n.SymbolTable = newSt

		for _, expr := range n.Expressions {
			if err := a.analyzeExpression(expr, newSt); err != nil {
				return err
			}
		}
	case ast.VariableDeclarationNode:
		st.Insert(n.Name, n.Type)
	case ast.AssignmentNode:
		var varName string
		switch n.Left.(type) {
		case ast.IdentifierNode:
			varName = n.Left.(ast.IdentifierNode).Name
		default:
			return fmt.Errorf("invalid assignment target at position %+v", n.Left.Pos())
		}

		if _, exists := st.Lookup(varName); !exists {
			return fmt.Errorf("undefined variable '%s' at position %+v", varName, n.Left.Pos())
		}

		if err := a.analyzeExpression(n.Right, st); err != nil {
			return err
		}
	case ast.BinaryOpNode, ast.NumberNode, ast.BooleanNode, ast.IdentifierNode:
		return a.analyzeBinaryExpression(node, st)
	case ast.FunctionDeclarationNode:
		st.Insert(n.Name, "function")
		newSt := symboltable.NewSymbolTable(st, false)
		newSt.Parent = st
		a.logger.Debug("Entering new function scope for '%s'", n.Name)
	case ast.FunctionCallNode:
		if _, exists := st.Lookup(n.Name); !exists {
			if _, exists := builtins.GetBuiltins()[n.Name]; !exists {
				return fmt.Errorf("undefined function '%s' at position %+v", n.Name, n.Pos())
			}
		}
		for _, arg := range n.Arguments {
			if err := a.analyzeBinaryExpression(arg, st); err != nil {
				return err
			}
		}
	default:
		a.logger.Warn("Unknown expression type: %T at position %+v", node, node.Pos())
	}
	return nil
}

func (a *Analyzer) analyzeBinaryExpression(expr ast.Node, st *symboltable.SymbolTable) error {
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
		if err := a.analyzeBinaryExpression(node.Left, st); err != nil {
			return err
		}
		if err := a.analyzeBinaryExpression(node.Right, st); err != nil {
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
