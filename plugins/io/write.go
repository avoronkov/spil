package main

import (
	"fmt"

	"github.com/avoronkov/spil/types"
)

// Write object into file
// (io.write :file objects...)
type IoWrite struct {
}

var _ types.Function = (*IoWrite)(nil)

func (f *IoWrite) Eval(args []types.Value) (*types.Value, error) {
	if err := f.checkParams(args); err != nil {
		return nil, err
	}
	file := args[0].E.(*IoFile)
	for _, arg := range args[1:] {
		arg.E.Print(file.file)
	}
	return &types.Value{
		T: types.TypeAny,
		E: types.QEmpty,
	}, nil
}

func (f *IoWrite) ReturnType() types.Type {
	return types.TypeAny
}

func (f *IoWrite) TryBindAll(params []types.Value) (types.Type, error) {
	if err := f.checkParams(params); err != nil {
		return "", err
	}
	return types.TypeAny, nil
}

func (f *IoWrite) checkParams(params []types.Value) error {
	if len(params) < 2 {
		return fmt.Errorf("io.write expects at least 2 arguments, found: %v", params)
	}
	if params[0].T != TypeFile {
		return fmt.Errorf("io.write expects at first argument to be :file, found: %v", params[0])
	}
	return nil
}
