package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Interpret struct {
	output      io.Writer
	funcs       map[string]Evaler
	types       map[Type]Type
	typeAliases map[Type]Type
	mainBody    []Expr

	// string->filepath map to control where function was initially defined.
	funcsOrigins map[string]string

	builtinDir string

	intMaker IntMaker

	lambdaCount int

	strictTypes bool

	main *FuncInterpret
}

func NewInterpreter(w io.Writer, builtinDir string) *Interpret {
	i := &Interpret{
		output:       w,
		builtinDir:   builtinDir,
		intMaker:     &Int64Maker{},
		funcsOrigins: make(map[string]string),
	}
	i.funcs = map[string]Evaler{
		"+":             EvalerFunc("+", FPlus, i.AllInts, TypeInt),
		"-":             EvalerFunc("-", FMinus, i.AllInts, TypeInt),
		"*":             EvalerFunc("*", FMultiply, i.AllInts, TypeInt),
		"/":             EvalerFunc("/", FDiv, i.AllInts, TypeInt),
		"mod":           EvalerFunc("mod", FMod, i.TwoInts, TypeInt),
		"<":             EvalerFunc("<", FLess, i.TwoInts, TypeBool),
		"<=":            EvalerFunc("<=", FLessEq, i.TwoInts, TypeBool),
		">":             EvalerFunc(">", FMore, i.TwoInts, TypeBool),
		">=":            EvalerFunc(">=", FMoreEq, i.TwoInts, TypeBool),
		"=":             EvalerFunc("=", FEq, TwoArgs, TypeBool),
		"not":           EvalerFunc("not", FNot, i.OneBoolArg, TypeBool),
		"print":         EvalerFunc("print", i.FPrint, AnyArgs, TypeAny),
		"head":          EvalerFunc("head", FHead, i.ListArg, TypeAny),
		"tail":          EvalerFunc("tail", FTail, i.ListArg, TypeList),
		"append":        EvalerFunc("append", FAppend, i.AppenderArgs, TypeList),
		"list":          EvalerFunc("list", FList, AnyArgs, TypeList),
		"space":         EvalerFunc("space", FSpace, i.StrArg, TypeBool),
		"eol":           EvalerFunc("eol", FEol, i.StrArg, TypeBool),
		"empty":         EvalerFunc("empty", FEmpty, i.ListArg, TypeBool),
		"native.length": EvalerFunc("native.length", i.FLength, i.ListArg, TypeInt),
		"native.nth":    EvalerFunc("native.nth", i.FNth, i.IntAndListArgs, TypeAny),
		"int":           EvalerFunc("int", i.FInt, i.StrArg, TypeInt),
		"open":          EvalerFunc("open", FOpen, i.StrArg, TypeStr),
		"type":          EvalerFunc("type", FType, SingleArg, TypeStr),
	}
	i.types = map[Type]Type{
		TypeUnknown: "",
		TypeAny:     "",
		TypeInt:     TypeAny,
		TypeStr:     TypeList,
		TypeBool:    TypeAny,
		TypeFunc:    TypeAny,
		":list[a]":  TypeAny,
	}
	for c := 'a'; c <= 'z'; c++ {
		i.types[Type(fmt.Sprintf(":%c", c))] = ""
	}
	i.typeAliases = map[Type]Type{
		TypeList: ":list[a]",
	}
	return i
}

func (i *Interpret) UseBigInt(v bool) {
	if v {
		i.intMaker = &BigIntMaker{}
	} else {
		i.intMaker = &Int64Maker{}
	}
}

