package parser

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
	case HighPrecedenceNode:
		fmt.Printf("%s%sHighPrecedence (%v)\n", indent, connector, n.Operator.Value)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Left, childIndent, false)
		PrintAST(n.Right, childIndent, true)
	case LowPrecedenceNode:
		fmt.Printf("%s%sLowPrecedence (%v)\n", indent, connector, n.Operator.Value)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Left, childIndent, false)
		PrintAST(n.Right, childIndent, true)
	case ParenthisedNode:
		fmt.Printf("%s%sParenthised\n", indent, connector)
		childIndent := indent
		if isLast {
			childIndent += "    "
		} else {
			childIndent += "│   "
		}
		PrintAST(n.Expression, childIndent, true)
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
	default:
		fmt.Printf("%s%sUnknown Node Type\n", indent, connector)
	}
}
