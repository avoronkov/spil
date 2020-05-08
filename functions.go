package main

import (
	"fmt"
	"os"
	"unicode"
)

type Evaler interface {
	Eval([]Expr) (Expr, error)
	ReturnType() Type
	TryBind(args []Expr) error
}

type nativeFunc struct {
	name   string
	fn     func([]Expr) (Expr, error)
	ret    Type
	binder func([]Expr) error
}

func (n *nativeFunc) Eval(args []Expr) (Expr, error) {
	return n.fn(args)
}

func (n *nativeFunc) ReturnType() Type {
	return n.ret
}

func (n *nativeFunc) TryBind(args []Expr) error {
	err := n.binder(args)
	if err != nil {
		return fmt.Errorf("%v: %w", n.name, err)
	}
	return nil
}

func EvalerFunc(name string, fn func([]Expr) (Expr, error), binder func([]Expr) error, ret Type) Evaler {
	return &nativeFunc{
		name:   name,
		fn:     fn,
		ret:    ret,
		binder: binder,
	}
}

func FPlus(args []Expr) (Expr, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FPlus: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Plus(a)
		}
	}
	return result, nil
}

func FMinus(args []Expr) (Expr, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FMinus: expected integer argument in position %v, found %v", i, arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Minus(a)
		}
	}
	return result, nil
}

func FMultiply(args []Expr) (Expr, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Mult(a)
		}
	}
	return result, nil
}

func FDiv(args []Expr) (Expr, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Div(a)
		}
	}
	return result, nil
}
func FMod(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMod: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: second argument should be integer, found %v", args[1])
	}
	return a.Mod(b), nil
}

func FLess(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1])
	}
	return Bool(a.Less(b)), nil
}

func FLessEq(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1])
	}
	return Bool(!b.Less(a)), nil
}

func FMore(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMore: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: second argument should be integer, found %v", args[1])
	}
	return Bool(b.Less(a)), nil
}

func FMoreEq(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMore: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: second argument should be integer, found %v", args[1])
	}
	return Bool(!a.Less(b)), nil
}

func FEq(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FEq: expected 2 arguments, found %v", args)
	}
	switch a := args[0].(type) {
	case Int:
		b, ok := args[1].(Int)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Int, found %v", args[1])
		}
		return Bool(a.Eq(b)), nil
	case Str:
		b, ok := args[1].(Str)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Str, found %v", args[1])
		}
		return Bool(a == b), nil
	case Ident:
		b, ok := args[1].(Ident)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Ident, found %v", args[1])
		}
		return Bool(a == b), nil
	case Bool:
		b, ok := args[1].(Bool)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Bool, found %v", args[1])
		}
		return Bool(a == b), nil
	case *Sexpr:
		b, ok := args[1].(*Sexpr)
		if ok {
			if len(a.List) != len(b.List) {
				return Bool(false), nil
			}
			for i, first := range a.List {
				cmp, _ := FEq([]Expr{first, b.List[i]})
				if cmp != Bool(true) {
					return Bool(false), nil
				}
			}
			return Bool(true), nil
		}
		if !a.Empty() {
			return nil, fmt.Errorf("FEq: Expected second argument to be List, found %v", args[1])
		}
		l, ok := args[1].(*LazyList)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be List or Lazy List, found %v", args[1])
		}
		return Bool(l.Empty()), nil
	case *LazyList:
		// Lazy list can be compared only to '()
		b, ok := args[1].(*Sexpr)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be '(), found %v", args[1])
		}
		if !b.Empty() {
			return nil, fmt.Errorf("FEq: Cannot compare lazy list with non-empty list: %v", args[1])
		}
		return Bool(a.Empty()), nil
	}
	panic(fmt.Errorf("Unknown argument type: %v (%T)", args[0], args[0]))
}

func FNot(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FNot: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(Bool)
	if !ok {
		return nil, fmt.Errorf("FNot: expected argument to be Bool, found %v", args[0])
	}
	return Bool(!bool(a)), nil
}

func FHead(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FHead: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FHead: expected argument to be List, found %v", args[0])
	}
	return a.Head()
}

func FTail(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FTail: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FTail: expected argument to be List, found %v", args[0])
	}
	return a.Tail()
}

func FEmpty(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEmpty: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FEmpty: expected argument to be List, found %v", args[0])
	}
	return Bool(a.Empty()), nil
}

type Appender interface {
	Append([]Expr) (Expr, error)
}

func FAppend(args []Expr) (Expr, error) {
	if len(args) == 0 {
		return QEmpty, nil
	}
	if len(args) == 1 {
		return args[0], nil
	}
	a, ok := args[0].(Appender)
	if !ok {
		return nil, fmt.Errorf("FAppend: expected first argument to be Appender, found %v", args[0])
	}
	return a.Append(args[1:])
}

func FList(args []Expr) (Expr, error) {
	return &Sexpr{
		List:   args,
		Quoted: true,
	}, nil
}

// test if symbol is white-space
func FSpace(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FSpace: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].(Str)
	if !ok {
		return nil, fmt.Errorf("FSpace: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FSpace: expected argument to be Str of length 1, found %v", s)
	}
	return Bool(unicode.IsSpace(rune(string(s)[0]))), nil
}

// test if symbol is eol (\n)
func FEol(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEol: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].(Str)
	if !ok {
		return nil, fmt.Errorf("FEol: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FEol: expected argument to be Str of length 1, found %v", s)
	}
	return Bool(s == "\n"), nil
}

func FOpen(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FOpen: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].(Str)
	if !ok {
		return nil, fmt.Errorf("FOpen: expected argument to be Str, found %v", args)
	}
	file, err := os.Open(string(s))
	if err != nil {
		return nil, err
	}
	return NewLazyInput(file), nil
}

// Binders
func AllInts(args []Expr) error {
	for i, arg := range args {
		if _, ok := arg.(Int); !ok {
			return fmt.Errorf("Expected all integer arguments, found%v at position %v", arg, i)
		}
	}
	return nil
}

func TwoInts(args []Expr) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", args)
	}
	_, ok := args[0].(Int)
	if !ok {
		return fmt.Errorf("first argument should be integer, found %v", args[0])
	}
	_, ok = args[1].(Int)
	if !ok {
		return fmt.Errorf("second argument should be integer, found %v", args[1])
	}
	return nil
}

func TwoArgs(args []Expr) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", args)
	}
	return nil
}

func OneBoolArg(args []Expr) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", args)
	}
	_, ok := args[0].(Bool)
	if !ok {
		return fmt.Errorf("expected argument to be Bool, found %v", args[0])
	}
	return nil
}

func AnyArgs(args []Expr) error {
	return nil
}

func ListArg(args []Expr) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", args)
	}
	_, ok := args[0].(List)
	if !ok {
		return fmt.Errorf("expected argument to be List, found %v", args[0])
	}
	return nil
}

func AppenderArgs(args []Expr) error {
	if len(args) <= 1 {
		return nil
	}
	_, ok := args[0].(Appender)
	if !ok {
		return fmt.Errorf("FAppend: expected first argument to be Appender, found %v", args[0])
	}
	return nil
}

func StrArg(args []Expr) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", args)
	}
	_, ok := args[0].(Str)
	if !ok {
		return fmt.Errorf("expected argument to be Str, found %v", args)
	}
	return nil
}