func (i *Interpret) loadBuiltin(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.lisp"))
	if err != nil {
		return fmt.Errorf("Error while loading builtins: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("Builtin source files not found in %v", dir)
	}
	for _, file := range files {
		err := func() error {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()
			absPath, err := filepath.Abs(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Cannot determine absolute path for %q: %e", file, err)
				absPath = file
			}
			if err := i.parse(absPath, f); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("Error whire loading %v: %w", file, err)
		}
	}
	return nil
}

func (i *Interpret) ParseInt(token string) (Int, bool) {
	return i.intMaker.ParseInt(token)
}

func (i *Interpret) parse(file string, input io.Reader) error {
	parser := NewParser(input, i)
L:
	for {
		expr, err := parser.NextExpr()
		if err == io.EOF {
			break L
		}
		if err != nil {
			return err
		}
		switch a := expr.(type) {
		case *Sexpr:
			if a.Quoted {
				return fmt.Errorf("Unexpected quoted s-expression: %v", a)
			}
			if a.Length() == 0 {
				return fmt.Errorf("Unexpected empty s-expression on top-level: %v", a)
			}
			head, _ := a.Head()
			if name, ok := head.V.(Ident); ok {
				switch name {
				case "func", "def", "func'", "def'":
					memo := false
					if name == "func'" || name == "def'" {
						memo = true
					}
					tail, _ := a.Tail()
					if err := i.defineFunc(file, tail.(*Sexpr), memo); err != nil {
						return err
					}
					continue L
				case "use":
					tail, _ := a.Tail()
					if err := i.use(tail.(*Sexpr).List); err != nil {
						return err
					}
					continue L
				case "deftype":
					tail, _ := a.Tail()
					if err := i.defineType(tail.(*Sexpr).List); err != nil {
						return err
					}
					continue L
				}
			}
		}
		i.mainBody = append(i.mainBody, expr)
	}
	return nil
}

func (i *Interpret) Parse(file string, input io.Reader) error {
	if err := i.parse(file, input); err != nil {
		return err
	}

	i.main = NewFuncInterpret(i, "__main__")
	if err := i.main.AddImpl(QList(Ident("__stdin")), i.mainBody, false, TypeAny); err != nil {
		return err
	}
	return nil
}

// type-checking
func (i *Interpret) Check() []error {
	return i.CheckReturnTypes()

}

func (i *Interpret) Run() error {
	stdin := NewLazyInput(os.Stdin)
	_, err := i.main.Eval([]Param{Param{V: stdin, T: TypeStr}})
	return err
}

// (func-name) args body...
func (i *Interpret) defineFunc(file string, se *Sexpr, memo bool) error {
	if se.Length() < 3 {
		return fmt.Errorf("Not enough arguments for function definition: %v", se)
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("func expected identifier first, found %v", se.List[0])
	}

	fname := string(name)
	if f1, ok := i.funcsOrigins[fname]; ok && f1 != file {
		return fmt.Errorf("cannot define function '%v' in file %v: it is already defined in %v", fname, file, f1)
	}
	var fi *FuncInterpret

	evaler, ok := i.funcs[fname]
	if ok {
		f, ok := evaler.(*FuncInterpret)
		if !ok {
			return fmt.Errorf("Cannot redefine builtin function %v", fname)
		}
		fi = f
	} else {
		fi = NewFuncInterpret(i, fname)
		i.funcs[fname] = fi
	}
	bodyIndex := 2
	returnType := TypeUnknown
	// Check if return type is specified
	if identType, ok := se.List[2].(Ident); ok {
		returnType, ok = ParseType(string(identType))
		if ok {
			if _, err := i.parseType(string(identType)); err != nil {
				return fmt.Errorf("%v: %v", fname, err)
			}
			bodyIndex++
		}
	}
	// TODO
	if err := fi.AddImpl(se.List[1], se.List[2:], memo, returnType); err != nil {
		return err
	}
	i.funcsOrigins[fname] = file
	return nil
}

func (i *Interpret) use(args []Expr) error {
	if len(args) != 1 {
		return fmt.Errorf("'use' expected one argument, found: %v", args)
	}
	module := args[0]
	switch a := module.(type) {
	case Str:
		f, err := os.Open(string(a))
		if err != nil {
			return err
		}
		defer f.Close()
		fpath, err := filepath.Abs(string(a))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot detect absolute path for %v: %v\n", string(a), err)
			fpath = string(a)
		}
		return i.parse(fpath, f)
	case Ident:
		switch string(a) {
		case "bigmath":
			i.UseBigInt(true)
		case "std":
			if err := i.loadBuiltin(i.builtinDir); err != nil {
				return err
			}
		case "strict":
			i.strictTypes = true
		default:
			return fmt.Errorf("Unknown use-directive: %v", string(a))
		}
		return nil
	}
	return fmt.Errorf("Unexpected argument type to 'use': %v (%T)", module, module)
}

// (new-type) (old-type)
func (in *Interpret) defineType(args []Expr) error {
	if len(args) != 2 {
		return fmt.Errorf("'deftype' expected two arguments, found: %v", args)
	}
	newId, ok := args[0].(Ident)
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}
	newType, ok := ParseType(string(newId))
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}

	if _, ok := in.types[newType]; ok {
		return fmt.Errorf("Cannot redefine type %v", newType)
	}

	oldId, ok := args[1].(Ident)
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[1])
	}
	oldType, ok := ParseType(string(oldId))
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}
	if alias, ok := in.typeAliases[oldType]; ok {
		oldType = alias
	}
	if _, ok := in.types[oldType]; !ok {
		return fmt.Errorf("Basic type does not exist: %v", oldType)
	}
	in.types[newType] = oldType
	return nil
}

