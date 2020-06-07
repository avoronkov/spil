package main

import "github.com/avoronkov/spil/types"

var Types = map[types.Type]types.Type{
	TypeDrawer: types.TypeAny,
}

var Funcs = map[string]types.Function{
	"drawer.new.native": new(DrawerNew),
	"draw.point.native": new(DrawPoint),
	"draw.line.native":  new(DrawLine),
}
