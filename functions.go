package main

import (
	"fmt"
	"os"
	"unicode"
)

type Evaler interface {
	Eval([]Param) (*Param, error)
	ReturnType() Type
	TryBind(params []Param) (int, error)
}

type nativeFunc struct {
	name   string
	fn     func([]Param) (*Param, error)
	ret    Type
	binder func([]Param) error
}

func (n *nativeFunc) Eval(args []Param) (*Param, error) {
	return n.fn(args)
}

func (n *nativeFunc) ReturnType() Type {
	return n.ret
}

func (n *nativeFunc) TryBind(params []Param) (int, error) {
	if err := n.binder(params); err != nil {
		return -1, fmt.Errorf("%v: %v", n.name, err)
	}
	return 0, nil
}

func EvalerFunc(name string, fn func([]Param) (*Param, error), binder func([]Param) error, ret Type) Evaler {
	return &nativeFunc{
		name:   name,
		fn:     fn,
		ret:    ret,
		binder: binder,
	}
}

func FPlus(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FPlus: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Plus(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMinus(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMinus: expected integer argument in position %v, found %v", i, arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Minus(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMultiply(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Mult(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FDiv(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Div(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMod(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMod: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: second argument should be integer, found %v", args[1])
	}
	return &Param{V: a.Mod(b), T: TypeInt}, nil
}

func FLess(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1])
	}
	return &Param{V: Bool(a.Less(b)), T: TypeBool}, nil
}

func FLessEq(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1])
	}
	return &Param{V: Bool(!b.Less(a)), T: TypeBool}, nil
}

func FMore(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMore: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: second argument should be integer, found %v", args[1])
	}
	return &Param{V: Bool(b.Less(a)), T: TypeBool}, nil
}

func FMoreEq(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMore: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: second argument should be integer, found %v", args[1])
	}
	return &Param{V: Bool(!a.Less(b)), T: TypeBool}, nil
}

func FEq(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FEq: expected 2 arguments, found %v", args)
	}
	return &Param{V: Bool(Equal(args[0].V, args[1].V)), T: TypeBool}, nil
	/*
		switch a := args[0].V.(type) {
		case Int:
			b, ok := args[1].V.(Int)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be Int, found %v", args[1])
			}
			return Bool(a.Eq(b)), nil
		case Str:
			b, ok := args[1].V.(Str)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be Str, found %v", args[1])
			}
			return Bool(a == b), nil
		case Ident:
			b, ok := args[1].V.(Ident)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be Ident, found %v", args[1])
			}
			return Bool(a == b), nil
		case Bool:
			b, ok := args[1].V.(Bool)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be Bool, found %v", args[1])
			}
			return Bool(a == b), nil
		case *Sexpr:
			b, ok := args[1].V.(*Sexpr)
			if ok {
				if len(a.List) != len(b.List) {
					return Bool(false), nil
				}
				for i, first := range a.List {
					cmp, _ := FEq([]Param{first, b.List[i]})
					if cmp != Bool(true) {
						return Bool(false), nil
					}
				}
				return Bool(true), nil
			}
			if !a.Empty() {
				return nil, fmt.Errorf("FEq: Expected second argument to be List, found %v", args[1])
			}
			l, ok := args[1].V.(*LazyList)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be List or Lazy List, found %v", args[1])
			}
			return Bool(l.Empty()), nil
		case *LazyList:
			// Lazy list can be compared only to '()
			b, ok := args[1].V.(*Sexpr)
			if !ok {
				return nil, fmt.Errorf("FEq: Expected second argument to be '(), found %v", args[1])
			}
			if !b.Empty() {
				return nil, fmt.Errorf("FEq: Cannot compare lazy list with non-empty list: %v", args[1])
			}
			return Bool(a.Empty()), nil
		}
		panic(fmt.Errorf("Unknown argument type: %v (%T)", args[0], args[0]))
	*/
}