func (in *Interpret) canConvertType(from, to Type) (bool, error) {
	from = from.Canonical()
	to = to.Canonical()
	if alias, ok := in.typeAliases[to]; ok {
		to = alias
	}
	if alias, ok := in.typeAliases[from]; ok {
		from = alias
	}
	if _, ok := in.types[to.Canonical()]; !ok {
		return false, fmt.Errorf("Cannot convert type %v into %v: %v is not defined", from, to, to)
	}
	if from == to {
		return true, nil
	}
	for {
		if alias, ok := in.typeAliases[from]; ok {
			from = alias
		}
		if from == to {
			return true, nil
		}
		parent, ok := in.types[from]
		if !ok {
			return false, fmt.Errorf("Cannot convert type %v into %v: %v is not defined", from, to, from)
		}
		if parent == to {
			return true, nil
		}
		if parent == "" {
			return false, nil
		}
		from = parent
	}
	panic("Unreachable return")
}

func (in *Interpret) FPrint(args []Param) (*Param, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		e.V.Print(in.output)
	}
	fmt.Fprintf(in.output, "\n")
	return &Param{V: QEmpty, T: TypeList}, nil
}

func (in *Interpret) fakeFunc(args []Expr) (Expr, error) {
	panic("The fakeFunc should not be called")
}

// convert string into int
func (in *Interpret) FInt(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FInt: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FInt: expected argument to be Str, found %v", args)
	}
	i, ok := in.intMaker.ParseInt(string(s))
	if !ok {
		return nil, fmt.Errorf("FInt: cannot convert argument into Int: %v", s)
	}
	return &Param{V: i, T: TypeInt}, nil
}

func (in *Interpret) FLength(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FLength: expected exaclty one argument, found %v", args)
	}

	if a, ok := args[0].V.(Lenghter); ok {
		return &Param{V: in.intMaker.MakeInt(int64(a.Length())), T: TypeInt}, nil
	}

	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FLength: expected argument to be List, found %v", args[0])
	}
	var l int64
	for !a.Empty() {
		l++
		var err error
		a, err = a.Tail()
		if err != nil {
			return nil, err
		}
	}
	return &Param{V: in.intMaker.MakeInt(l), T: TypeInt}, nil
}

func (in *Interpret) FNth(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FNth: expected exaclty two arguments, found %v", args)
	}
	bign, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FNth: expected first argument to be Int, fount: %v", args[0])
	}
	n := int(bign.Int64())
	if nther, ok := args[1].V.(Nther); ok {
		return nther.Nth(n)
	}
	a, ok := args[1].V.(List)
	if !ok {
		return nil, fmt.Errorf("FNth: expected second argument to be List, found %v", args[1])
	}
	if n <= 0 {
		return nil, fmt.Errorf("FNth: n should be >= 1, found: %v", n)
	}
	// numeration starts with 1
	for n > 1 {
		var err error
		a, err = a.Tail()
		if err != nil {
			return nil, err
		}
		n--
	}
	return a.Head()
}

func (in *Interpret) NewLambdaName() (name string) {
	name = fmt.Sprintf("__lambda__%03d", in.lambdaCount)
	in.lambdaCount++
	return
}

func (in *Interpret) DeleteLambda(name string) {
	if !strings.HasPrefix(name, "__lambda__") {
		return
	}
	if _, ok := in.funcs[name]; ok {
		delete(in.funcs, name)
	}
}

func (in *Interpret) Stat() {
	fmt.Fprintf(os.Stderr, "Functions:\n")
	for fname, _ := range in.funcs {
		fmt.Fprintf(os.Stderr, "%v\n", fname)
	}
}

func (i *Interpret) CheckReturnTypes() (errs []error) {
	mainArgs := map[string]Type{
		"__stdin": TypeStr,
	}
	_, err := i.evalBodyType("__main__", i.mainBody, mainArgs)
	if err != nil {
		errs = append(errs, err)
	}
	for _, fn := range i.funcs {
		fi, ok := fn.(*FuncInterpret)
		if !ok {
			// native function
			continue
		}

		for _, impl := range fi.bodies {
			if i.strictTypes {
				if fi.returnType == TypeUnknown {
					err := fmt.Errorf("%v : %v: return type should be specified in strict mode", i.funcsOrigins[fi.name], fi.name)
					errs = append(errs, err)
				}
				if impl.argfmt.Wildcard == "" {
					for _, a := range impl.argfmt.Args {
						if a.T == TypeUnknown {
							err := fmt.Errorf("%v : %v: arument type should be specified in strict mode: %v", i.funcsOrigins[fi.name], fi.name, a.Name)
							errs = append(errs, err)
						}
					}
				}
			}
			t, err := i.evalBodyType(fi.name, impl.body, impl.argfmt.Values())
			if err != nil {
				errs = append(errs, err)
			}
			if fi.returnType != TypeAny && fi.returnType != TypeUnknown {
				if t != fi.returnType {
					err := fmt.Errorf("Incorrect return value in function %v(%v): expected %v actual %v", fi.name, impl.argfmt, fi.returnType, t)
					errs = append(errs, err)
				}
			}
		}
	}
	return
}

