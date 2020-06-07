package main

import "github.com/avoronkov/spil/types"

// Declare type ":file"
var Types = map[types.Type]types.Type{
	TypeFile: types.TypeAny,
}

var Funcs = map[string]types.Function{
	"io.openfile": new(IoOpen),
	"io.write":    new(IoWrite),
}
