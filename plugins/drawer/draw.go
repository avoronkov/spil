package main

import "github.com/avoronkov/spil/types"

// (draw.point <drawer> x y r g b)
type DrawPoint struct{}

var _ types.Function = (*DrawPoint)(nil)

func (f *DrawPoint) Eval(args []types.Value) (*types.Value, error) {
	drawer := args[0].E.(*Drawer)
	x := int(args[1].E.(types.Int).Int64())
	y := int(args[2].E.(types.Int).Int64())
	r := int(args[5].E.(types.Int).Int64())
	g := int(args[6].E.(types.Int).Int64())
	b := int(args[7].E.(types.Int).Int64())
	drawer.DrawPoint(x, y, r, g, b)
	return &types.Value{
		T: types.TypeAny,
		E: types.QEmpty,
	}, nil
}

// (draw.line <drawer> x0 y0 x1 y1 r g b)
type DrawLine struct{}

var _ types.Function = (*DrawLine)(nil)

func (f *DrawLine) Eval(args []types.Value) (*types.Value, error) {
	drawer := args[0].E.(*Drawer)
	x0 := int(args[1].E.(types.Int).Int64())
	y0 := int(args[2].E.(types.Int).Int64())
	x1 := int(args[3].E.(types.Int).Int64())
	y1 := int(args[4].E.(types.Int).Int64())
	r := int(args[5].E.(types.Int).Int64())
	g := int(args[6].E.(types.Int).Int64())
	b := int(args[7].E.(types.Int).Int64())
	drawer.DrawLine(x0, y0, x1, y1, r, g, b)
	return &types.Value{
		T: types.TypeAny,
		E: types.QEmpty,
	}, nil
}
