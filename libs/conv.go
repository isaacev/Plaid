package libs

import (
	"fmt"
	"plaid/types"
	"plaid/vm"
)

// Conv exposes builtin functions for type conversions
var Conv = &vm.Module{
	Name: "Conv",
	Exports: map[string]*vm.Export{
		"intToStr": &vm.Export{
			Type: types.TypeFunction{
				Params: types.TypeTuple{Children: []types.Type{
					types.Int,
				}},
				Ret: types.Str,
			},
			Register: vm.MakeRegisterTemplate("intToStr"),
			Object: &vm.ObjectBuiltin{
				Val: &vm.Builtin{
					Type: types.TypeFunction{
						Params: types.TypeTuple{Children: []types.Type{
							types.Int,
						}},
						Ret: types.Str,
					},
					Func: func(args []vm.Object) (vm.Object, error) {
						if len(args) != 1 {
							err := fmt.Errorf("wanted 1 argument, got %d", len(args))
							return &vm.ObjectNone{}, err
						}

						if num, ok := args[0].(*vm.ObjectInt); ok {
							obj := &vm.ObjectStr{Val: fmt.Sprintf("%d", num.Val)}
							return obj, nil
						}

						err := fmt.Errorf("wanted Int, got %T", args[0])
						return &vm.ObjectNone{}, err
					},
				},
			},
		},
	},
}

// var Conv = map[string]*vm.Builtin{
// 	"intToStr": &vm.Builtin{
// 		Type: types.TypeFunction{
// 			Params: types.TypeTuple{Children: []types.Type{
// 				types.Int,
// 			}},
// 			Ret: types.Str,
// 		},
// 		Func: func(args []vm.Object) (vm.Object, error) {
// 			if len(args) != 1 {
// 				err := fmt.Errorf("wanted 1 argument, got %d", len(args))
// 				return &vm.ObjectNone{}, err
// 			}
//
// 			if num, ok := args[0].(*vm.ObjectInt); ok {
// 				obj := &vm.ObjectStr{Val: fmt.Sprintf("%d", num.Val)}
// 				return obj, nil
// 			}
//
// 			err := fmt.Errorf("wanted Int, got %T", args[0])
// 			return &vm.ObjectNone{}, err
// 		},
// 	},
// }
