package main

import (
	"fmt"
	"io"
)

type Interpret struct {
	parser   *Parser
	output   io.Writer
	vars     map[string]Expr
	deffuncs map[string]Func
	funcs    map[string]*FuncInterpet
}

func NewInterpreter(r io.Reader, w io.Writer) *Interpret {
	i := &Interpret{
		parser: NewParser(r),
		output: w,
		vars:   make(map[string]Expr),
		funcs:  make(map[string]*FuncInterpet),
	}
	i.deffuncs = map[string]Func{
		"+":     FPlus,
		"-":     FMinus,
		"*":     FMultiply,
		"<":     FLess,
		"=":     FEq,
		"not":   FNot,
		"print": i.FPrint,
		"head":  FHead,
		"tail":  FTail,
	}
	return i
}

func (i *Interpret) Run() error {
	for {
		expr, err := i.parser.NextExpr()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
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
		if value, ok := i.vars[string(a)]; ok {
			return value, nil
		}
		return nil, fmt.Errorf("Unknown identifier: %v", a)
	case *Sexpr:
		if a.Quoted {
			return a, nil
		}
		if a.Len() == 0 {
			return nil, fmt.Errorf("Unexpected empty s-expression: %v", a)
		}
		head := a.Head()
		if name, ok := head.(Ident); ok {
			if name == "if" {
				return i.evalIf(a.Tail())
			}
			if name == "set" {
				if err := i.setVar(a.Tail()); err != nil {
					return nil, err
				}
				return &Sexpr{Quoted: true}, nil
			}
			if name == "func" || name == "def" {
				if err := i.defineFunc(a.Tail()); err != nil {
					return nil, err
				}
				return &Sexpr{Quoted: true}, nil
			}
		}

		return i.evalFunc(a)
	}
	panic(fmt.Errorf("Unexpected Expr type: %v (%T)", e, e))
}

// (cond) (expr-if-true) (expr-if-false)
func (i *Interpret) evalIf(se *Sexpr) (Expr, error) {
	if len(se.List) != 3 {
		return nil, fmt.Errorf("Expected 3 arguments to if, found: %v", se)
	}
	arg := se.List[0]
	res, err := i.evalExpr(arg)
	if err != nil {
		return nil, err
	}
	boolRes, ok := res.(Bool)
	if !ok {
		return nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
	}
	if bool(boolRes) {
		return i.evalExpr(se.List[1])
	} else {
		return i.evalExpr(se.List[2])
	}
}

// (func-name) (args...)
func (i *Interpret) evalFunc(se *Sexpr) (Expr, error) {
	head := se.Head()
	name, ok := head.(Ident)
	if !ok {
		return nil, fmt.Errorf("Wanted identifier, found: %v", head)
	}
	if fu, ok := i.funcs[string(name)]; ok {
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
		fr, err := fu.Bind(args)
		if err != nil {
			return nil, err
		}
		return fr.Eval()
	}
	fn, ok := i.deffuncs[string(name)]
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

// (var-name) (value)
func (i *Interpret) setVar(se *Sexpr) error {
	if se.Len() != 2 {
		return fmt.Errorf("set wants 2 argument, found %v", se.Repr())
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0].Repr())
	}
	value, err := i.evalExpr(se.List[1])
	if err != nil {
		return err
	}
	i.vars[string(name)] = value
	return nil
}

// (func-name) args body...
func (i *Interpret) defineFunc(se *Sexpr) error {
	if se.Len() < 3 {
		return fmt.Errorf("Not enough arguments for function definition: %v", se.Repr())
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("func expected identifier first, found %v", se.List[0].Repr())
	}

	fi, err := NewFuncInterpret(i, string(name), se.List[1], se.List[2:])
	if err != nil {
		return err
	}
	i.funcs[string(name)] = fi
	return nil
}

func (in *Interpret) FPrint(args []Expr) (Expr, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		fmt.Fprintf(in.output, "%v", e.String())
	}
	fmt.Fprintf(in.output, "\n")
	return &Sexpr{Quoted: true}, nil
}
