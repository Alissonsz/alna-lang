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
	case IfExpressionNode:
		fmt.Printf("%s%sIfExpression\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		fmt.Printf("%s├── Condition:\n", childIndent)
		PrintAST(n.Condition, childIndent+"│   ", false)
		fmt.Printf("%s├── ThenBlock:\n", childIndent)
		PrintAST(n.ThenBranch, childIndent+"│   ", n.ElseBranch == nil)
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
		for i, expr := range n.Expressions {
			PrintAST(expr, childIndent, i == len(n.Expressions)-1)
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
		for i, expr := range n.Expressions {
			PrintAST(expr, childIndent, i == len(n.Expressions)-1)
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
	case FunctionDeclarationNode:
		fmt.Printf("%s%sFunctionDeclaration: %s\n", indent, connector, n.Name)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		// Print parameters
		fmt.Printf("%s├── Parameters:\n", childIndent)
		paramIndent := childIndent + "│   "
		for i, param := range n.Parameters {
			fmt.Printf("%s%sParameter: %s Type: %s\n", paramIndent, func() string {
				if i == len(n.Parameters)-1 {
					return "└── "
				}
				return "├── "
			}(), param.Name, param.Type)
		}
		// Print return type
		fmt.Printf("%s├── ReturnType: %s\n", childIndent, n.ReturnType)
		// Print body
		fmt.Printf("%s└── Body:\n", childIndent)
		PrintAST(n.Body, childIndent+"    ", true)
	case ReturnNode:
		fmt.Printf("%s%sReturn\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Value, childIndent, true)
	default:
		fmt.Printf("%s%sUnknown Node Type\n", indent, connector)
	}
}
