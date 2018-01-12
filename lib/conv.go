package lib

import (
	"plaid/lang"
)

var Conv = lang.BuildNativeModule("conv", map[string]lang.Type{
	"int2str": lang.TypeFunction{
		Params: lang.TypeTuple{[]lang.Type{
			lang.TypeNativeInt,
		}},
		Ret: lang.TypeNativeStr,
	},
})
