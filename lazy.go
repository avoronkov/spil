package main

import (
	"fmt"
)

var _ List = (*LazyList)(nil)

type LazyList struct {
	iter       Evaler
	state      Expr
	value      Expr
	valueReady bool
}

func NewLazyList(iter Evaler, state Expr) *LazyList {
	return &LazyList{
		iter:       iter,
		state:      state,
		valueReady: false,
	}
}

// String() should evaluate the whole list
func (l *LazyList) String() string {
	var ll List = l
	res := make([]Expr, 0, 64)
	for !ll.Empty() {
		val, err := ll.Head()
		if err != nil {
			panic(fmt.Errorf("Head() failed: %v", err))
		}
		res = append(res, val)
		ll, err = ll.Tail()
		if err != nil {
			panic(fmt.Errorf("Tail() failed: %v", err))
		}
	}
	se := &Sexpr{
		List:   res,
		Quoted: true,
	}
	return se.String()
}

func (l *LazyList) Repr() string {
	return "TODO"
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
		return nil, fmt.Errorf("List is empty")
	}
	return l.value, nil
}

func (l *LazyList) next() (value Expr, state Expr, err error) {
	args := []Expr{l.state}
	expr, err := l.iter.Eval(args)
	if err != nil {
		return nil, nil, err
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
		return nil, nil, fmt.Errorf("Iterator result is too long: %v", res.Repr())
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
		return nil, fmt.Errorf("List is empty")
	}
	return &LazyList{
		iter:       l.iter,
		state:      l.state,
		valueReady: false,
	}, nil
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