func (in *Interpret) evalBodyType(fname string, body []Expr, vars map[string]Type) (rt Type, err error) {
	if len(body) == 0 {
		// This should be possible only for __main__ function
		return TypeAny, err
	}

	u := TypeUnknown
L:
	for i, stt := range body[:len(body)-1] {
		_ = i
		switch a := stt.(type) {
		case Int, Str, Bool, Ident:
			continue L
		case *Sexpr:
			if a.Quoted || a.Empty() {
				continue L
			}
			ident, ok := a.List[0].(Ident)
			if !ok {
				return u, fmt.Errorf("Expected ident, found: %v", a.List[0])
			}
			switch name := string(ident); name {
			case "set", "set'":
				if i == len(body)-1 {
					return u, fmt.Errorf("Unexpected %v statement at the end of the function", name)
				}
				varname, ok := a.List[1].(Ident)
				if !ok {
					return u, fmt.Errorf("%v: second argument should be variable name, found: %v", name, a.List[1])
				}
				if len(a.List) == 4 {
					id, ok := a.List[3].(Ident)
					if !ok {
						return u, fmt.Errorf("Fourth statement of %v should be type identifier, found: %v", name, a.List[3])
					}
					tp, err := in.parseType(string(id))
					if err != nil {
						return u, fmt.Errorf("Fourth statement of %v should be type identifier, found: %v (%v)", name, a.List[3], err)
					}
					vars[string(varname)] = tp
				} else if len(a.List) == 3 {
					tp, err := in.exprType(fname, a.List[2], vars)
					if err != nil {
						return u, err
					}
					vars[string(varname)] = tp
				} else {
					return u, fmt.Errorf("%v: incorrect number of arguments %v: %v", fname, name, a.List)
				}
			case "print":
				for i, arg := range a.List[1:] {
					_, err := in.exprType(fname, arg, vars)
					if err != nil {
						return u, fmt.Errorf("%v: incorrect argument to print at posision %v: %v", fname, i, err)
					}
				}
			default:
				if _, err := in.exprType(fname, a, vars); err != nil {
					return u, fmt.Errorf("%v: %v", fname, err)
				}
			}
		}
	}
	return in.exprType(fname, body[len(body)-1], vars)
}

