package analyzer

import "fmt"

type VariableInfo struct {
	Name  string
	Type  string
	Index int
}

type SymbolTable struct {
	parent  *SymbolTable
	symbols map[string]VariableInfo
}

func NewSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{parent: parent, symbols: make(map[string]VariableInfo)}
}

func (st *SymbolTable) Lookup(name string) (VariableInfo, bool) {
	if info, exists := st.symbols[name]; exists {
		return info, true
	}
	if st.parent != nil {
		return st.parent.Lookup(name)
	}
	return VariableInfo{}, false
}

func (st *SymbolTable) Insert(name string, varType string) error {
	if _, exists := st.symbols[name]; exists {
		return fmt.Errorf("variable '%s' already declared in this scope", name)
	}
	st.symbols[name] = VariableInfo{Name: name, Type: varType, Index: len(st.symbols)}
	return nil
}

func (st *SymbolTable) Print() {
	for name, info := range st.symbols {
		println("Variable:", name, "Type:", info.Type)
	}
	if st.parent != nil {
		println("Parent Symbol Table:")
		st.parent.Print()
	}
}
