package lib

import (
	"plaid/lang"
)

var Conv = lang.BuildNativeModule("conv", map[string]lang.Type{
	"intToStr": lang.TypeFunction{
		Params: lang.TypeTuple{[]lang.Type{
			lang.TypeNativeInt,
		}},
		Ret: lang.TypeNativeStr,
	},
})
