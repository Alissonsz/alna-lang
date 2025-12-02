package symboltable

import "fmt"

type VariableInfo struct {
	Name  string
	Type  string
	Index int
}

type SymbolTable struct {
	Parent  *SymbolTable
	symbols map[string]VariableInfo
	Global  bool
}

func NewSymbolTable(parent *SymbolTable, isGlobal bool) *SymbolTable {
	return &SymbolTable{Parent: parent, symbols: make(map[string]VariableInfo), Global: isGlobal}
}

func (st *SymbolTable) Lookup(name string) (VariableInfo, bool) {
	if info, exists := st.symbols[name]; exists {
		return info, true
	}

	fmt.Println("Looking up in parent symbol table")
	if st.Parent != nil {
		return st.Parent.Lookup(name)
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
	if st.Parent != nil {
		println("Parent Symbol Table:")
		st.Parent.Print()
	}
}
