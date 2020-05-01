package main

import (
	"fmt"
	"regexp"
)

// User-defined functions

type FuncInterpret struct {
	interpret *Interpret
	name      string
	bodies    []FuncImpl
}

type FuncImpl struct {
	argfmt Expr
	body   []Expr
}

func NewFuncInterpret(i *Interpret, name string) *FuncInterpret {
	return &FuncInterpret{
		interpret: i,
		name:      name,
	}
}

func (f *FuncInterpret) AddImpl(argfmt Expr, body []Expr) error {
	switch argfmt.(type) {
	case Ident:
		// pass arguments as list with specified name
		f.bodies = append(f.bodies, FuncImpl{argfmt, body})
	case *Sexpr:
		// bind arguments
		f.bodies = append(f.bodies, FuncImpl{argfmt, body})
	default:
		return fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}
	return nil
}

func (f *FuncInterpret) Eval(args []Expr) (Expr, error) {
	run := NewFuncRuntime(f)
	body, err := run.bind(args)
	if err != nil {
		return nil, err
	}
	return run.Eval(body)
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

func (f *FuncRuntime) bind(args []Expr) ([]Expr, error) {
	f.vars = make(map[string]Expr)
	var (
		argfmt Expr
		body   []Expr
	)
	argfmtFound := false
	for _, impl := range f.fi.bodies {
		if matchArgs(impl.argfmt, args) {
			argfmt = impl.argfmt
			body = impl.body
			argfmtFound = true
			break
		}
	}
	if !argfmtFound {
		return nil, fmt.Errorf("No matching function implementation for %v found", f.fi.name)
	}
	if argfmt != nil {
		switch a := argfmt.(type) {
		case Ident:
			f.vars[string(a)] = &Sexpr{List: args, Quoted: true}
		case *Sexpr:
			if l := a.Len(); l != len(args) {
				return nil, fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.fi.name, l, len(args))
			}
			for i, ident := range a.List {
				if iname, ok := ident.(Ident); ok {
					f.vars[string(iname)] = args[i]
				}
			}
		}
	}
	// bind to __args and _1, _2 ... variables
	f.vars["__args"] = &Sexpr{List: args, Quoted: true}
	for i, arg := range args {
		f.vars[fmt.Sprintf("_%d", i+1)] = arg
	}
	return body, nil
}

