package main

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

// User-defined functions

type FuncInterpret struct {
	interpret *Interpret
	name      string
	bodies    []*FuncImpl
}

type FuncImpl struct {
	argfmt Expr
	body   []Expr
	// Do we need to remenber function results?
	memo bool
	// Function results: args.Repr() -> Result
	results map[string]Expr
}

func NewFuncImpl(argfmt Expr, body []Expr, memo bool) *FuncImpl {
	i := &FuncImpl{
		argfmt: argfmt,
		body:   body,
		memo:   memo,
	}
	if memo {
		i.results = make(map[string]Expr)
	}
	return i
}

func (i *FuncImpl) RememberResult(name string, args []Expr, result Expr) {
	// log.Printf("%v: rememberRusult %v -> %v", name, args, result)
	keyArgs, err := keyOfArgs(args)
	if err != nil {
		log.Printf("Cannot rememer result for %v: %v", args, err)
		return
	}
	if _, ok := i.results[keyArgs]; ok {
		panic(fmt.Errorf("%v: already have saved result for arguments %v (%v)", name, args, keyArgs))
	}
	i.results[keyArgs] = result
}

func NewFuncInterpret(i *Interpret, name string) *FuncInterpret {
	return &FuncInterpret{
		interpret: i,
		name:      name,
	}
}

func (f *FuncInterpret) AddImpl(argfmt Expr, body []Expr, memo bool) error {
	if argfmt == nil {
		f.bodies = append(f.bodies, NewFuncImpl(nil, body, memo))
		return nil
	}
	switch argfmt.(type) {
	case Ident:
		// pass arguments as list with specified name
		f.bodies = append(f.bodies, NewFuncImpl(argfmt, body, memo))
	case *Sexpr:
		// bind arguments
		f.bodies = append(f.bodies, NewFuncImpl(argfmt, body, memo))
	default:
		return fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}
	return nil
}

func (f *FuncInterpret) Eval(args []Expr) (Expr, error) {
	// log.Printf("%v: Eval(%v)", f.name, args)
	run := NewFuncRuntime(f)
	impl, result, err := run.bind(args)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	return run.Eval(impl)
}

type FuncRuntime struct {
	fi   *FuncInterpret
	vars map[string]Expr
	args []Expr
}

func NewFuncRuntime(fi *FuncInterpret) *FuncRuntime {
	return &FuncRuntime{
		fi: fi,
	}
}

func keyOfArgs(args []Expr) (string, error) {
	b := &strings.Builder{}
	for _, arg := range args {
		hash, err := arg.Hash()
		if err != nil {
			return "", err
		}
		io.WriteString(b, hash+" ")
	}
	return b.String(), nil
}

