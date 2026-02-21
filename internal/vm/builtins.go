package vm

import "alna-lang/internal/builtins"

func (vm *VM) registerBuiltins() (map[string]builtins.Function, []FunctionDefinition) {
	builtinFn := builtins.GetBuiltins()
	var builtinList []FunctionDefinition
	for name, fn := range builtinFn {
		builtinList = append(builtinList, FunctionDefinition{
			Name:           name,
			Implementation: fn,
			Type:           FunctionTypeBuiltin,
		})
	}

	vm.Functions = append(vm.Functions, builtinList...)
	return builtinFn, builtinList
}
