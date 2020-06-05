package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// User-defined functions

type FuncInterpret struct {
	interpret    *Interpret
	name         string
	bodies       []*FuncImpl
	returnType   Type
	capturedVars map[string]*Param

	genericReturnTypes map[string]Type
}

func (f *FuncInterpret) FuncType() Type {
	ft := f.bodies[0].funcType
	for _, impl := range f.bodies[1:] {
		if impl.funcType != ft {
			ft = TypeFunc
			break
		}
	}
	return ft
}

type FuncImpl struct {
	argfmt *ArgFmt
	body   []Param
	// Do we need to remenber function results?
	memo bool
	// Function results: args.Repr() -> Result
	results map[string]*Param
	// return type
	returnType Type
	// function type
	funcType Type
}

func NewFuncImpl(argfmt *ArgFmt, body []Param, memo bool, returnType Type) *FuncImpl {
	i := &FuncImpl{
		argfmt:     argfmt,
		body:       body,
		memo:       memo,
		returnType: returnType,
		funcType:   makeFuncType(argfmt, returnType),
	}
	if memo {
		i.results = make(map[string]*Param)
	}
	return i
}

func makeFuncType(argfmt *ArgFmt, retType Type) Type {
	if argfmt == nil {
		return TypeFunc
	}
	res := "func["
	if argfmt.Wildcard != "" {
		res += "list...,"
	} else {
		for _, arg := range argfmt.Args {
			res += string(arg.T) + ","
		}
	}
	res += string(retType) + "]"
	return Type(res)
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
		interpret:          i,
		name:               name,
		returnType:         TypeUnknown,
		capturedVars:       make(map[string]*Param),
		genericReturnTypes: make(map[string]Type),
	}
}

func (f *FuncInterpret) AddImpl(argfmt Expr, body []Param, memo bool, returnType Type) error {
	returnType = f.interpret.UnaliasType(returnType)
	if len(f.bodies) > 0 && returnType != f.returnType {
		return fmt.Errorf("%v: cannot redefine return type: previous %v, current %v", f.name, f.returnType, returnType)
	}
	if argfmt == nil {
		f.bodies = append(f.bodies, NewFuncImpl(nil, body, memo, returnType))
		return nil
	}
	af, err := ParseArgFmt(argfmt)
	if err != nil {
		return err
	}
	f.bodies = append(f.bodies, NewFuncImpl(af, body, memo, returnType))

	f.returnType = returnType
	return nil
}

func (f *FuncInterpret) AddVar(name string, p *Param) {
	f.capturedVars[name] = p
}

func (f *FuncInterpret) TryBindAll(params []Param) (rt Type, err error) {
	// a bit of hack
	hash := fmt.Sprintf("%v", params)
	if t, ok := f.genericReturnTypes[hash]; ok {
		return t, nil
	}

	rtDefined := false
	for _, im := range f.bodies {
		if ok, types := f.matchParameters(im.argfmt, params); ok {
			t := im.returnType.Expand(types)
			if rtDefined {
				if t != rt {
					return "", fmt.Errorf("%v: different implmentations returns different type: %v != %v", f.name, rt, t)
				}
			} else {
				f.genericReturnTypes[hash] = t
				rtDefined = true
			}
			rt = t
			if len(types) > 0 {
				// check that generics are matching
				values := map[string]Type{}
				for i, arg := range im.argfmt.Args {
					values[arg.Name] = params[i].T
				}
				tt, err := f.interpret.evalBodyType(f.name, im.body, values, types)
				if newTt, ok := types[tt.Basic()]; ok {
					tt = newTt
				}

				if err != nil {
					return "", err
				}
				if t != tt && tt != TypeUnknown {
					return "", fmt.Errorf("%v: mismatch return type: declared %v != actual %v", f.name, t, tt)
				}
			}
		}
	}
	if !rtDefined {
		return "", fmt.Errorf("%v: no matching function implementation found for %v", f.name, params)
	}
	return rt, nil
}

