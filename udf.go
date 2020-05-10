package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

// User-defined functions

type FuncInterpret struct {
	interpret  *Interpret
	name       string
	bodies     []*FuncImpl
	returnType Type
}

type FuncImpl struct {
	argfmt *ArgFmt
	body   []Expr
	// Do we need to remenber function results?
	memo bool
	// Function results: args.Repr() -> Result
	results map[string]*Param
}

func NewFuncImpl(argfmt *ArgFmt, body []Expr, memo bool) *FuncImpl {
	i := &FuncImpl{
		argfmt: argfmt,
		body:   body,
		memo:   memo,
	}
	if memo {
		i.results = make(map[string]*Param)
	}
	return i
}

func (i *FuncImpl) RememberResult(name string, args []Expr, result *Param) {
	keyArgs, err := keyOfArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: cannot rememer result for %v: %v\n", name, args, err)
		return
	}
	if _, ok := i.results[keyArgs]; ok {
		panic(fmt.Errorf("%v: already have saved result for arguments %v (%v)", name, args, keyArgs))
	}
	i.results[keyArgs] = result
}

func NewFuncInterpret(i *Interpret, name string) *FuncInterpret {
	return &FuncInterpret{
		interpret:  i,
		name:       name,
		returnType: TypeUnknown,
	}
}

func (f *FuncInterpret) AddImpl(argfmt Expr, body []Expr, memo bool, returnType Type) error {
	if len(f.bodies) > 0 && returnType != f.returnType {
		return fmt.Errorf("%v: cannot redefine return type: previous %v, current %v", f.name, f.returnType, returnType)
	}
	if argfmt == nil {
		f.bodies = append(f.bodies, NewFuncImpl(nil, body, memo))
		return nil
	}
	af, err := ParseArgFmt(argfmt)
	if err != nil {
		return err
	}
	f.bodies = append(f.bodies, NewFuncImpl(af, body, memo))

	f.returnType = returnType
	return nil
}

func (f *FuncInterpret) TryBind(params []Param) (int, error) {
	for idx, im := range f.bodies {
		if f.matchParameters(im.argfmt, params) {
			return idx, nil
		}
	}
	return -1, fmt.Errorf("%v: no matching function implementaion found for %v", f.name, params)
}

func (f *FuncInterpret) Eval(params []Param) (*Param, error) {
	fmt.Fprintf(os.Stderr, "%v: Eval (%v)\n", f.name, params)
	run := NewFuncRuntime(f)
	impl, result, err := run.bind(params)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	res, err := run.Eval(impl)
	if err != nil {
		return nil, err
	}
	run.cleanup()
	return res, err
}

func (f *FuncInterpret) ReturnType() Type {
	return f.returnType
}

type FuncRuntime struct {
	fi   *FuncInterpret
	vars map[string]Param
	args []Expr
	// variables that should be Closed after leaving this variable scope.
	scopedVars []string
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

func (f *FuncRuntime) bind(params []Param) (impl *FuncImpl, result *Param, err error) {
	f.cleanup()
	// params := MakeParametersFromArgs(args)
	args := make([]Expr, 0, len(params))
	for _, p := range params {
		args = append(args, p.V)
	}
	idx, err := f.fi.TryBind(params)
	if err != nil {
		return nil, nil, err
	}
	impl = f.fi.bodies[idx]
	if impl.memo {
		keyArgs, err := keyOfArgs(args)
		if err != nil {
			log.Printf("Cannot compute hash of args: %v, %v", args, err)
		} else if res, ok := impl.results[keyArgs]; ok {
			return nil, res, nil
		}
	}

	if impl.argfmt != nil {
		if impl.argfmt.Wildcard != "" {
			f.vars[impl.argfmt.Wildcard] = Param{
				V: &Sexpr{List: args, Quoted: true},
				T: TypeList,
			}
		} else {
			if l := len(impl.argfmt.Args); l != len(params) {
				err = fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.fi.name, l, len(params))
				return
			}
			for i, arg := range impl.argfmt.Args {
				if arg.V == nil {
					f.vars[arg.Name] = params[i]
				}
			}
		}
	}
	// bind to __args and _1, _2 ... variables
	f.vars["__args"] = Param{
		V: &Sexpr{List: args, Quoted: true},
		T: TypeList,
	}
	for i, arg := range params {
		f.vars[fmt.Sprintf("_%d", i+1)] = arg
	}
	f.args = args
	return impl, nil, nil
}

