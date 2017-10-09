package libs

import (
	"fmt"
	"plaid/types"
	"plaid/vm"
)

// IO exposes builtin functions used for input and output
var IO = map[string]*vm.Builtin{
	"print": &vm.Builtin{
		Type: types.TypeFunction{
			Params: types.TypeTuple{Children: []types.Type{
				types.TypeIdent{Name: "Str"},
			}},
			Ret: types.TypeVoid{},
		},
		Func: func(args []vm.Object) (vm.Object, error) {
			if len(args) != 1 {
				err := fmt.Errorf("wanted 1 argument, got %d", len(args))
				return &vm.ObjectNone{}, err
			}

			if str, ok := args[0].(*vm.ObjectStr); ok {
				fmt.Println(str.String())
				return &vm.ObjectNone{}, nil
			}

			err := fmt.Errorf("wanted Str, got %T", args[0])
			return &vm.ObjectNone{}, err
		},
	},
}