func (f *FuncInterpret) TryBind(params []Param) (num int, rt Type, types map[string]Type, err error) {
	for idx, im := range f.bodies {
		if ok, types := f.matchParameters(im.argfmt, params); ok {
			t := im.returnType.Expand(types)
			return idx, t, types, nil
		}
	}
	return -1, TypeUnknown, nil, fmt.Errorf("%v: no matching function implementation found for %v", f.name, params)
}

func (f *FuncInterpret) Eval(params []Param) (result *Param, err error) {
	run := NewFuncRuntime(f)
	impl, result, rt, types, err := run.bind(params)
	if err != nil {
		return nil, err
	}
	if result != nil {
		return result, nil
	}
	run.types = types
	res, err := run.Eval(impl)
	if err != nil {
		return nil, err
	}
	run.cleanup()
	newT, err := run.updateType(res.T, rt)
	if err != nil {
		return nil, fmt.Errorf("Cannot cast type %v to %v: %v", res.T, rt, err)
	}
	res.T = newT
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
	types      map[string]Type
}

func NewFuncRuntime(fi *FuncInterpret) *FuncRuntime {
	return &FuncRuntime{
		fi:   fi,
		vars: make(map[string]Param),
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

func (f *FuncRuntime) bind(params []Param) (impl *FuncImpl, result *Param, resultType Type, types map[string]Type, err error) {
	f.cleanup()
	args := make([]Expr, 0, len(params))
	for _, p := range params {
		args = append(args, p.V)
	}
	idx, rt, types, err := f.fi.TryBind(params)
	if err != nil {
		return nil, nil, "", nil, err
	}
	impl = f.fi.bodies[idx]
	if impl.memo {
		keyArgs, err := keyOfArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot compute hash of args: %v, %v\n", args, err)
		} else if res, ok := impl.results[keyArgs]; ok {
			return nil, res, "", nil, nil
		}
	}

	if impl.argfmt != nil {
		if impl.argfmt.Wildcard != "" {
			f.vars[impl.argfmt.Wildcard] = Param{
				V: &Sexpr{List: params, Quoted: true},
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
		V: &Sexpr{List: params, Quoted: true},
		T: TypeList,
	}
	for i, arg := range params {
		f.vars[fmt.Sprintf("_%d", i+1)] = arg
	}
	f.args = args
	return impl, nil, rt, types, nil
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
		var bodyForceType *Type
		if id, ok := impl.body[last].V.(Ident); ok {
			if tp, ok := ParseType(string(id)); ok {
				// Last statement is type declaration
				last--
				tp = tp.Expand(f.types)
				bodyForceType = &tp
			}
		}
		for i, expr := range impl.body {
			if i == last {
				// check for tail call
				e, forceType, err := f.lastParameter(&expr)
				if err != nil {
					return nil, err
				}
				lst, ok := e.V.(*Sexpr)
				if !ok {
					if forceType != nil {
						newT, err := f.updateType(e.T, *forceType)
						if err != nil {
							return nil, fmt.Errorf("Cannot cast %v to %v: %v", e.T, *forceType, err)
						}
						e.T = newT
					}
					if bodyForceType != nil {
						newT, err := f.updateType(e.T, *bodyForceType)
						if err != nil {
							return nil, fmt.Errorf("Cannot cast %v to %v: %v", e.T, *bodyForceType, err)
						}
						e.T = newT
					}
					if memoImpl.memo {
						// lets remenber the result
						memoImpl.RememberResult(f.fi.name, memoArgs, e)
					}
					// nothing to evaluate
					return e, nil
				}
				if lst.Quoted || lst.Length() == 0 {
					p := &Param{V: lst, T: TypeList}
					if forceType != nil {
						newT, err := f.updateType(p.T, *forceType)
						if err != nil {
							return nil, err
						}
						p.T = newT
					}
					if bodyForceType != nil {
						newT, err := f.updateType(p.T, *bodyForceType)
						if err != nil {
							return nil, err
						}
						p.T = newT
					}
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
					if forceType != nil {
						newT, err := f.updateType(result.T, *forceType)
						if err != nil {
							return nil, err
						}
						result.T = newT
					}
					if bodyForceType != nil {
						newT, err := f.updateType(result.T, *bodyForceType)
						if err != nil {
							return nil, err
						}
						result.T = newT
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
					arg, err := f.evalParameter(&ar)
					if err != nil {
						return nil, err
					}
					args = append(args, *arg)
				}
				var result *Param
				impl, result, _, _, err = f.bind(args)
				if err != nil {
					return nil, err
				}
				if result != nil {
					return result, nil
				}
				continue L
			} else {
				res, err = f.evalParameter(&expr)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	// Empty main body
	return &Param{V: QEmpty, T: TypeList}, nil
}

func (f *FuncRuntime) lastParameter(e *Param) (*Param, *Type, error) {
	switch a := e.V.(type) {
	case Int:
		return e, nil, nil
	case Str:
		return e, nil, nil
	case Bool:
		return e, nil, nil
	case Ident:
		result := e
		if value, ok := f.findVar(string(a)); ok {
			result = value
		}
		if id, ok := result.V.(Ident); ok {
			if fe, ok := f.fi.interpret.funcs[string(id)]; ok {
				if fi, ok := fe.(*FuncInterpret); ok {
					result.T = fi.FuncType()
				} else {
					result.T = TypeFunc
				}
			}
		}
		return result, nil, nil
	case *Sexpr:
		if a.Quoted {
			return &Param{V: a, T: TypeList}, nil, nil
		}
		if a.Length() == 0 {
			return nil, nil, fmt.Errorf("%v: Unexpected empty s-expression: %v", f.fi.name, a)
		}
		head, _ := a.Head()
		if name, ok := head.V.(Ident); ok {
			if a.Lambda {
				lm, err := f.evalLambda(&Sexpr{List: []Param{{V: a, T: TypeList}}, Quoted: true})
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
				res, err := f.evalParameter(&arg)
				if err != nil {
					return nil, nil, err
				}
				boolRes, ok := res.V.(Bool)
				if !ok {
					return nil, nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
				}
				if bool(boolRes) {
					return f.lastParameter(&a.List[2])
				}
				return f.lastParameter(&a.List[3])
			}
			if name == "do" {
				var retType *Type
				last := len(a.List) - 1
				if id, ok := a.List[last].V.(Ident); ok {
					if rt, ok := ParseType(string(id)); ok {
						// Last statement is type declaration
						last--
						rt = rt.Expand(f.types)
						retType = &rt
					}
				}
				if last == 0 {
					return nil, nil, fmt.Errorf("do: empty body")
				}
				if last > 1 {
					for _, st := range a.List[1:last] {
						if _, err := f.evalParameter(&st); err != nil {
							return nil, nil, err
						}
					}
				}
				ret, ft, err := f.lastParameter(&a.List[last])
				if err != nil {
					return nil, nil, err
				}
				if retType != nil {
					// TODO check matching types
					newT, err := f.updateType(ret.T, *retType)
					if err != nil {
						return nil, nil, fmt.Errorf("Cannot cast %v to %v: %v", ret.T, *retType, err)
					}
					ret.T = newT
					ft = retType
				}
				return ret, ft, nil
			}
			if name == "and" {
				for _, arg := range a.List[1:] {
					res, err := f.evalParameter(&arg)
					if err != nil {
						return nil, nil, err
					}
					boolRes, ok := res.V.(Bool)
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
					res, err := f.evalParameter(&arg)
					if err != nil {
						return nil, nil, err
					}
					boolRes, ok := res.V.(Bool)
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

func (f *FuncRuntime) updateType(oldT, newT Type) (Type, error) {
	if oldT == TypeUnknown {
		return newT, nil
	}
	ok, err := f.fi.interpret.canConvertType(oldT, newT)
	if err != nil {
		return TypeUnknown, err
	}
	if ok {
		return oldT, nil
	}
	return newT, nil
}

func (f *FuncRuntime) evalParameter(expr *Param) (p *Param, err error) {
	var forceType *Type
	defer func() {
		if p != nil && forceType != nil {
			p.T, err = f.updateType(p.T, *forceType)
			if err != nil {
				p = nil
			}
		}
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
	if lst.Quoted || lst.Length() == 0 {
		return e, nil
	}
	return f.evalFunc(lst)
}

// (var-name) (value)
func (f *FuncRuntime) setVar(se *Sexpr, scoped bool) error {
	if se.Length() != 2 && se.Length() != 3 {
		return fmt.Errorf("set wants 2 or 3 arguments, found %v", se)
	}
	name, ok := se.List[0].V.(Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0])
	}
	value, err := f.evalParameter(&se.List[1])
	if err != nil {
		return err
	}
	if se.Length() == 3 {
		id, ok := se.List[2].V.(Ident)
		if !ok {
			return fmt.Errorf("%v: set expects type identifier, found: %v", f.fi.name, se.List[2])
		}
		t, err := f.fi.interpret.parseType(string(id))
		if err != nil {
			return err
		}
		newT, err := f.updateType(value.T, t.Expand(f.types))
		if err != nil {
			return fmt.Errorf("Cannot cast type %v to %v: %v", value.T, t, err)
		}
		value.T = newT
	}
	f.vars[string(name)] = *value
	if scoped {
		f.scopedVars = append(f.scopedVars, string(name))
	}
	return nil
}

func (f *FuncRuntime) findVar(name string) (*Param, bool) {
	if p, ok := f.vars[name]; ok {
		return &p, true
	} else if p, ok := f.fi.capturedVars[name]; ok {
		return p, true
	}
	return nil, false
}

// (iter) (init-state)
func (f *FuncRuntime) evalGen(se *Sexpr, hashable bool) (Expr, error) {
	if se.Length() < 2 {
		return nil, fmt.Errorf("gen wants at least 2 arguments, found %v", se)
	}
	fn, err := f.evalParameter(&se.List[0])
	if err != nil {
		return nil, err
	}
	fident, ok := fn.V.(Ident)
	if !ok {
		return nil, fmt.Errorf("gen expects first argument to be a funtion, found: %v", se.List[0])
	}
	fu, err := f.findFunc(string(fident))
	if err != nil {
		return nil, err
	}
	var state []Param
	for _, a := range se.List[1:] {
		s, err := f.evalParameter(&a)
		if err != nil {
			return nil, err
		}
		state = append(state, *s)
	}
	return NewLazyList(fu, state, hashable), nil
}

func (f *FuncRuntime) findFunc(fname string) (result Evaler, err error) {
	// Ability to pass function name as argument
	if v, ok := f.findVar(fname); ok {
		if v.T.Basic() != "func" && v.T != TypeUnknown {
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
func (f *FuncRuntime) evalFunc(se *Sexpr) (result *Param, err error) {
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
		res, err := f.evalParameter(&arg)
		if err != nil {
			return nil, err
		}
		args = append(args, *res)
	}
	result, err = fu.Eval(args)
	return result, err
}

func (f *FuncRuntime) evalLambda(se *Sexpr) (Expr, error) {
	name := f.fi.interpret.NewLambdaName()
	fi := NewFuncInterpret(f.fi.interpret, name)
	body := f.replaceVars(se.List, fi)
	fi.AddImpl(nil, body, false, TypeUnknown)
	f.fi.interpret.funcs[name] = fi
	return Ident(name), nil
}

var lambdaArgRe = regexp.MustCompile(`^(_[0-9]+|__args)$`)

func (f *FuncRuntime) replaceVars(st []Param, fi *FuncInterpret) (res []Param) {
	for _, s := range st {
		switch a := s.V.(type) {
		case *Sexpr:
			v := &Sexpr{Quoted: a.Quoted}
			v.List = f.replaceVars(a.List, fi)
			res = append(res, Param{V: v, T: s.T})
		case Ident:
			if lambdaArgRe.MatchString(string(a)) {
				res = append(res, Param{V: a, T: s.T})
			} else if v, ok := f.findVar(string(a)); ok {
				fi.AddVar(string(a), v)
				res = append(res, Param{V: a, T: s.T})
			} else {
				res = append(res, Param{V: a, T: s.T})
			}
		default:
			res = append(res, s)
		}
	}
	return res
}

// function list-of-args
func (f *FuncRuntime) evalApply(se *Sexpr) (Expr, error) {
	if len(se.List) != 2 {
		return nil, fmt.Errorf("apply expects function with list of arguments")
	}
	res, err := f.evalParameter(&se.List[1])
	if err != nil {
		return nil, err
	}
	args, ok := res.V.(List)
	if !ok {
		return nil, fmt.Errorf("apply expects result to be a list of argument")
	}
	cmd := []Param{se.List[0]}
	for !args.Empty() {
		h, _ := args.Head()
		cmd = append(cmd, *h)
		args, _ = args.Tail()
	}

	return &Sexpr{
		List: cmd,
	}, nil
}

func (f *FuncInterpret) matchParameters(argfmt *ArgFmt, params []Param) (result bool, types map[string]Type) {
	if argfmt == nil {
		// null matches everything (lambda case)
		return true, nil
	}
	if argfmt.Wildcard != "" {
		return true, nil
	}

	binds := map[string]Expr{}
	typeBinds := map[string]Type{}
	if len(argfmt.Args) != len(params) {
		return false, nil
	}
	for i, arg := range argfmt.Args {
		param := params[i]
		match, err := f.interpret.matchType(arg.T, param.T, &typeBinds)
		if err != nil {
			return false, nil
		}
		if !match {
			return false, nil
		}
		if !f.matchValue(&arg, &param) {
			return false, nil
		}
		if arg.Name == "" {
			continue
		}
		if param.V == nil {
			continue
		}
		if binded, ok := binds[arg.Name]; ok {
			if !Equal(binded, param.V) {
				return false, nil
			}
		}
		binds[arg.Name] = param.V
	}
	return true, typeBinds
}

func (f *FuncInterpret) matchValue(a *Arg, p *Param) bool {
	if a.V == nil || p.V == nil {
		return true
	}
	return Equal(a.V, p.V)
}

func (f *FuncInterpret) matchParam(a *Arg, p *Param) bool {
	if f.interpret.IsContract(a.T) {
		return true
	}
	if a.T == TypeUnknown && a.V == nil {
		return true
	}
	if p.T == TypeUnknown && a.V == nil {
		return true
	}
	if a.T == TypeAny && a.V == nil {
		return true
	}
	if l, ok := a.V.(List); ok && l.Empty() {
		pl, ok := p.V.(List)
		return ok && pl.Empty()
	}
	if a.T != p.T && p.T != TypeUnknown {
		canConvert, err := f.interpret.canConvertType(p.T, a.T)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", f.name, err)
			return false
		}
		if !canConvert && p.T != TypeUnknown {
			return false
		}
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
				fmt.Fprintf(os.Stderr, "Close() failed: %v\n", err)
			}
		default:
			fmt.Fprintf(os.Stderr, "Don't know how to clean variable of type: %v\n", expr)
		}
	}
	f.scopedVars = f.scopedVars[:0]
}

func (i *Interpret) matchType(arg Type, val Type, typeBinds *map[string]Type) (result bool, eerroorr error) {
	arg = i.UnaliasType(arg)
	val = i.UnaliasType(val)

	if i.IsContract(arg) {
		if bind, ok := (*typeBinds)[arg.Basic()]; ok && string(bind) != strings.TrimLeft(string(val), ":") {
			return false, nil
		}
		(*typeBinds)[arg.Basic()] = Type(strings.TrimLeft(string(val), ":"))
		return true, nil
	}
	if val == TypeUnknown || arg == TypeUnknown {
		return true, nil
	}
	if (arg == TypeFunc && val.Basic() == "func") || (val == TypeFunc && arg.Basic() == "func") {
		return true, nil
	}

	parent, err := i.toParent(val, Type(arg.Basic()))
	if err != nil {
		return false, err
	}

	aParams := arg.Arguments()
	vParams := parent.Arguments()
	if len(aParams) != len(vParams) {
		return false, nil
	}
	for j, p := range aParams {
		ok, err := i.matchType(Type(p), Type(vParams[j]), typeBinds)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}
