package main

import (
	"fmt"
	"io"
)

var _ List = (*LazyList)(nil)

type LazyList struct {
	iter       Evaler
	state      Expr
	value      Expr
	valueReady bool
	tail       *LazyList
	id         int64
}

var lazyHashCount int64

func NewLazyList(iter Evaler, state Expr, hashable bool) *LazyList {
	l := &LazyList{
		iter:       iter,
		state:      state,
		valueReady: false,
	}
	if hashable {
		lazyHashCount++
		l.id = lazyHashCount
	}
	return l
}

// String() should evaluate the whole list
func (l *LazyList) String() string {
	if l.Empty() {
		return "{Lazy: }"
	}
	h, _ := l.Head()
	return fmt.Sprintf("{Lazy: %v}", h)
}

func (l *LazyList) Hash() (string, error) {
	if l.Empty() {
		return l.String(), nil
	}
	if l.id > 0 {
		return fmt.Sprintf("{Lazy[%d]}", l.id), nil
	}
	return "", fmt.Errorf("Hash is not applicable to non-empty lazy lists")
}

func (l *LazyList) Print(w io.Writer) {
	var ll List = l
	io.WriteString(w, "'(")
	first := true
	for !ll.Empty() {
		if !first {
			io.WriteString(w, " ")
		} else {
			first = false
		}

		val, err := ll.Head()
		if err != nil {
			panic(fmt.Errorf("Head() failed: %v", err))
		}
		val.Print(w)
		ll, err = ll.Tail()
		if err != nil {
			panic(fmt.Errorf("Tail() failed: %v", err))
		}
	}
	io.WriteString(w, ")")
}

func (l *LazyList) Head() (Expr, error) {
	// iter: state -> '(value, new-state)
	// iter: value -> '(new-value)
	// iter: value -> new-value
	// iter: value -> '()  ; list finished
	if !l.valueReady {
		value, state, err := l.next()
		if err != nil {
			return nil, err
		}
		l.valueReady = true
		l.value = value
		l.state = state
	}
	if l.value == nil {
		return nil, fmt.Errorf("LazyList.Head(): list is empty")
	}
	return l.value, nil
}

func (l *LazyList) next() (value Expr, state Expr, err error) {
	args := []Expr{l.state}
	expr, err := l.iter.Eval(args)
	if err != nil {
		return nil, nil, fmt.Errorf("LazyList: Eval(%v) failed: %v", args, err)
	}
	res, ok := expr.(*Sexpr)
	if !ok {
		return expr, expr, nil
	}
	if len(res.List) == 0 {
		// list is finished
		return nil, nil, nil
	}
	if len(res.List) == 1 {
		// state = value
		return res.List[0], res.List[0], nil
	}
	if len(res.List) != 2 {
		return nil, nil, fmt.Errorf("Iterator result is too long: %v", res)
	}
	return res.List[0], res.List[1], nil
}

func (l *LazyList) Tail() (List, error) {
	if !l.valueReady {
		value, state, err := l.next()
		if err != nil {
			return nil, err
		}
		l.valueReady = true
		l.value = value
		l.state = state
	}
	if l.value == nil {
		return nil, fmt.Errorf("LazyList.Tail(): list is empty")
	}
	if l.tail == nil {
		l.tail = NewLazyList(l.iter, l.state, l.id > 0)
	}
	return l.tail, nil
}

func (l *LazyList) Empty() bool {
	if !l.valueReady {
		value, state, err := l.next()
		if err != nil {
			panic(err)
		}
		l.valueReady = true
		l.value = value
		l.state = state
	}
	return l.value == nil
}