func (f *FuncRuntime) bind(args []Expr) (impl *FuncImpl, result Expr, err error) {
	// log.Printf("%v: bind args %v", f.fi.name, args)
	f.vars = make(map[string]Expr)
	argfmtFound := false
	for idx, im := range f.fi.bodies {
		if matchArgs(im.argfmt, args) {
			impl = f.fi.bodies[idx]
			argfmtFound = true
			if im.memo {
				keyArgs, err := keyOfArgs(args)
				if err != nil {
					log.Printf("Cannot compute hash of args: %v, %v", args, err)
				} else if res, ok := im.results[keyArgs]; ok {
					// log.Printf("%v: bind returns result: %v -> %v", f.fi.name, args, res)
					return nil, res, nil
				}
			}
			break
		}
	}
	if !argfmtFound {
		err = fmt.Errorf("No matching function implementation for %v found", f.fi.name)
		return
	}
	// log.Printf("%v: bind impl.argfmt = %v", f.fi.name, impl.argfmt)
	if impl.argfmt != nil {
		switch a := impl.argfmt.(type) {
		case Ident:
			f.vars[string(a)] = &Sexpr{List: args, Quoted: true}
		case *Sexpr:
			if l := a.Len(); l != len(args) {
				err = fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.fi.name, l, len(args))
				return
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
	f.args = args
	return impl, nil, nil
}

func (f *FuncRuntime) Eval(impl *FuncImpl) (res Expr, err error) {
L:
	for {
		last := len(impl.body) - 1
		// // log.Printf("Function %q: eval %v over %+v", f.fi.name, body, f.vars)
		for i, expr := range impl.body {
			if i == last {
				// check for tail call
				e, err := f.lastExpr(expr)
				if err != nil {
					return nil, err
				}
				lst, ok := e.(*Sexpr)
				if !ok {
					if impl.memo {
						// lets remenber the result
						// log.Printf("%v: remember result 1", f.fi.name)
						impl.RememberResult(f.fi.name, f.args, e)
					}
					// nothing to evaluate
					return e, nil
				}
				if lst.Quoted || lst.Len() == 0 {
					if impl.memo {
						// lets remenber the result
						// log.Printf("%v: remember result 2", f.fi.name)
						impl.RememberResult(f.fi.name, f.args, lst)
					}
					return lst, nil
				}
				head, _ := lst.Head()
				hident, ok := head.(Ident)
				if !ok || (string(hident) != f.fi.name && string(hident) != "self") {
					result, err := f.evalFunc(lst)
					if err != nil {
						return nil, err
					}
					if impl.memo {
						// lets remenber the result
						// log.Printf("%v: remember result 3", f.fi.name)
						impl.RememberResult(f.fi.name, f.args, result)
					}
					return result, nil
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
				var result Expr
				impl, result, err = f.bind(args)
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
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
			if a.Lambda {
				return f.evalLambda(&Sexpr{List: []Expr{a}, Quoted: true})
			}
			if name == "lambda" {
				tail, _ := a.Tail()
				return f.evalLambda(tail.(*Sexpr))
			}
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
			if name == "do" {
				if len(a.List) < 2 {
					return nil, fmt.Errorf("do: empty body")
				}
				if len(a.List) > 2 {
					for _, st := range a.List[1 : len(a.List)-1] {
						if _, err := f.evalExpr(st); err != nil {
							return nil, err
						}
					}
				}
				return f.lastExpr(a.List[len(a.List)-1])
			}
			if name == "and" {
				for _, arg := range a.List[1:] {
					res, err := f.evalExpr(arg)
					if err != nil {
						return nil, err
					}
					boolRes, ok := res.(Bool)
					if !ok {
						return nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if !bool(boolRes) {
						return Bool(false), nil
					}
				}
				return Bool(true), nil
			}
			if name == "or" {
				for _, arg := range a.List[1:] {
					res, err := f.evalExpr(arg)
					if err != nil {
						return nil, err
					}
					boolRes, ok := res.(Bool)
					if !ok {
						return nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if bool(boolRes) {
						return Bool(true), nil
					}
				}
				return Bool(false), nil
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
			if name == "apply" {
				tail, _ := a.Tail()
				return f.evalApply(tail.(*Sexpr))
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
		return fmt.Errorf("set wants 2 argument, found %v", se)
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0])
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
		return nil, fmt.Errorf("gen wants 2 argument, found %v", se)
	}
	fn, err := f.evalExpr(se.List[0])
	if err != nil {
		return nil, err
	}
	fident, ok := fn.(Ident)
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
			return nil, fmt.Errorf("%v: cannot use argument %v as function", f.fi.name, v)
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
		return nil, fmt.Errorf("Wanted identifier, found: %v (%v)", head, se)
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
	fi.AddImpl(nil, body, false)
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

func (f *FuncRuntime) evalApply(se *Sexpr) (Expr, error) {
	if len(se.List) != 2 {
		return nil, fmt.Errorf("apply expects function with list of arguments")
	}
	res, err := f.evalExpr(se.List[1])
	if err != nil {
		return nil, err
	}
	args, ok := res.(List)
	if !ok {
		return nil, fmt.Errorf("apply expects result to be a list of argument")
	}
	cmd := []Expr{se.List[0]}
	for !args.Empty() {
		h, _ := args.Head()
		cmd = append(cmd, h)
		args, _ = args.Tail()
	}

	return &Sexpr{
		List: cmd,
	}, nil
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
				if !ok || !at.Eq(v) {
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
					if !ok {
						return false
					}
					h1, err := at.Hash()
					if err != nil {
						return false
					}
					h2, err := v.Hash()
					if err != nil {
						return false
					}
					if h1 != h2 {
						return false
					}
				}
			case Ident:
				// check if param is already binded
				if val, ok := binds[string(at)]; ok {
					if val.String() != args[i].String() {
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