func (i *Interpret) exprType(fname string, e Expr, vars map[string]Type) (result Type, err error) {
	const u = TypeUnknown
	switch a := e.(type) {
	case Int:
		return TypeInt, nil
	case Str:
		return TypeStr, nil
	case Bool:
		return TypeBool, nil
	case Ident:
		// fmt.Fprintf(os.Stderr, "exprType: %v\n", a)
		if t, ok := vars[string(a)]; ok {
			return t, nil
		} else if _, ok := i.funcs[string(a)]; ok {
			return TypeFunc, nil
		} else if _, err := i.parseType(string(a)); err == nil {
			return Type(a), nil
		}
		return u, fmt.Errorf("Undefined variable: %v", string(a))
	case *Sexpr:
		if a.Quoted || a.Empty() {
			return TypeList, nil
		}
		if a.Lambda {
			return TypeFunc, nil
		}
		ident, ok := a.List[0].(Ident)
		if !ok {
			return u, fmt.Errorf("%v: expected ident, found: %v", fname, a.List[0])
		}
		switch name := string(ident); name {
		case "set", "set'":
			return u, fmt.Errorf("%v: unexpected %v and the end of function", fname, ident)
		case "lambda":
			return TypeFunc, nil
		case "and", "or":
			return TypeBool, nil
		case "gen", "gen'":
			return TypeList, nil
		case "apply":
			if len(a.List) != 3 {
				return u, fmt.Errorf("%v: incorrect number of arguments to 'apply': %v", fname, a.List)
			}
			ftype, err := i.exprType(fname, a.List[1], vars)
			if err != nil {
				return u, err
			}
			if ftype != TypeFunc {
				return u, fmt.Errorf("%v: apply expects function on first place, found: %v", fname, a.List[1])
			}
			atype, err := i.exprType(fname, a.List[2], vars)
			if err != nil {
				return u, err
			}
			if atype != TypeList && atype != TypeUnknown {
				return u, fmt.Errorf("%v: apply expects list on second place, found: %v", fname, a.List[2])
			}
			fi, ok := i.funcs[string(a.List[1].(Ident))]
			if !ok {
				return u, fmt.Errorf("%v: unknown function supplied to apply: %v", fname, a.List[1])
			}
			return fi.ReturnType(), nil
		case "if":
			if len(a.List) != 4 {
				return u, fmt.Errorf("%v: incorrect number of arguments to 'if': %v", fname, a.List)
			}
			condType, err := i.exprType(fname, a.List[1], vars)
			if err != nil {
				return u, err
			}
			if condType != TypeBool && condType != TypeUnknown {
				return u, fmt.Errorf("%v: condition in if-statement should return :bool, found: %v", fname, condType)
			}
			t1, err := i.exprType(fname, a.List[2], vars)
			if err != nil {
				return u, err
			}
			t2, err := i.exprType(fname, a.List[3], vars)
			if err != nil {
				return u, err
			}
			if t1 == TypeUnknown || t2 == TypeUnknown {
				return TypeUnknown, nil
			}
			if t1 != t2 {
				return TypeAny, nil
			}
			return t1, nil
		case "do":
			return i.evalBodyType(fname, a.List[1:], vars)
		default:
			// this is a function call
			if tvar, ok := vars[name]; ok {
				if tvar == TypeFunc || tvar == TypeUnknown {
					return TypeUnknown, nil
				}
				return u, fmt.Errorf("%v: expected '%v' to be function, found: %v", fname, name, tvar)
			}
			f, ok := i.funcs[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "%v: cannot detect return type of function %v\n", fname, name)
				return TypeAny, nil
			}

			// check if we have matching func impl
			params := []Param{}
			for _, item := range a.List[1:] {
				switch a := item.(type) {
				case Int:
					params = append(params, Param{T: TypeInt, V: item})
				case Str:
					params = append(params, Param{T: TypeStr, V: item})
				case Bool:
					params = append(params, Param{T: TypeBool, V: item})
				case *Sexpr:
					if a.Empty() || a.Quoted {
						params = append(params, Param{T: TypeList, V: a})
					} else if a.Lambda {
						params = append(params, Param{T: TypeFunc})
					} else {
						itemType, err := i.exprType(fname, item, vars)
						if err != nil {
							return u, err
						}
						params = append(params, Param{T: itemType})
					}
				case Ident:
					itemType, err := i.exprType(fname, item, vars)
					if err != nil {
						return u, err
					}
					params = append(params, Param{T: itemType})
				default:
					panic(fmt.Errorf("%v: unexpected type: %v", fname, item))
				}
			}
			_, t, err := f.TryBind(params)
			if err != nil {
				return u, fmt.Errorf("%v: %v", fname, err)
			}

			return t, nil
		}
	}
	fmt.Fprintf(os.Stderr, "Unexpected return. (TypeAny)\n")
	return TypeAny, nil
}

func (in *Interpret) parseType(token string) (Type, error) {
	if alias, ok := in.typeAliases[Type(token)]; ok {
		token = string(alias)
	}
	_, ok := in.types[Type(token)]
	if !ok {
		return "", fmt.Errorf("Cannot parse type %v: not defined", token)
	}
	return Type(token), nil
}

func (in *Interpret) toParent(from, parent Type) (Type, error) {
	binds := map[string]string{}
	for i, p := range from.Arguments() {
		if len(p) == 1 {
			// generic
			continue
		}
		binds[string('a'+i)] = p
	}
	f := from.Canonical()
	for {
		if f == "" {
			return TypeUnknown, fmt.Errorf("Cannot convert %v into %v", from, parent)
		}
		if f.Basic() == parent.Basic() {
			parent = f
			break
		}
		if alias, ok := in.typeAliases[f]; ok {
			f = alias
		}
		if f.Basic() == parent.Basic() {
			parent = f
			break
		}
		par, ok := in.types[f]
		if !ok {
			return TypeUnknown, fmt.Errorf("Cannot convert type %v into %v: %v is not defined", from, parent, f)
		}
		f = par
	}
	res := ":" + parent.Basic()
	if len(parent.Arguments()) > 0 {
		res += "["
		for j, a := range parent.Arguments() {
			if j > 0 {
				res += ","
			}
			// a := string('a' + j)
			if b, ok := binds[a]; ok {
				res += b
			} else {
				res += a
			}
		}
		res += "]"
	}

	return Type(res), nil
}
