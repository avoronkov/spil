package main

import (
	"fmt"
)

type Evaler interface {
	Eval([]Expr) (Expr, error)
}

type EvalerFunc func([]Expr) (Expr, error)

func (f EvalerFunc) Eval(args []Expr) (Expr, error) {
	return f(args)
}

func FPlus(args []Expr) (Expr, error) {
	var result int
	for _, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FPlus: expected integer argument, found %v", arg.Repr())
		}
		result += int(a)
	}
	return Int(result), nil
}

func FMinus(args []Expr) (Expr, error) {
	var result int
	for i, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FMinus: expected integer argument, found %v", arg.Repr())
		}
		if i == 0 {
			result = int(a)
		} else {
			result -= int(a)
		}
	}
	return Int(result), nil
}

func FMultiply(args []Expr) (Expr, error) {
	var result int = 1
	for _, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg.Repr())
		}
		result *= int(a)
	}
	return Int(result), nil
}

func FDiv(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FDiv: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0].Repr())
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1].Repr())
	}
	return Int(a / b), nil
}

func FLess(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: first argument should be integer, found %v", args[0].Repr())
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FLess: second argument should be integer, found %v", args[1].Repr())
	}
	return Bool(a < b), nil
}

func FMore(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMore: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: first argument should be integer, found %v", args[0].Repr())
	}
	b, ok := args[1].(Int)
	if !ok {
		return nil, fmt.Errorf("FMore: second argument should be integer, found %v", args[1].Repr())
	}
	return Bool(a > b), nil
}

func FEq(args []Expr) (Expr, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FEq: expected 2 arguments, found %v", args)
	}
	switch a := args[0].(type) {
	case Int:
		b, ok := args[1].(Int)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Int, found %v", args[1].Repr())
		}
		return Bool(a == b), nil
	case Str:
		b, ok := args[1].(Str)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Str, found %v", args[1].Repr())
		}
		return Bool(a == b), nil
	case Ident:
		b, ok := args[1].(Ident)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Ident, found %v", args[1].Repr())
		}
		return Bool(a == b), nil
	case Bool:
		b, ok := args[1].(Bool)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Bool, found %v", args[1].Repr())
		}
		return Bool(a == b), nil
	case *Sexpr:
		b, ok := args[1].(*Sexpr)
		if !ok {
			return nil, fmt.Errorf("FEq: Expected second argument to be Str, found %v", args[1].Repr())
		}
		return Bool(a.Repr() == b.Repr()), nil
	}
	panic(fmt.Errorf("Unknown argument type: %v (%T)", args[0], args[0]))
}

func FNot(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FNot: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(Bool)
	if !ok {
		return nil, fmt.Errorf("FNot: expected argument to be Bool, found %v", args[0].Repr())
	}
	return Bool(!bool(a)), nil
}

func FHead(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FHead: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FHead: expected argument to be List, found %v", args[0].Repr())
	}
	return a.Head()
}

func FTail(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FTail: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FTail: expected argument to be List, found %v", args[0].Repr())
	}
	return a.Tail()
}

func FEmpty(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEmpty: expected 1 argument, found %v", args)
	}
	a, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf("FEmpty: expected argument to be List, found %v", args[0].Repr())
	}
	return Bool(a.Empty()), nil
}

func FAppend(args []Expr) (Expr, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("FAppend: expected at least 2 arguments, found %v", args)
	}
	l, ok := args[0].(*Sexpr)
	if !ok {
		return nil, fmt.Errorf("FAppend: expected first argument to be List, found %v", args[0].Repr())
	}
	newList := make([]Expr, len(l.List))
	copy(newList, l.List)
	newList = append(newList, args[1:]...)
	return &Sexpr{
		List:   newList,
		Quoted: l.Quoted,
	}, nil
}
