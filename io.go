package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type LazyInput struct {
	input      *bufio.Reader
	valueReady bool
	value      Expr
	tailCalled bool
}

var _ List = (*LazyInput)(nil)

func NewLazyInput() *LazyInput {
	return &LazyInput{
		input: bufio.NewReader(os.Stdin),
	}
}

func (i *LazyInput) Head() (Expr, error) {
	log.Printf("Input.Head()")
	if err := i.next(); err != nil {
		return nil, err
	}
	if i.value == nil {
		return nil, fmt.Errorf("Input: cannot perform Head() on empty stream")
	}
	return i.value, nil
}

func (i *LazyInput) Tail() (List, error) {
	log.Printf("Input.Tail()")
	if err := i.next(); err != nil {
		return nil, err
	}
	if i.value == nil {
		return nil, fmt.Errorf("Input: cannot perform Tail() on empty stream")
	}
	if i.tailCalled {
		return nil, fmt.Errorf("Input: cannot perform Tail() twice")
	}
	i.tailCalled = true
	return &LazyInput{
		input: i.input,
	}, nil

}

func (i *LazyInput) Empty() bool {
	log.Printf("Input.Empty()")
	if err := i.next(); err != nil {
		panic(err)
	}
	return i.value != nil
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
	i.value = Str(string([]byte{b}))
	return nil
}

func (i *LazyInput) String() string {
	if i.Empty() {
		return "{Input: }"
	}
	return "{Input: ...}"
}

func (i *LazyInput) Print(w io.Writer) {
	i.tailCalled = true
	io.Copy(w, i.input)
}

func (i *LazyInput) Hash() (string, error) {
	return "", fmt.Errorf("Hash() is not applicable for LazyInput")
}
