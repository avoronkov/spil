package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/avoronkov/spil/types"
)

// User-defined functions

type FuncInterpret struct {
	interpret    *Interpret
	name         string
	bodies       []*FuncImpl
	capturedVars map[string]*types.Value

	genericReturnTypes map[string]types.Type
}

func (f *FuncInterpret) FuncType() types.Type {
	ft := f.bodies[0].funcType
	for _, impl := range f.bodies[1:] {
		if impl.funcType != ft {
			ft = types.TypeFunc
			break
		}
	}
	return ft
}

type FuncImpl struct {
	argfmt *ArgFmt
	body   []types.Value
	// Do we need to remenber function results?
	memo bool
	// Function results: args.Repr() -> Result
	results map[string]*types.Value
	// return type
	returnType types.Type
	// function type
	funcType types.Type
}

func NewFuncImpl(argfmt *ArgFmt, body []types.Value, memo bool, returnType types.Type) *FuncImpl {
	i := &FuncImpl{
		argfmt:     argfmt,
		body:       body,
		memo:       memo,
		returnType: returnType,
		funcType:   makeFuncType(argfmt, returnType),
	}
	if memo {
		i.results = make(map[string]*types.Value)
	}
	return i
}

func makeFuncType(argfmt *ArgFmt, retType types.Type) types.Type {
	if argfmt == nil {
		return types.TypeFunc
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
	return types.Type(res)
}

func (i *FuncImpl) RememberResult(name string, args []types.Expr, result *types.Value) {
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
		capturedVars:       make(map[string]*types.Value),
		genericReturnTypes: make(map[string]types.Type),
	}
}

func (f *FuncInterpret) AddImpl(argfmt types.Expr, body []types.Value, memo bool, returnType types.Type) error {
	returnType = f.interpret.UnaliasType(returnType)
	if argfmt == nil {
		f.bodies = append(f.bodies, NewFuncImpl(nil, body, memo, returnType))
		return nil
	}
	af, err := ParseArgFmt(argfmt)
	if err != nil {
		return err
	}
	f.bodies = append(f.bodies, NewFuncImpl(af, body, memo, returnType))

	return nil
}

func (f *FuncInterpret) AddVar(name string, p *types.Value) {
	f.capturedVars[name] = p
}

