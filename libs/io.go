package libs

import (
	"fmt"
	"plaid/types"
	"plaid/vm"
)

// IO exposes builtin functions used for input and output
var IO = &vm.Module{
	Name: "IO",
	Exports: map[string]*vm.Export{
		"print": &vm.Export{
			Type: types.Function{
				Params: types.Tuple{Children: []types.Type{
					types.Any{},
				}},
				Ret: types.Void{},
			},
			Register: vm.MakeRegisterTemplate("print"),
			Object: &vm.ObjectBuiltin{
				Val: &vm.Builtin{
					Type: types.Function{
						Params: types.Tuple{Children: []types.Type{
							types.Any{},
						}},
						Ret: types.Void{},
					},
					Func: func(args []vm.Object) (vm.Object, error) {
						if len(args) != 1 {
							err := fmt.Errorf("wanted 1 argument, got %d", len(args))
							return &vm.ObjectNone{}, err
						}

						fmt.Println(args[0].String())
						return &vm.ObjectNone{}, nil
					},
				},
			},
		},
	},
}
