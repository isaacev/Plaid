package lib

import (
	"plaid/lang"
	"plaid/lang/types"
)

var IO = lang.BuildNativeModule("IO", map[string]types.Type{
	"print": types.TypeFunction{
		Params: types.TypeTuple{[]types.Type{
			types.TypeAny{},
		}},
		Ret: types.TypeVoid{},
	},
})
