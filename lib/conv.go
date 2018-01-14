package lib

import (
	"plaid/lang"
	"plaid/lang/types"
)

var Conv = lang.BuildNativeModule("conv", map[string]types.Type{
	"intToStr": types.Function{
		Params: types.Tuple{[]types.Type{
			types.BuiltinInt,
		}},
		Ret: types.BuiltinStr,
	},
})
