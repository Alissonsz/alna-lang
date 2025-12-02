package builtins

import "fmt"

type Function = func(args ...any) any

func GetBuiltins() map[string]Function {
	builtins := map[string]Function{
		"__write": func(args ...any) any {
			if len(args) != 1 {
				panic("Invalid number of arguments to __write")
			}

			fmt.Println(args[0])
			return nil
		},
	}
	return builtins
}