func (f *FuncRuntime) Eval(body []Expr) (res Expr, err error) {
L:
	for {
		last := len(body) - 1
		// log.Printf("Function %q: eval %v over %+v", f.fi.name, body, f.vars)
		for i, expr := range body {
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
				head, _ := lst.Head()
				hident, ok := head.(Ident)
				if !ok || string(hident) != f.fi.name {
					return f.evalFunc(lst)
				}
				// Tail call!
				t, _ := lst.Tail()
				tail := t.(*Sexpr)
				// eval args
				args := make([]Expr, 0, len(tail.List))
				for _, ar := range tail.List {
					arg, err := f.evalExpr(ar)
					if err != nil {
						return nil, err
					}
					args = append(args, arg)
				}
				body, err = f.bind(args)
				if err != nil {
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
			return nil, fmt.Errorf("%v: Unexpected empty s-expression: %v", f.fi.name, a)
		}
		head, _ := a.Head()
		if name, ok := head.(Ident); ok {
			if name == "if" {
				// (cond) (expr-if-true) (expr-if-false)
				if len(a.List) != 4 {
					return nil, fmt.Errorf("Expected 3 arguments to if, found: %v", a.List[1:])
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
				tail, _ := a.Tail()
				if err := f.setVar(tail.(*Sexpr)); err != nil {
					return nil, err
				}
				return QEmpty, nil
			}
			if name == "gen" {
				tail, _ := a.Tail()
				return f.evalGen(tail.(*Sexpr))
			}
			if name == "lambda" {
				tail, _ := a.Tail()
				return f.evalLambda(tail.(*Sexpr))
			}
		}

		// return unevaluated list
		return a, nil
	}
	panic(fmt.Errorf("Unexpected Expr type: %v (%T)", e, e))
}
func (f *FuncRuntime) evalExpr(expr Expr) (Expr, error) {
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
	return f.evalFunc(lst)
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

// (iter) (init-state)
func (f *FuncRuntime) evalGen(se *Sexpr) (Expr, error) {
	if se.Len() != 2 {
		return nil, fmt.Errorf("gen wants 2 argument, found %v", se.Repr())
	}
	fident, ok := se.List[0].(Ident)
	if !ok {
		return nil, fmt.Errorf("gen expects first argument to be a funtion, found: %v", se.List[0])
	}
	fu, err := f.findFunc(string(fident))
	if err != nil {
		return nil, err
	}
	state, err := f.evalExpr(se.List[1])
	if err != nil {
		return nil, err
	}
	return NewLazyList(fu, state), nil
}

func (f *FuncRuntime) findFunc(fname string) (Evaler, error) {
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
		return nil, fmt.Errorf("%v: Unknown function: %v", f.fi.name, fname)
	}
	return fu, nil
}

// (func-name) (args...)
func (f *FuncRuntime) evalFunc(se *Sexpr) (Expr, error) {
	head, err := se.Head()
	if err != nil {
		return nil, err
	}
	name, ok := head.(Ident)
	if !ok {
		return nil, fmt.Errorf("Wanted identifier, found: %v", head)
	}
	fname := string(name)
	fu, err := f.findFunc(fname)
	if err != nil {
		return nil, err
	}

	// evaluate arguments
	t, _ := se.Tail()
	tail := t.(*Sexpr)
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

var lambdaCount = 0

func (f *FuncRuntime) evalLambda(se *Sexpr) (Expr, error) {
	name := fmt.Sprintf("__lambda__%03d", lambdaCount)
	lambdaCount++
	fi := NewFuncInterpret(f.fi.interpret, name)
	body := f.replaceVars(se.List)
	fi.AddImpl(QList(Ident("__args")), body)
	f.fi.interpret.funcs[name] = fi
	return Ident(name), nil
}

var lambdaArgRe = regexp.MustCompile(`^(_[0-9]+|__args)$`)

func (f *FuncRuntime) replaceVars(st []Expr) (res []Expr) {
	for _, s := range st {
		switch a := s.(type) {
		case *Sexpr:
			v := &Sexpr{Quoted: a.Quoted}
			v.List = f.replaceVars(a.List)
			res = append(res, v)
		case Ident:
			if lambdaArgRe.MatchString(string(a)) {
				res = append(res, a)
			} else if v, ok := f.vars[string(a)]; ok {
				res = append(res, v)
			} else {
				res = append(res, a)
			}
		default:
			res = append(res, s)
		}
	}
	return res
}

func matchArgs(argfmt Expr, args []Expr) (result bool) {
	if argfmt == nil {
		// null matches everything (lambda case)
		return true
	}
	binds := map[string]Expr{}
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
				if !ok || at != v {
					return false
				}
			case Str:
				v, ok := args[i].(Str)
				if !ok || at != v {
					return false
				}
			case Bool:
				v, ok := args[i].(Bool)
				if !ok || at != v {
					return false
				}
			case *Sexpr:
				if at.Empty() {
					// special case to match empty lazy list
					if v, ok := args[i].(List); !ok || !v.Empty() {
						return false
					}
				} else {
					v, ok := args[i].(*Sexpr)
					if !ok || at.Repr() != v.Repr() {
						return false
					}
				}
			case Ident:
				// check if param is already binded
				if val, ok := binds[string(at)]; ok {
					if val.Repr() != args[i].Repr() {
						return false
					}
				}
				binds[string(at)] = args[i]
			default:
				panic(fmt.Errorf("Unexpected expr: %v (%T)", t, t))
			}
		}
		return true
	case Ident:
		// Ident matches everything
		return true
	}
	panic(fmt.Errorf("Unexpected argument format type: %v (%T)", argfmt, argfmt))
}
