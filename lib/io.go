package lib

import (
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
		return nil, nil
	})

	return lib
}
