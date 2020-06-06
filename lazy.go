package main

import (
	"fmt"
	"io"

	"github.com/avoronkov/spil/types"
)

var _ types.List = (*LazyList)(nil)

type LazyList struct {
	iter       types.Function
	state      []types.Value
	value      *types.Value
	valueReady bool
	tail       *LazyList
	id         int64
}

var lazyHashCount int64

func NewLazyList(iter types.Function, state []types.Value, hashable bool) *LazyList {
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
	var ll types.List = l
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
		val.E.Print(w)
		ll, err = ll.Tail()
		if err != nil {
			panic(fmt.Errorf("Tail() failed: %v", err))
		}
	}
	io.WriteString(w, ")")
}

func (l *LazyList) Head() (*types.Value, error) {
	// iter: state -> '(value, new-state)
	// iter: value -> '(new-value)
	// iter: value -> new-value
	// iter: value -> '()  ; list finished
	if !l.valueReady {
		err := l.next()
		if err != nil {
			return nil, err
		}
	}
	if l.value == nil {
		return nil, fmt.Errorf("LazyList.Head(): list is empty")
	}
	return l.value, nil
}

func (l *LazyList) next() (err error) {
	expr, err := l.iter.Eval(l.state)
	if err != nil {
		return fmt.Errorf("LazyList: Eval(%v) failed: %v", l.state, err)
	}
	res, ok := expr.E.(*types.Sexpr)
	if !ok {
		l.valueReady = true
		l.value = expr
		l.state = []types.Value{*expr}
		return nil
	}
	if len(res.List) == 0 {
		// list is finished
		l.valueReady = true
		l.value = nil
		l.state = nil
		return nil
	}
	if len(res.List) == 1 {
		// state = value
		l.valueReady = true
		l.value = &res.List[0]
		l.state = []types.Value{res.List[0]}
		return nil
	}
	p1 := res.List[0]
	tail, err := res.Tail()
	if err != nil {
		return err
	}
	var newState []types.Value
	for !tail.Empty() {
		p, err := tail.Head()
		if err != nil {
			return err
		}
		newState = append(newState, *p)
		tail, err = tail.Tail()
		if err != nil {
			return err
		}
	}
	l.valueReady = true
	l.value = &p1
	l.state = newState
	return nil
}

func (l *LazyList) Tail() (types.List, error) {
	if !l.valueReady {
		err := l.next()
		if err != nil {
			return nil, err
		}
	}
	if l.value == nil {
		return nil, fmt.Errorf("LazyList.Tail(): list is empty")
	}
	if l.tail == nil {
		l.tail = NewLazyList(l.iter, l.state, l.id > 0)
	}
	return l.tail, nil
}

func (l *LazyList) Empty() (result bool) {
	if !l.valueReady {
		err := l.next()
		if err != nil {
			panic(err)
		}
	}
	return l.value == nil
}

func (l *LazyList) Type() types.Type {
	return types.TypeList
}
