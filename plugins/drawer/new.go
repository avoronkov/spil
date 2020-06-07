package main

import (
	"github.com/avoronkov/spil/types"
)

// (drawer.new "filename" width height)
type DrawerNew struct{}

var _ types.Function = (*DrawerNew)(nil)

func (f *DrawerNew) Eval(args []types.Value) (*types.Value, error) {
	filename := string(args[0].E.(types.Str))
	width := args[1].E.(types.Int).Int64()
	height := args[2].E.(types.Int).Int64()
	return &types.Value{
		T: TypeDrawer,
		E: NewDrawer(filename, int(width), int(height)),
	}, nil
}
