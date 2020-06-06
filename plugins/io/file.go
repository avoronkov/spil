package main

import (
	"fmt"
	"io"
	"os"

	"github.com/avoronkov/spil/types"
)

var TypeFile = types.Type("file")

type IoFile struct {
	name string
	file *os.File
}

var _ types.Expr = (*IoFile)(nil)
var _ io.Closer = (*IoFile)(nil)

func (f *IoFile) String() string {
	return fmt.Sprintf("{io.File: %v}", f.name)
}

func (f *IoFile) Print(w io.Writer) {
	fmt.Fprintf(w, "%v", f.String())
}

func (f *IoFile) Hash() (string, error) {
	return "", fmt.Errorf("Hashing is not supported for io.File")
}

func (f *IoFile) Type() types.Type {
	return TypeFile
}

func (f *IoFile) Close() error {
	fmt.Fprintf(os.Stderr, "Closing file %v...\n", f.name)
	return f.file.Close()
}
