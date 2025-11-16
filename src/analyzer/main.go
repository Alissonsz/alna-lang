package analyzer

import (
	"alna-lang/src/common"
	"alna-lang/src/parser"
	"alna-lang/src/symbol_table"
	"fmt"
)

type Analyzer struct {
	ast         parser.RootNode
	SymbolTable *symboltable.SymbolTable
	sourceLines []string
}

func NewAnalyzer(ast parser.RootNode, srcLines []string) *Analyzer {
	return &Analyzer{ast: ast, SymbolTable: symboltable.NewSymbolTable(nil), sourceLines: srcLines}
}

func (a *Analyzer) Analyze() error {
	for _, stmt := range a.ast.Children {
		if err := a.analyzeStatement(stmt, a.SymbolTable); err != nil {
			panic(common.CompilerError(stmt.Pos(), err.Error(), a.sourceLines))
		}
	}

	return nil
}

func (a *Analyzer) analyzeStatement(stmt parser.Node, st *symboltable.SymbolTable) error {
	switch node := stmt.(type) {
	case parser.IfStatementNode:
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
	case *parser.BlockNode:
		if node != nil {
			newSt := symboltable.NewSymbolTable(st)
			newSt.Parent = st

			fmt.Println("Entering new block scope")
			newSt.Print()

			for _, statement := range node.Statements {
				if err := a.analyzeStatement(statement, newSt); err != nil {
					return err
				}
			}
		}
	case parser.BlockNode:
		newSt := symboltable.NewSymbolTable(st)
		newSt.Parent = st
		fmt.Println("Entering new block scope")
		newSt.Print()

		for _, statement := range node.Statements {
			if err := a.analyzeStatement(statement, newSt); err != nil {
				return err
			}
		}
	case parser.VariableDeclarationNode:
		st.Insert(node.Name, node.Type)
	case parser.AssignmentNode:
		var varName string
		switch node.Left.(type) {
		case parser.IdentifierNode:
			varName = node.Left.(parser.IdentifierNode).Name
		default:
			return fmt.Errorf("invalid assignment target at position %+v", node.Left.Pos())
		}

		if _, exists := st.Lookup(varName); !exists {
			return fmt.Errorf("undefined variable '%s' at position %+v", varName, node.Left.Pos())
		}

		if err := a.analyzeStatement(node.Right, st); err != nil {
			return err
		}
	case parser.BinaryOpNode, parser.NumberNode, parser.BooleanNode, parser.IdentifierNode:
		return a.analyzeExpression(node, st)
	default:
		fmt.Printf("Unknown statement type: %T at position %+v\n", node, node.Pos())
	}
	return nil
}

func (a *Analyzer) analyzeExpression(expr parser.Node, st *symboltable.SymbolTable) error {
	fmt.Printf("Analyzing expression: %T at position %+v\n", expr, expr.Pos())
	switch node := expr.(type) {
	case parser.NumberNode:
		return nil
	case parser.BooleanNode:
		return nil
	case parser.IdentifierNode:
		if _, exists := st.Lookup(node.Name); !exists {
			return fmt.Errorf("undefined variable '%s' at position %+v", node.Name, node.Pos())
		}
		return nil
	case parser.BinaryOpNode:
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

func (a *Analyzer) inferType(expr parser.Node, st *symboltable.SymbolTable) (string, error) {
	switch node := expr.(type) {
	case parser.NumberNode:
		return "int", nil
	case parser.BooleanNode:
		return "bool", nil
	case parser.IdentifierNode:
		if varInfo, exists := st.Lookup(node.Name); exists {
			return varInfo.Type, nil
		}
		return "", fmt.Errorf("undefined variable '%s' at position %+v", node.Name, node.Pos())
	case parser.BinaryOpNode:
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
