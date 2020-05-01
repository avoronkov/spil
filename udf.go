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
	run := NewFuncRuntime(f)
	if err := run.bind(args); err != nil {
		return nil, err
	}
	return run.Eval()
}

type FuncRuntime struct {
	fi   *FuncInterpret
	vars map[string]Expr
}

func NewFuncRuntime(fi *FuncInterpret) *FuncRuntime {
	return &FuncRuntime{
		fi: fi,
	}
}

func (f *FuncRuntime) bind(args []Expr) error {
	f.vars = make(map[string]Expr)
	switch a := f.fi.argfmt.(type) {
	case Ident:
		f.vars[string(a)] = &Sexpr{List: args, Quoted: true}
	case *Sexpr:
		if l := a.Len(); l != len(args) {
			return fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.fi.name, l, len(args))
		}
		for i, ident := range a.List {
			name := string(ident.(Ident))
			f.vars[name] = args[i]
		}
	}
	return nil
}

func (f *FuncRuntime) Eval() (res Expr, err error) {
	last := len(f.fi.body) - 1
L:
	for {
		for i, expr := range f.fi.body {
			if i == last {
				// check for tail call
				e, err := f.lastExpr(expr)
				if err != nil {
					return nil, err
				}
				lst, ok := e.(*Sexpr)
				if !ok {
					// nothing to evaluate
					return e, nil
				}
				if lst.Quoted || lst.Len() == 0 {
					return lst, nil
				}
				head := lst.Head()
				hident, ok := head.(Ident)
				if !ok || string(hident) != f.fi.name {
					return f.evalFunc(lst)
				}
				// Tail call!
				tail := lst.Tail()
				// eval args
				args := make([]Expr, 0, len(tail.List))
				for _, ar := range tail.List {
					arg, err := f.evalExpr(ar)
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
				if err := f.bind(args); err != nil {
					return nil, err
				}
				continue L
			} else {
				res, err = f.evalExpr(expr)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return
}

func (f *FuncRuntime) lastExpr(e Expr) (Expr, error) {
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
				if len(a.List) != 4 {
					return nil, fmt.Errorf("Expected 3 arguments to if, found: %v", a.Tail())
				}
				arg := a.List[1]
				res, err := f.evalExpr(arg)
				if err != nil {
					return nil, err
				}
				boolRes, ok := res.(Bool)
				if !ok {
					return nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
				}
				if bool(boolRes) {
					return f.lastExpr(a.List[2])
				}
				return f.lastExpr(a.List[3])
			}
			if name == "set" {
				if err := f.setVar(a.Tail()); err != nil {
					return nil, err
				}
				return &Sexpr{Quoted: true}, nil
			}
		}

		// return unevaluated list
		return a, nil
	}
	panic(fmt.Errorf("Unexpected Expr type: %v (%T)", e, e))
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

func matchArgs(argfmt Expr, args []Expr) bool {
	switch a := argfmt.(type) {
	case *Sexpr:
		if len(a.List) == 0 && len(args) == 0 {
			return true
		}
		if len(a.List) != len(args) {
			return false
		}
		for i, t := range a.List {
			switch at := t.(type) {
			case Int:
				v, ok := args[i].(Int)
				return ok && at == v
			case Str:
				v, ok := args[i].(Str)
				return ok && at == v
			case Bool:
				v, ok := args[i].(Bool)
				return ok && at == v
			case *Sexpr:
				v, ok := args[i].(*Sexpr)
				return ok && at.Repr() == v.Repr()
			case Ident:
				// Ident matches everything
				return true
			}
			panic(fmt.Errorf("Unexpected expr: %v (%T)", t, t))
		}
	case Ident:
		// Ident matches everything
		return true
	}
	panic(fmt.Errorf("Unexpected argument format type: %v (%T)", argfmt, argfmt))
}