func FNot(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FNot: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(Bool)
	if !ok {
		return nil, fmt.Errorf("FNot: expected argument to be Bool, found %v", args[0])
	}
	return &Param{V: Bool(!bool(a)), T: TypeBool}, nil
}

func FHead(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FHead: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FHead: expected argument to be List, found %v", args[0])
	}
	return a.Head()
}

func FTail(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FTail: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FTail: expected argument to be List, found %v", args[0])
	}
	t, err := a.Tail()
	if err != nil {
		return nil, err
	}
	return &Param{V: t, T: TypeList}, nil
}

func FEmpty(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEmpty: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FEmpty: expected argument to be List, found %v", args[0])
	}
	return &Param{V: Bool(a.Empty()), T: TypeBool}, nil
}

type Appender interface {
	Append([]Param) (*Param, error)
}

func FAppend(args []Param) (*Param, error) {
	if len(args) == 0 {
		return &Param{V: QEmpty, T: TypeList}, nil
	}
	if len(args) == 1 {
		return &args[0], nil
	}
	a, ok := args[0].V.(Appender)
	if !ok {
		return nil, fmt.Errorf("FAppend: expected first argument to be Appender, found %v", args[0])
	}
	return a.Append(args[1:])
}

func FList(args []Param) (*Param, error) {
	s := new(Sexpr)
	for _, a := range args {
		s.List = append(s.List, a.V)
	}
	s.Quoted = true
	return &Param{V: s, T: TypeList}, nil
}

// test if symbol is white-space
func FSpace(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FSpace: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FSpace: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FSpace: expected argument to be Str of length 1, found %v", s)
	}
	return &Param{V: Bool(unicode.IsSpace(rune(string(s)[0]))), T: TypeBool}, nil
}

// test if symbol is eol (\n)
func FEol(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEol: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FEol: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FEol: expected argument to be Str of length 1, found %v", s)
	}
	return &Param{V: Bool(s == "\n"), T: TypeBool}, nil
}

func FOpen(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FOpen: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FOpen: expected argument to be Str, found %v", args)
	}
	file, err := os.Open(string(s))
	if err != nil {
		return nil, err
	}
	return &Param{V: NewLazyInput(file), T: TypeStr}, nil
}

// Binders
func AllInts(params []Param) error {
	for i, p := range params {
		if p.T == TypeUnknown {
			continue
		}
		if p.T != TypeInt {
			return fmt.Errorf("Expected all integer arguments, found%v at position %v", p, i)
		}
	}
	return nil
}

func TwoInts(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	if params[0].T != TypeInt && params[0].T != TypeUnknown {
		return fmt.Errorf("first argument should be integer, found %v", params[0])
	}
	if params[1].T != TypeInt && params[1].T != TypeUnknown {
		return fmt.Errorf("second argument should be integer, found %v", params[1])
	}
	return nil
}

func TwoArgs(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	return nil
}

func OneBoolArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}
	if params[0].T != TypeBool && params[0].T != TypeUnknown {
		return fmt.Errorf("expected argument to be Bool, found %v", params[0])
	}
	return nil
}

func AnyArgs(params []Param) error {
	return nil
}

func ListArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}
	if params[0].T != TypeList && params[0].T != TypeStr && params[0].T != TypeUnknown {
		return fmt.Errorf("expected argument to be List, found %v", params[0])
	}
	return nil
}

func AppenderArgs(params []Param) error {
	if len(params) <= 1 {
		return nil
	}
	if params[0].T != TypeList && params[0].T != TypeStr && params[0].T != TypeUnknown {
		return fmt.Errorf("FAppend: expected first argument to be Appender, found %v", params[0])
	}
	return nil
}

func StrArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", params)
	}
	if params[0].T != TypeStr && params[0].T != TypeUnknown {
		return fmt.Errorf("expected argument to be Str, found %v", params)
	}
	return nil
}
