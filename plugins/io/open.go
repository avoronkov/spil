package main

import (
	"fmt"
	"os"

	"github.com/avoronkov/spil/types"
)

// Open filename for writing. File is truncated.
// (io.openfile "filename")
type IoOpen struct {
}

var _ types.Function = (*IoOpen)(nil)

func (f *IoOpen) Eval(args []types.Value) (*types.Value, error) {
	if err := f.checkParams(args); err != nil {
		return nil, err
	}
	filename := string(args[0].E.(types.Str))
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &types.Value{
		E: &IoFile{name: filename, file: file},
		T: TypeFile,
	}, nil
}

func (f *IoOpen) ReturnType() types.Type {
	return TypeFile
}

// (openfile "filename" "mode")
func (f *IoOpen) TryBindAll(params []types.Value) (types.Type, error) {
	if err := f.checkParams(params); err != nil {
		return "", err
	}
	return TypeFile, nil
}

func (f *IoOpen) checkParams(params []types.Value) error {
	if len(params) != 1 {
		return fmt.Errorf("io.openfile expects 1 arguments, found: %v", params)
	}
	for i, p := range params {
		if p.T != types.TypeStr {
			return fmt.Errorf("io.openfile expects argument %v to be a string, found: %v", i+1, p)
		}
	}
	return nil
}
