package vm

import "alna-lang/internal/builtins"

func (vm *VM) registerBuiltins() (map[string]builtins.Function, []FunctionDefinition) {
	builtinFn := builtins.GetBuiltins()
	builtinList := make([]FunctionDefinition, len(builtinFn))
	i := 0
	for _, fn := range builtinFn {
		builtinList[i] = FunctionDefinition{
			Name:           "", // Name is not used in this context
			Implementation: fn,
			Type:           FunctionTypeBuiltin,
		}

		i++
	}

	vm.Functions = append(vm.Functions, builtinList...)
	return builtinFn, builtinList
}
