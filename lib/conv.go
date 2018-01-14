package lib

import (
	"plaid/lang"
	"plaid/lang/types"
)

var Conv = lang.BuildNativeModule("conv", map[string]types.Type{
	"intToStr": types.TypeFunction{
		Params: types.TypeTuple{[]types.Type{
			types.TypeNativeInt,
		}},
		Ret: types.TypeNativeStr,
	},
})