func (f *FuncRuntime) Eval(impl *FuncImpl) (res *Param, err error) {
	memoImpl := impl
	memoArgs := f.args
L:
	for {
		last := len(impl.body) - 1
		if last < 0 {
			break L
		}
		if id, ok := impl.body[last].(Ident); ok {
			if _, ok := ParseType(string(id)); ok {
				// Last statement is type declaration
				last--
			}
		}
		for i, expr := range impl.body {
			if i == last {
				// check for tail call
				e, _, err := f.lastParameter(expr)
				if err != nil {
					return nil, err
				}
				lst, ok := e.V.(*Sexpr)
				if !ok {
					if memoImpl.memo {
						// lets remenber the result
						memoImpl.RememberResult(f.fi.name, memoArgs, e)
					}
					// nothing to evaluate
					return e, nil
				}
				if lst.Quoted || lst.Len() == 0 {
					p := &Param{V: lst, T: TypeList}
					if memoImpl.memo {
						// lets remenber the result
						memoImpl.RememberResult(f.fi.name, memoArgs, p)
					}
					return p, nil
				}
				head, _ := lst.Head()
				hident, ok := head.V.(Ident)
				if !ok || (string(hident) != f.fi.name && string(hident) != "self") {
					result, err := f.evalFunc(lst)
					if err != nil {
						return nil, err
					}
					if memoImpl.memo {
						// lets remenber the result
						memoImpl.RememberResult(f.fi.name, f.args, result)
					}
					return result, nil
				}
				// Tail call!
				t, _ := lst.Tail()
				tail := t.(*Sexpr)
				// eval args
				args := make([]Param, 0, len(tail.List))
				for _, ar := range tail.List {
					arg, err := f.evalParameter(ar)
					if err != nil {
						return nil, err
					}
					args = append(args, *arg)
				}
				var result *Param
				impl, result, err = f.bind(args)
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
				}
				continue L
			} else {
				res, err = f.evalParameter(expr)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return
}

func (f *FuncRuntime) lastParameter(e Expr) (*Param, *Type, error) {
	switch a := e.(type) {
	case Int:
		return &Param{V: a, T: TypeInt}, nil, nil
	case Str:
		return &Param{V: a, T: TypeStr}, nil, nil
	case Bool:
		return &Param{V: a, T: TypeBool}, nil, nil
	case Ident:
		if value, ok := f.vars[string(a)]; ok {
			return &value, nil, nil
		}
		return &Param{V: a, T: TypeUnknown}, nil, nil
	case *Sexpr:
		if a.Quoted {
			return &Param{V: a, T: TypeList}, nil, nil
		}
		if a.Len() == 0 {
			return nil, nil, fmt.Errorf("%v: Unexpected empty s-expression: %v", f.fi.name, a)
		}
		head, _ := a.Head()
		if name, ok := head.V.(Ident); ok {
			if a.Lambda {
				lm, err := f.evalLambda(&Sexpr{List: []Expr{a}, Quoted: true})
				if err != nil {
					return nil, nil, err
				}
				return &Param{
					V: lm,
					T: TypeFunc,
				}, nil, nil
			}
			if name == "lambda" {
				tail, _ := a.Tail()
				lm, err := f.evalLambda(tail.(*Sexpr))
				if err != nil {
					return nil, nil, err
				}
				return &Param{
					V: lm,
					T: TypeFunc,
				}, nil, nil
			}
			if name == "if" {
				// (cond) (expr-if-true) (expr-if-false)
				if len(a.List) != 4 {
					return nil, nil, fmt.Errorf("Expected 3 arguments to if, found: %v", a.List[1:])
				}
				arg := a.List[1]
				res, err := f.evalExpr(arg)
				if err != nil {
					return nil, nil, err
				}
				boolRes, ok := res.(Bool)
				if !ok {
					return nil, nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
				}
				if bool(boolRes) {
					return f.lastParameter(a.List[2])
				}
				return f.lastParameter(a.List[3])
			}
			if name == "do" {
				// fmt.Fprintf(os.Stderr, "DO\n")
				var retType *Type
				last := len(a.List) - 1
				if id, ok := a.List[last].(Ident); ok {
					if rt, ok := ParseType(string(id)); ok {
						// Last statement is type declaration
						last--
						retType = &rt
					}
				}
				// fmt.Fprintf(os.Stderr, "DO: retType = %v\n", retType)
				if last == 0 {
					return nil, nil, fmt.Errorf("do: empty body")
				}
				if last > 1 {
					for _, st := range a.List[1:last] {
						if _, err := f.evalExpr(st); err != nil {
							return nil, nil, err
						}
					}
				}
				ret, ft, err := f.lastParameter(a.List[last])
				if err != nil {
					return nil, nil, err
				}
				if retType != nil {
					// TODO check matching types
					ret.T = *retType
					ft = retType
				}
				// fmt.Fprintf(os.Stderr, "DO return %v\n", ret)
				return ret, ft, nil
			}
			if name == "and" {
				for _, arg := range a.List[1:] {
					res, err := f.evalExpr(arg)
					if err != nil {
						return nil, nil, err
					}
					boolRes, ok := res.(Bool)
					if !ok {
						return nil, nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if !bool(boolRes) {
						return &Param{
							V: Bool(false),
							T: TypeBool,
						}, nil, nil
					}
				}
				return &Param{
					V: Bool(true),
					T: TypeBool,
				}, nil, nil
			}
			if name == "or" {
				for _, arg := range a.List[1:] {
					res, err := f.evalExpr(arg)
					if err != nil {
						return nil, nil, err
					}
					boolRes, ok := res.(Bool)
					if !ok {
						return nil, nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if bool(boolRes) {
						return &Param{
							V: Bool(true),
							T: TypeBool,
						}, nil, nil
					}
				}
				return &Param{
					V: Bool(false),
					T: TypeBool,
				}, nil, nil
			}
			if name == "set" || name == "set'" {
				tail, _ := a.Tail()
				if err := f.setVar(tail.(*Sexpr) /*scoped*/, name == "set'"); err != nil {
					return nil, nil, err
				}
				return &Param{V: QEmpty, T: TypeAny}, nil, nil
			}
			if name == "gen" || name == "gen'" {
				tail, _ := a.Tail()
				gen, err := f.evalGen(tail.(*Sexpr) /*hashable*/, name == "gen'")
				if err != nil {
					return nil, nil, err
				}
				return &Param{V: gen, T: TypeList}, nil, nil
			}
			if name == "apply" {
				tail, _ := a.Tail()
				res, err := f.evalApply(tail.(*Sexpr))
				if err != nil {
					return nil, nil, err
				}
				return &Param{V: res, T: TypeUnknown}, nil, nil
			}
		}

		// return unevaluated list
		return &Param{V: a, T: TypeUnknown}, nil, nil
	case *LazyList:
		return &Param{V: a, T: TypeList}, nil, nil
	}
	panic(fmt.Errorf("%v: Unexpected Expr type: %v (%T)", f.fi.name, e, e))
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
			return value.V, nil
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
		if name, ok := head.V.(Ident); ok {
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
				last := len(a.List) - 1
				if id, ok := a.List[last].(Ident); ok {
					if _, ok := ParseType(string(id)); ok {
						// Last statement is type declaration
						last--
					}
				}
				if last == 0 {
					return nil, fmt.Errorf("do: empty body")
				}
				if last > 1 {
					for _, st := range a.List[1:last] {
						if _, err := f.evalExpr(st); err != nil {
							return nil, err
						}
					}
				}
				return f.lastExpr(a.List[last])
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
			if name == "set" || name == "set'" {
				tail, _ := a.Tail()
				if err := f.setVar(tail.(*Sexpr) /*scoped*/, name == "set'"); err != nil {
					return nil, err
				}
				return QEmpty, nil
			}
			if name == "gen" || name == "gen'" {
				tail, _ := a.Tail()
				return f.evalGen(tail.(*Sexpr) /*hashable*/, name == "gen'")
			}
			if name == "apply" {
				tail, _ := a.Tail()
				return f.evalApply(tail.(*Sexpr))
			}
		}

		// return unevaluated list
		return a, nil
	case *LazyList:
		return a, nil
	}
	panic(fmt.Errorf("%v: Unexpected Expr type: %v (%T)", f.fi.name, e, e))
}

func (f *FuncRuntime) evalParameter(expr Expr) (p *Param, err error) {
	var forceType *Type
	defer func() {
		if p != nil && forceType != nil {
			p.T = *forceType
		}
		// fmt.Fprintf(os.Stderr, "%v: evalParameter(%v) = %v, %v\n", f.fi.name, expr, p, err)
	}()
	e, ft, err := f.lastParameter(expr)
	forceType = ft
	if err != nil {
		return nil, err
	}
	lst, ok := e.V.(*Sexpr)
	if !ok {
		// nothing to evaluate
		return e, nil
	}
	if lst.Quoted || lst.Len() == 0 {
		p := &Param{V: lst, T: TypeList}
		return p, nil
	}
	return f.evalFunc(lst)
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
	res, err := f.evalFunc(lst)
	if err != nil {
		return nil, err
	}
	return res.V, err
}

// (var-name) (value)
func (f *FuncRuntime) setVar(se *Sexpr, scoped bool) error {
	if se.Len() != 2 && se.Len() != 3 {
		return fmt.Errorf("set wants 2 or 3 arguments, found %v", se)
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0])
	}
	value, err := f.evalParameter(se.List[1])
	if err != nil {
		return err
	}
	f.vars[string(name)] = *value
	if scoped {
		f.scopedVars = append(f.scopedVars, string(name))
	}
	return nil
}

// (iter) (init-state)
func (f *FuncRuntime) evalGen(se *Sexpr, hashable bool) (Expr, error) {
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
	state, err := f.evalParameter(se.List[1])
	if err != nil {
		return nil, err
	}
	return NewLazyList(fu, state, hashable), nil
}

func (f *FuncRuntime) findFunc(fname string) (Evaler, error) {
	// Ability to pass function name as argument
	if v, ok := f.vars[fname]; ok {
		if v.T != TypeFunc && v.T != TypeUnknown {
			return nil, fmt.Errorf("%v: incorrect type of '%v', expected :func, found: %v", f.fi.name, fname, v)
		}
		vident, ok := v.V.(Ident)
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
func (f *FuncRuntime) evalFunc(se *Sexpr) (*Param, error) {
	head, err := se.Head()
	if err != nil {
		return nil, err
	}
	name, ok := head.V.(Ident)
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
	args := make([]Param, 0, len(tail.List))
	for _, arg := range tail.List {
		res, err := f.evalParameter(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, *res)
	}
	result, err := fu.Eval(args)
	return result, err
}

func (f *FuncRuntime) evalLambda(se *Sexpr) (Expr, error) {
	name := f.fi.interpret.NewLambdaName()
	fi := NewFuncInterpret(f.fi.interpret, name)
	body := f.replaceVars(se.List)
	fi.AddImpl(nil, body, false, TypeAny)
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
				res = append(res, v.V)
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
		cmd = append(cmd, h.V)
		args, _ = args.Tail()
	}

	return &Sexpr{
		List: cmd,
	}, nil
}

func (f *FuncInterpret) matchParameters(argfmt *ArgFmt, params []Param) (result bool) {
	defer func() {
		// fmt.Fprintf(os.Stderr, "%v: match (%v, %v) = %v\n", f.name, argfmt, params, result)
	}()
	if argfmt == nil {
		// null matches everything (lambda case)
		return true
	}
	if argfmt.Wildcard != "" {
		return true
	}

	binds := map[string]Expr{}
	if len(argfmt.Args) != len(params) {
		return false
	}
	for i, arg := range argfmt.Args {
		param := params[i]
		if !f.matchParam(&arg, &param) {
			return false
		}
		if arg.Name == "" {
			continue
		}
		if param.V == nil {
			continue
		}
		if binded, ok := binds[arg.Name]; ok {
			if !Equal(binded, param.V) {
				return false
			}
		}
		binds[arg.Name] = param.V
	}
	return true
}

func (f *FuncInterpret) matchParam(a *Arg, p *Param) bool {
	defer func() {
		// fmt.Fprintf(os.Stderr, "%v: match (%v, %v) = %v\n", f.name, argfmt, params, result)
	}()
	if a.T == TypeUnknown && a.V == nil {
		return true
	}
	if p.T == TypeUnknown && a.V == nil {
		return true
	}
	if a.T == TypeAny && a.V == nil {
		return true
	}
	if a.T != p.T && p.T != TypeUnknown && p.T != TypeAny {
		return false
	}
	if p.V == nil {
		// not a real parameter, just a Type binder
		return true
	}
	if a.V == nil {
		// anything of this corresponding type matches
		return true
	}
	return Equal(a.V, p.V)
}

func (f *FuncRuntime) cleanup() {
	for _, varname := range f.scopedVars {
		expr := f.vars[varname]
		switch a := expr.V.(type) {
		case Ident:
			f.fi.interpret.DeleteLambda(string(a))
		case io.Closer:
			if err := a.Close(); err != nil {
				log.Printf("Close() failed: %v", err)
			}
		default:
			log.Printf("Don't know how to clean variable of type: %v", expr)
		}
	}
	f.scopedVars = f.scopedVars[:0]
	f.vars = make(map[string]Param)
}