func (f *FuncInterpret) TryBindAll(params []types.Value) (rt types.Type, err error) {
	// a bit of hack
	hash := fmt.Sprintf("%v", params)
	if t, ok := f.genericReturnTypes[hash]; ok {
		return t, nil
	}

	rtDefined := false
	for _, im := range f.bodies {
		if ok, tps := f.matchParameters(im.argfmt, params); ok {
			t := im.returnType.Expand(tps)
			if rtDefined {
				if t != rt {
					rt = types.TypeAny
					f.genericReturnTypes[hash] = rt
					// fmt.Fprintf(os.Stderr, "%v: different implmentations returns different type: %v != %v", f.name, rt, t)
				}
			} else {
				f.genericReturnTypes[hash] = t
				rtDefined = true
				rt = t
			}
			if len(tps) > 0 {
				// check that generics are matching
				values := make(map[string](types.Type))
				for i, arg := range im.argfmt.Args {
					values[arg.Name] = params[i].T
				}
				tt, err := f.interpret.evalBodyType(f.name, im.body, values, tps)
				if newTt, ok := tps[tt.Basic()]; ok {
					tt = newTt
				}

				if err != nil {
					return "", err
				}
				if t != tt && tt != types.TypeUnknown {
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

func (f *FuncInterpret) TryBind(params []types.Value) (num int, rt types.Type, tps map[string]types.Type, err error) {
	for idx, im := range f.bodies {
		if ok, tps := f.matchParameters(im.argfmt, params); ok {
			t := im.returnType.Expand(tps)
			return idx, t, tps, nil
		}
	}
	return -1, types.TypeUnknown, nil, fmt.Errorf("%v: TryBind: no matching function implementation found for %v", f.name, params)
}

func (f *FuncInterpret) Eval(params []types.Value) (result *types.Value, err error) {
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

type FuncRuntime struct {
	fi   *FuncInterpret
	vars map[string]types.Value
	args []types.Expr
	// variables that should be Closed after leaving this variable scope.
	scopedVars []string
	types      map[string]types.Type
}

func NewFuncRuntime(fi *FuncInterpret) *FuncRuntime {
	return &FuncRuntime{
		fi:   fi,
		vars: make(map[string]types.Value),
	}
}

func keyOfArgs(args []types.Expr) (string, error) {
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

func (f *FuncRuntime) bind(params []types.Value) (impl *FuncImpl, result *types.Value, resultType types.Type, tps map[string]types.Type, err error) {
	f.cleanup()
	args := make([]types.Expr, 0, len(params))
	for _, p := range params {
		args = append(args, p.E)
	}
	idx, rt, tps, err := f.fi.TryBind(params)
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
			f.vars[impl.argfmt.Wildcard] = types.Value{
				E: &types.Sexpr{List: params, Quoted: true},
				T: types.TypeList,
			}
		} else {
			varArgs := false
			for i, arg := range impl.argfmt.Args {
				if arg.T.Basic() == "args" {
					// var args
					varArgs = true
					if len(params) > i && params[i].T.Basic() == "args" {
						f.vars[arg.Name] = params[i]
					} else {
						f.vars[arg.Name] = types.Value{
							E: &types.Sexpr{
								Quoted: true,
								List:   params[i:],
							},
							T: arg.T.Expand(tps),
						}
					}
					break
				}
				if arg.V == nil {
					f.vars[arg.Name] = params[i]
				}
			}
			if !varArgs && len(impl.argfmt.Args) != len(params) {
				err = fmt.Errorf("Incorrect number of arguments to %v: expected %v, found %v", f.fi.name, len(impl.argfmt.Args), len(params))
				return
			}
		}
	}
	// bind to __args and _1, _2 ... variables
	f.vars["__args"] = types.Value{
		E: &types.Sexpr{List: params, Quoted: true},
		T: types.TypeList,
	}
	for i, arg := range params {
		f.vars[fmt.Sprintf("_%d", i+1)] = arg
	}
	f.args = args
	return impl, nil, rt, tps, nil
}

func (f *FuncRuntime) Eval(impl *FuncImpl) (res *types.Value, err error) {
	memoImpl := impl
	memoArgs := f.args
L:
	for {
		last := len(impl.body) - 1
		if last < 0 {
			break L
		}
		var bodyForceType *types.Type
		if id, ok := impl.body[last].E.(types.Ident); ok {
			if tp, ok := types.ParseType(string(id)); ok {
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
				lst, ok := e.E.(*types.Sexpr)
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
					p := &types.Value{E: lst, T: types.TypeList}
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
				hident, ok := head.E.(types.Ident)
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
				tail := t.(*types.Sexpr)
				// eval args
				args := make([]types.Value, 0, len(tail.List))
				for _, ar := range tail.List {
					arg, err := f.evalParameter(&ar)
					if err != nil {
						return nil, err
					}
					args = append(args, *arg)
				}
				var result *types.Value
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
	return &types.Value{E: types.QEmpty, T: types.TypeList}, nil
}

func (f *FuncRuntime) lastParameter(e *types.Value) (*types.Value, *types.Type, error) {
	switch a := e.E.(type) {
	case types.Int, types.Float, types.Str, types.Bool:
		return e, nil, nil
	case types.Ident:
		result := e
		if value, ok := f.findVar(string(a)); ok {
			result = value
		}
		if id, ok := result.E.(types.Ident); ok {
			if fe, ok := f.fi.interpret.funcs[string(id)]; ok {
				if fi, ok := fe.(*FuncInterpret); ok {
					result.T = fi.FuncType()
				} else {
					result.T = types.TypeFunc
				}
			}
		}
		return result, nil, nil
	case *types.Sexpr:
		if a.Quoted {
			return &types.Value{E: a, T: types.TypeList}, nil, nil
		}
		if a.Length() == 0 {
			return nil, nil, fmt.Errorf("%v: Unexpected empty s-expression: %v", f.fi.name, a)
		}
		head, _ := a.Head()
		if name, ok := head.E.(types.Ident); ok {
			if a.Lambda {
				lm, err := f.evalLambda(&types.Sexpr{List: []types.Value{{E: a, T: types.TypeList}}, Quoted: true})
				if err != nil {
					return nil, nil, err
				}
				return &types.Value{
					E: lm,
					T: types.TypeFunc,
				}, nil, nil
			}
			if name == "lambda" {
				tail, _ := a.Tail()
				lm, err := f.evalLambda(tail.(*types.Sexpr))
				if err != nil {
					return nil, nil, err
				}
				return &types.Value{
					E: lm,
					T: types.TypeFunc,
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
				boolRes, ok := res.E.(types.Bool)
				if !ok {
					return nil, nil, fmt.Errorf("Argument %v should evaluate to boolean value, actual %v", arg, res)
				}
				if bool(boolRes) {
					return f.lastParameter(&a.List[2])
				}
				return f.lastParameter(&a.List[3])
			}
			if name == "do" {
				var retType *types.Type
				last := len(a.List) - 1
				if id, ok := a.List[last].E.(types.Ident); ok {
					if rt, ok := types.ParseType(string(id)); ok {
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
					boolRes, ok := res.E.(types.Bool)
					if !ok {
						return nil, nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if !bool(boolRes) {
						return &types.Value{
							E: types.Bool(false),
							T: types.TypeBool,
						}, nil, nil
					}
				}
				return &types.Value{
					E: types.Bool(true),
					T: types.TypeBool,
				}, nil, nil
			}
			if name == "or" {
				for _, arg := range a.List[1:] {
					res, err := f.evalParameter(&arg)
					if err != nil {
						return nil, nil, err
					}
					boolRes, ok := res.E.(types.Bool)
					if !ok {
						return nil, nil, fmt.Errorf("and: rrgument %v should evaluate to boolean value, actual %v", arg, res)
					}
					if bool(boolRes) {
						return &types.Value{
							E: types.Bool(true),
							T: types.TypeBool,
						}, nil, nil
					}
				}
				return &types.Value{
					E: types.Bool(false),
					T: types.TypeBool,
				}, nil, nil
			}
			if name == "set" || name == "set'" {
				tail, _ := a.Tail()
				if err := f.setVar(tail.(*types.Sexpr) /*scoped*/, name == "set'"); err != nil {
					return nil, nil, err
				}
				return &types.Value{E: types.QEmpty, T: types.TypeAny}, nil, nil
			}
			if name == "gen" || name == "gen'" {
				tail, _ := a.Tail()
				gen, err := f.evalGen(tail.(*types.Sexpr) /*hashable*/, name == "gen'")
				if err != nil {
					return nil, nil, err
				}
				return &types.Value{E: gen, T: types.TypeList}, nil, nil
			}
			if name == "apply" {
				tail, _ := a.Tail()
				res, err := f.evalApply(tail.(*types.Sexpr))
				if err != nil {
					return nil, nil, err
				}
				return &types.Value{E: res, T: types.TypeUnknown}, nil, nil
			}
		}

		// return unevaluated list
		return &types.Value{E: a, T: types.TypeUnknown}, nil, nil
	case *LazyList:
		return &types.Value{E: a, T: types.TypeList}, nil, nil
	}
	panic(fmt.Errorf("%v: Unexpected Expr type: %v (%T)", f.fi.name, e, e))
}

func (f *FuncRuntime) updateType(oldT, newT types.Type) (types.Type, error) {
	if oldT == types.TypeUnknown {
		return newT, nil
	}
	ok, err := f.fi.interpret.canConvertType(oldT, newT)
	if err != nil {
		return types.TypeUnknown, err
	}
	if ok {
		return oldT, nil
	}
	return newT, nil
}

func (f *FuncRuntime) evalParameter(expr *types.Value) (p *types.Value, err error) {
	var forceType *types.Type
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
	lst, ok := e.E.(*types.Sexpr)
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
func (f *FuncRuntime) setVar(se *types.Sexpr, scoped bool) error {
	if se.Length() != 2 && se.Length() != 3 {
		return fmt.Errorf("set wants 2 or 3 arguments, found %v", se)
	}
	name, ok := se.List[0].E.(types.Ident)
	if !ok {
		return fmt.Errorf("set expected identifier first, found %v", se.List[0])
	}
	value, err := f.evalParameter(&se.List[1])
	if err != nil {
		return err
	}
	if se.Length() == 3 {
		id, ok := se.List[2].E.(types.Ident)
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

func (f *FuncRuntime) findVar(name string) (*types.Value, bool) {
	if p, ok := f.vars[name]; ok {
		return &p, true
	} else if p, ok := f.fi.capturedVars[name]; ok {
		return p, true
	}
	return nil, false
}

// (iter) (init-state)
func (f *FuncRuntime) evalGen(se *types.Sexpr, hashable bool) (types.Expr, error) {
	if se.Length() < 2 {
		return nil, fmt.Errorf("gen wants at least 2 arguments, found %v", se)
	}
	fn, err := f.evalParameter(&se.List[0])
	if err != nil {
		return nil, err
	}
	fident, ok := fn.E.(types.Ident)
	if !ok {
		return nil, fmt.Errorf("gen expects first argument to be a funtion, found: %v", se.List[0])
	}
	fu, err := f.findFunc(string(fident))
	if err != nil {
		return nil, err
	}
	var state []types.Value
	for _, a := range se.List[1:] {
		s, err := f.evalParameter(&a)
		if err != nil {
			return nil, err
		}
		state = append(state, *s)
	}
	return NewLazyList(fu, state, hashable), nil
}

func (f *FuncRuntime) findFunc(fname string) (result types.Function, err error) {
	// Ability to pass function name as argument
	if v, ok := f.findVar(fname); ok {
		if v.T.Basic() != "func" && v.T != types.TypeUnknown {
			return nil, fmt.Errorf("%v: incorrect type of '%v', expected :func, found: %v", f.fi.name, fname, v)
		}
		vident, ok := v.E.(types.Ident)
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
func (f *FuncRuntime) evalFunc(se *types.Sexpr) (result *types.Value, err error) {
	head, err := se.Head()
	if err != nil {
		return nil, err
	}
	name, ok := head.E.(types.Ident)
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
	tail := t.(*types.Sexpr)
	args := make([]types.Value, 0, len(tail.List))
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

func (f *FuncRuntime) evalLambda(se *types.Sexpr) (types.Expr, error) {
	name := f.fi.interpret.NewLambdaName()
	fi := NewFuncInterpret(f.fi.interpret, name)
	body := f.replaceVars(se.List, fi)
	fi.AddImpl(nil, body, false, types.TypeUnknown)
	f.fi.interpret.funcs[name] = fi
	return types.Ident(name), nil
}

var lambdaArgRe = regexp.MustCompile(`^(_[0-9]+|__args)$`)

func (f *FuncRuntime) replaceVars(st []types.Value, fi *FuncInterpret) (res []types.Value) {
	for _, s := range st {
		switch a := s.E.(type) {
		case *types.Sexpr:
			v := &types.Sexpr{Quoted: a.Quoted}
			v.List = f.replaceVars(a.List, fi)
			res = append(res, types.Value{E: v, T: s.T})
		case types.Ident:
			if lambdaArgRe.MatchString(string(a)) {
				res = append(res, types.Value{E: a, T: s.T})
			} else if v, ok := f.findVar(string(a)); ok {
				fi.AddVar(string(a), v)
				res = append(res, types.Value{E: a, T: s.T})
			} else {
				res = append(res, types.Value{E: a, T: s.T})
			}
		default:
			res = append(res, s)
		}
	}
	return res
}

// function list-of-args
func (f *FuncRuntime) evalApply(se *types.Sexpr) (types.Expr, error) {
	if len(se.List) != 2 {
		return nil, fmt.Errorf("apply expects function with list of arguments")
	}
	res, err := f.evalParameter(&se.List[1])
	if err != nil {
		return nil, err
	}
	args, ok := res.E.(types.List)
	if !ok {
		return nil, fmt.Errorf("apply expects result to be a list of argument")
	}
	cmd := []types.Value{se.List[0]}
	for !args.Empty() {
		h, _ := args.Head()
		cmd = append(cmd, *h)
		args, _ = args.Tail()
	}

	return &types.Sexpr{
		List: cmd,
	}, nil
}

func (f *FuncInterpret) matchParameters(argfmt *ArgFmt, params []types.Value) (result bool, tps map[string]types.Type) {
	if argfmt == nil {
		// null matches everything (lambda case)
		return true, nil
	}
	if argfmt.Wildcard != "" {
		return true, nil
	}

	binds := map[string]types.Expr{}
	typeBinds := map[string]types.Type{}
	for i, arg := range argfmt.Args {
		// check for varargs
		if arg.T.Basic() == "args" {
			if i != len(argfmt.Args)-1 {
				fmt.Fprintf(os.Stderr, "%v: parameter of type %v shoud go last\n", f.name, arg.T)
				return false, nil
			}
			// the rest of params should match the type
			targs := arg.T.Arguments()
			if len(targs) != 1 {
				fmt.Fprintf(os.Stderr, "%v: type %v should have exactly one parameter\n", f.name, arg.T)
				return false, nil
			}
			if len(params) > i {
				if params[i].T == arg.T {
					if len(params) != i+1 {
						return false, nil
					}
					return true, typeBinds
				}
			}
			expt := types.Type(targs[0])
			for _, param := range params[i:] {
				ok, err := f.interpret.matchType(expt, param.T, &typeBinds)
				if err != nil {
					// fmt.Fprintf(os.Stderr, "%v: %v\n", f.name, err)
					return false, nil
				}
				if !ok {
					// fmt.Fprintf(os.Stderr, "%v: incorrect type of parameter %v: expected %v, found %v\n", f.name, i+j, expt, param.T)
					return false, nil
				}
			}
			return true, typeBinds
		}
		if i >= len(params) {
			return false, nil
		}
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
		if param.E == nil {
			continue
		}
		if binded, ok := binds[arg.Name]; ok {
			if !types.Equal(binded, param.E) {
				return false, nil
			}
		}
		binds[arg.Name] = param.E
	}
	if len(argfmt.Args) != len(params) {
		return false, nil
	}
	return true, typeBinds
}

func (f *FuncInterpret) matchValue(a *Arg, p *types.Value) bool {
	if a.V == nil || p.E == nil {
		return true
	}
	return types.Equal(a.V, p.E)
}

func (f *FuncInterpret) matchParam(a *Arg, p *types.Value) bool {
	if f.interpret.IsContract(a.T) {
		return true
	}
	if a.T == types.TypeUnknown && a.V == nil {
		return true
	}
	if p.T == types.TypeUnknown && p.E == nil {
		return true
	}
	if a.T == types.TypeAny && a.V == nil {
		return true
	}
	if l, ok := a.V.(types.List); ok && l.Empty() {
		pl, ok := p.E.(types.List)
		return ok && pl.Empty()
	}
	if a.T != p.T && p.T != types.TypeUnknown {
		canConvert, err := f.interpret.canConvertType(p.T, a.T)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", f.name, err)
			return false
		}
		if !canConvert && p.T != types.TypeUnknown {
			return false
		}
	}
	if p.E == nil {
		// not a real parameter, just a Type binder
		return true
	}
	if a.V == nil {
		// anything of this corresponding type matches
		return true
	}
	return types.Equal(a.V, p.E)
}

func (f *FuncRuntime) cleanup() {
	for _, varname := range f.scopedVars {
		expr := f.vars[varname]
		switch a := expr.E.(type) {
		case types.Ident:
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

func (i *Interpret) matchType(arg types.Type, val types.Type, typeBinds *map[string]types.Type) (result bool, eerroorr error) {
	arg = i.UnaliasType(arg)
	val = i.UnaliasType(val)

	if i.IsContract(arg) {
		if bind, ok := (*typeBinds)[arg.Basic()]; ok && string(bind) != strings.TrimLeft(string(val), ":") {
			return false, nil
		}
		(*typeBinds)[arg.Basic()] = types.Type(strings.TrimLeft(string(val), ":"))
		return true, nil
	}
	if val == types.TypeUnknown || arg == types.TypeUnknown {
		return true, nil
	}
	if (arg == types.TypeFunc && val.Basic() == "func") || (val == types.TypeFunc && arg.Basic() == "func") {
		return true, nil
	}

	parent, err := i.toParent(val, types.Type(arg.Basic()))
	if err != nil {
		return false, err
	}

	aParams := arg.Arguments()
	vParams := parent.Arguments()
	if len(aParams) != len(vParams) {
		return false, nil
	}
	for j, p := range aParams {
		ok, err := i.matchType(types.Type(p), types.Type(vParams[j]), typeBinds)
		if err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}
