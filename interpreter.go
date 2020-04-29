package main

import (
	"fmt"
	"io"
)

type Interpret struct {
	parser *Parser
}

func NewInterpreter(r io.Reader) *Interpret {
	return &Interpret{
		parser: NewParser(r),
	}
}

func (i *Interpret) Run() error {
	for {
		expr, err := i.parser.NextExpr()
		if err == io.EOF {
			break
		}
		_, err = i.evalExpr(expr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpret) evalExpr(e Expr) (Expr, error) {
	switch a := e.(type) {
	case Int:
		return a, nil
	case Str:
		return a, nil
	case Bool:
		return a, nil
	case Ident:
		return nil, fmt.Errorf("Unexpected identifier: %v", a)
	case *Sexpr:
		if a.Quoted {
			return a, nil
		}
		if a.Len() == 0 {
			return nil, fmt.Errorf("Unexpected empty s-expression: %v", a)
		}
		return i.evalFunc(a)
	}
	panic(fmt.Errorf("Unexpected Expr type: %v (%T)", e, e))
}

func (i *Interpret) evalFunc(se *Sexpr) (Expr, error) {
	head := se.Head()
	name, ok := head.(Ident)
	if !ok {
		return nil, fmt.Errorf("Wanted identifier, found: %v", head)
	}
	fn, ok := Funcs[string(name)]
	if !ok {
		return nil, fmt.Errorf("Unknown function: %v", name)
	}
	// evaluate arguments
	tail := se.Tail()
	args := make([]Expr, 0, len(tail.List))
	for _, arg := range tail.List {
		res, err := i.evalExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, res)
	}
	result, err := fn(args)
	return result, err
}
