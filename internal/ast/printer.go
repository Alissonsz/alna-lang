package ast

import (
	"fmt"
)

// PrintAST prints the AST in a tree-like visual format
func PrintAST(node Node, indent string, isLast bool) {
	if node == nil {
		return
	}

	// Determine the connector for this node
	connector := ""
	if indent != "" {
		if isLast {
			connector = "└── "
		} else {
			connector = "├── "
		}
	}

	// Print the node based on its type
	switch n := node.(type) {
	case RootNode:
		fmt.Printf("%s%sRoot\n", indent, connector)
		for i, child := range n.Children {
			childIndent := indent
			if indent != "" {
				if isLast {
					childIndent += "    "
				} else {
					childIndent += "│   "
				}
			}
			PrintAST(child, childIndent, i == len(n.Children)-1)
		}
	case NumberNode:
		fmt.Printf("%s%sNumber: %v\n", indent, connector, n.Value)
	case BooleanNode:
		fmt.Printf("%s%sBoolean: %v\n", indent, connector, n.Value)
	case BinaryOpNode:
		fmt.Printf("%s%sBinaryOp (%v)\n", indent, connector, n.Operator.Value)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Left, childIndent, false)
		PrintAST(n.Right, childIndent, true)
	case IdentifierNode:
		fmt.Printf("%s%sIdentifier: %s\n", indent, connector, n.Name)
	case VariableDeclarationNode:
		fmt.Printf("%s%sVariableDeclaration\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		// Print name
		fmt.Printf("%s├── Name: %s\n", childIndent, n.Name)
		// Print type
		fmt.Printf("%s├── Type: %s\n", childIndent, n.Type)
		// Print initializer (if present)
		if n.Initializer != nil {
			fmt.Printf("%s└── Initializer:\n", childIndent)
			PrintAST(n.Initializer, childIndent+"    ", true)
		} else {
			fmt.Printf("%s└── Initializer: none\n", childIndent)
		}
	case AssignmentNode:
		fmt.Printf("%s%sAssignment\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		// Print target
		fmt.Printf("%s├── Target:\n", childIndent)
		PrintAST(n.Left, childIndent+"│   ", true)
		// Print value
		fmt.Printf("%s└── Value:\n", childIndent)
		PrintAST(n.Right, childIndent+"    ", true)
	case IfStatementNode:
		fmt.Printf("%s%sIfStatement\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		// Print condition
		fmt.Printf("%s├── Condition:\n", childIndent)
		PrintAST(n.Condition, childIndent+"│   ", false)
		// Print then block
		fmt.Printf("%s├── ThenBlock:\n", childIndent)
		PrintAST(n.ThenBranch, childIndent+"│   ", n.ElseBranch == nil)
		// Print else block (if present)
		if n.ElseBranch != nil {
			fmt.Printf("%s└── ElseBlock:\n", childIndent)
			PrintAST(n.ElseBranch, childIndent+"    ", true)
		}
	case BlockNode:
		fmt.Printf("%s%sBlock\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		for i, stmt := range n.Statements {
			PrintAST(stmt, childIndent, i == len(n.Statements)-1)
		}
	case *BlockNode:
		if n == nil {
			return
		}

		fmt.Printf("%s%sBlock\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		for i, stmt := range n.Statements {
			PrintAST(stmt, childIndent, i == len(n.Statements)-1)
		}
	case FunctionCallNode:
		fmt.Printf("%s%sFunctionCall: %s\n", indent, connector, n.Name)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		for i, arg := range n.Arguments {
			PrintAST(arg, childIndent, i == len(n.Arguments)-1)
		}
	default:
		fmt.Printf("%s%sUnknown Node Type\n", indent, connector)
	}
}
