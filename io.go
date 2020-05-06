package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

type LazyInput struct {
	file       io.ReadCloser
	input      *bufio.Reader
	valueReady bool
	value      Expr
	tail       *LazyInput
}

var _ List = (*LazyInput)(nil)

func NewLazyInput(f io.ReadCloser) *LazyInput {
	return &LazyInput{
		file:  f,
		input: bufio.NewReader(f),
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
	if i.tail == nil {
		i.tail = &LazyInput{
			input: i.input,
		}
	}
	return i.tail, nil
}

func (i *LazyInput) Empty() (result bool) {
	defer func() {
		log.Printf("Input.Empty(): %v", result)
	}()
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
	if i.Empty() {
		return
	}
	h, _ := i.Head()
	io.WriteString(w, string(h.(Str)))
	t, _ := i.Tail()
	t.Print(w)
}

func (i *LazyInput) Hash() (string, error) {
	return "", fmt.Errorf("Hash() is not applicable for LazyInput")
}

func (i *LazyInput) Close() error {
	log.Printf("Closing file")
	if i.file != nil {
		return i.file.Close()
	}
	return nil
}
