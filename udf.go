package main

import (
	"fmt"
)

// User-defined functions

type FuncInterpret struct {
	interpret *Interpret
	name      string
	argfmt    Expr
	body      []Expr
}

func NewFuncInterpret(i *Interpret, name string, argfmt Expr, body []Expr) (*FuncInterpret, error) {
	fi := &FuncInterpret{
		interpret: i,
		name:      name,
		body:      body,
	}
	switch a := argfmt.(type) {
	case Ident:
		// pass arguments as list with specified name
		fi.argfmt = a
	case *Sexpr:
		// bind arguments
		for _, arg := range a.List {
			if _, ok := arg.(Ident); !ok {
				return nil, fmt.Errorf("argument name should be identifier, found %v", arg.Repr())
			}
		}
		fi.argfmt = a
	default:
		return nil, fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}
	return fi, nil
}

func (f *FuncInterpret) Eval(args []Expr) (Expr, error) {
	run, err := f.bind(args)
	if err != nil {
		return nil, err
	}
	return run.Eval()
}

// Evaluate function with specified arguments
func (f *FuncInterpret) bind(args []Expr) (*FuncRuntime, error) {
	fr := &FuncRuntime{
		fi:   f,
		vars: make(map[string]Expr),
	}
	switch a := f.argfmt.(type) {
	case Ident:
		fr.vars[string(a)] = &Sexpr{List: args, Quoted: true}
	case *Sexpr:
		if l := a.Len(); l != len(args) {
			return nil, fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.name, l, len(args))
		}
		for i, ident := range a.List {
			name := string(ident.(Ident))
			fr.vars[name] = args[i]
		}
	}
	return fr, nil
}

type FuncRuntime struct {
	fi   *FuncInterpret
	vars map[string]Expr
}

func (f *FuncRuntime) Eval() (res Expr, err error) {
	for _, expr := range f.fi.body {
		res, err = f.evalExpr(expr)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (f *FuncRuntime) evalExpr(e Expr) (Expr, error) {
	switch a := e.(type) {
	case Int:
		return a, nil
	case Str:
		return a, nil
	case Bool:
		return a, nil
	case Ident:
		if value, ok := f.vars[string(a)]; ok {
			return value, nil
		}
		return a, nil
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
				return f.evalIf(a.Tail())
			}
			if name == "set" {
				if err := f.setVar(a.Tail()); err != nil {
					return nil, err
				}
				return &Sexpr{Quoted: true}, nil
			}
		}

		return f.evalFunc(a)
	}
	panic(fmt.Errorf("Unexpected Expr type: %v (%T)", e, e))
}

// (cond) (expr-if-true) (expr-if-false)
func (f *FuncRuntime) evalIf(se *Sexpr) (Expr, error) {
	if len(se.List) != 3 {
		return nil, fmt.Errorf("Expected 3 arguments to if, found: %v", se)
	}
	arg := se.List[0]
	res, err := f.evalExpr(arg)
	if err != nil {
		return nil, err
	}
	boolRes, ok := res.(Bool)
	if !ok {
		return nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
	}
	if bool(boolRes) {
		return f.evalExpr(se.List[1])
	} else {
		return f.evalExpr(se.List[2])
	}
}

// (var-name) (value)
func (f *FuncRuntime) setVar(se *Sexpr) error {
	if se.Len() != 2 {
		return fmt.Errorf("set wants 2 argument, found %v", se.Repr())
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0].Repr())
	}
	value, err := f.evalExpr(se.List[1])
	if err != nil {
		return err
	}
	f.vars[string(name)] = value
	return nil
}

// (func-name) (args...)
func (f *FuncRuntime) evalFunc(se *Sexpr) (Expr, error) {
	head := se.Head()
	name, ok := head.(Ident)
	if !ok {
		return nil, fmt.Errorf("Wanted identifier, found: %v", head)
	}
	fname := string(name)
	// Ability to pass function name as argument
	if v, ok := f.vars[fname]; ok {
		vident, ok := v.(Ident)
		if !ok {
			return nil, fmt.Errorf("Cannot use argument %v as function", v.Repr())
		}
		fname = string(vident)
	}
	fu, ok := f.fi.interpret.funcs[fname]
	if !ok {
		return nil, fmt.Errorf("%v: Unknown function: %v", f.fi.name, name)
	}

	// evaluate arguments
	tail := se.Tail()
	args := make([]Expr, 0, len(tail.List))
	for _, arg := range tail.List {
		res, err := f.evalExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, res)
	}
	result, err := fu.Eval(args)
	return result, err
}
