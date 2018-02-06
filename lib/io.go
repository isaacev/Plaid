package lib

import (
	"fmt"
	"plaid/lang"
	"plaid/lang/types"
)

func IO() *lang.Library {
	lib := lang.MakeLibrary("io")

	lib.Function("print", types.Function{
		Params: types.Tuple{[]types.Type{
			types.Any{},
		}},
		Ret: types.Void{},
	}, func(args []lang.Object) (lang.Object, error) {
		if len(args) != 1 {
			err := fmt.Errorf("wanted 1 argument, got %d", len(args))
			return lang.ObjectNone{}, err
		}

		fmt.Println(args[0].Value())
		return lang.ObjectNone{}, nil
	})

	return lib
}
