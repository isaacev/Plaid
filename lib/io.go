package lib

import (
	"plaid/lang"
	"plaid/lang/types"
)

var IO = lang.BuildNativeModule("IO", map[string]types.Type{
	"print": types.Function{
		Params: types.Tuple{[]types.Type{
			types.Any{},
		}},
		Ret: types.Void{},
	},
})
