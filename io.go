package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/avoronkov/spil/types"
)

type LazyInput struct {
	file       io.ReadCloser
	input      *bufio.Reader
	valueReady bool
	value      *types.Value
	tail       *LazyInput
}

var _ types.List = (*LazyInput)(nil)

func NewLazyInput(f io.ReadCloser) *LazyInput {
	return &LazyInput{
		file:  f,
		input: bufio.NewReader(f),
	}
}

func (i *LazyInput) Head() (*types.Value, error) {
	if err := i.next(); err != nil {
		return nil, err
	}
	if i.value == nil {
		return nil, fmt.Errorf("Input: cannot perform Head() on empty stream")
	}
	return i.value, nil
}

func (i *LazyInput) Tail() (types.List, error) {
	if err := i.next(); err != nil {
		return nil, err
	}
	if i.value == nil {
		return nil, fmt.Errorf("Input: cannot perform Tail() on empty stream")
	}
	if i.tail == nil {
		i.tail = &LazyInput{
			input: i.input,
		}
	}
	return i.tail, nil
}

func (i *LazyInput) Empty() (result bool) {
	if err := i.next(); err != nil {
		panic(err)
	}
	return i.value == nil
}

func (i *LazyInput) next() error {
	if i.valueReady {
		return nil
	}
	i.valueReady = true
	b, err := i.input.ReadByte()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	i.value = &types.Value{E: types.Str(string([]byte{b})), T: types.TypeStr}
	return nil
}

func (i *LazyInput) String() string {
	if i.Empty() {
		return "{Input: }"
	}
	return "{Input: ...}"
}

func (i *LazyInput) Print(w io.Writer) {
	if i.Empty() {
		return
	}
	h, _ := i.Head()
	io.WriteString(w, string(h.E.(types.Str)))
	t, _ := i.Tail()
	t.Print(w)
}

func (i *LazyInput) Hash() (string, error) {
	return "", fmt.Errorf("Hash() is not applicable for LazyInput")
}

func (i *LazyInput) Close() error {
	if i.file != nil {
		return i.file.Close()
	}
	return nil
}

func (i *LazyInput) Type() types.Type {
	return types.TypeStr
}
