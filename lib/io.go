package lib

import (
	"plaid/lang"
)

var IO = lang.BuildNativeModule("IO", map[string]lang.Type{
	"print": lang.TypeFunction{
		Params: lang.TypeTuple{[]lang.Type{
			lang.TypeAny{},
		}},
		Ret: lang.TypeVoid{},
	},
})
