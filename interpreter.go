package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"regexp"
	"strings"

	"github.com/avoronkov/spil/library"
	"github.com/avoronkov/spil/types"
)

type Interpret struct {
	output      io.Writer
	funcs       map[string]types.Function
	types       map[types.Type]types.Type
	typeAliases map[types.Type]types.Type
	contracts   map[types.Type]struct{}
	mainBody    []types.Value

	// string->filepath map to control where function was initially defined.
	funcsOrigins map[string]string

	PluginDir   string
	IncludeDirs []string

	intMaker   types.IntMaker
	floatMaker types.FloatMaker

	lambdaCount int

	strictTypes bool

	main *FuncInterpret
}

func NewInterpreter(w io.Writer) *Interpret {
	i := &Interpret{
		output:       w,
		intMaker:     &types.Int64Maker{},
		floatMaker:   &types.Float64Maker{},
		funcsOrigins: make(map[string]string),
		contracts:    make(map[types.Type]struct{}),
	}
	i.funcs = map[string]types.Function{
		"int.plus":      EvalerFunc("+", FPlus, AnyArgs, types.TypeInt),
		"int.minus":     EvalerFunc("-", FMinus, AnyArgs, types.TypeInt),
		"int.mult":      EvalerFunc("*", FMultiply, AnyArgs, types.TypeInt),
		"int.div":       EvalerFunc("/", FDiv, AnyArgs, types.TypeInt),
		"mod":           EvalerFunc("mod", FMod, i.TwoInts, types.TypeInt),
		"float.plus":    EvalerFunc("+", FloatPlus, AnyArgs, types.TypeFloat),
		"float.minus":   EvalerFunc("-", FloatMinus, AnyArgs, types.TypeFloat),
		"float.mult":    EvalerFunc("*", FloatMult, AnyArgs, types.TypeFloat),
		"float.div":     EvalerFunc("/", FloatDiv, AnyArgs, types.TypeFloat),
		"math.sin":      EvalerFunc("math.sin", FSin, AnyArgs, types.TypeFloat),
		"math.cos":      EvalerFunc("math.cos", FCos, AnyArgs, types.TypeFloat),
		"int.less":      EvalerFunc("int.less", FIntLess, i.TwoInts, types.TypeBool),
		"str.less":      EvalerFunc("str.less", FStrLess, i.TwoStrs, types.TypeBool),
		"float.less":    EvalerFunc("float.less", FFloatLess, AnyArgs, types.TypeBool),
		"=":             EvalerFunc("=", FEq, TwoArgs, types.TypeBool),
		"not":           EvalerFunc("not", FNot, i.OneBoolArg, types.TypeBool),
		"print":         EvalerFunc("print", i.FPrint, AnyArgs, types.TypeAny),
		"native.head":   EvalerFunc("native.head", FHead, AnyArgs, types.TypeAny),
		"native.tail":   EvalerFunc("native.tail", FTail, AnyArgs, types.TypeList),
		"append":        EvalerFunc("append", FAppend, i.AppenderArgs, types.TypeList),
		"list":          EvalerFunc("list", FList, AnyArgs, types.TypeList),
		"space":         EvalerFunc("space", FSpace, i.StrArg, types.TypeBool),
		"eol":           EvalerFunc("eol", FEol, i.StrArg, types.TypeBool),
		"empty":         EvalerFunc("empty", FEmpty, i.ListArg, types.TypeBool),
		"native.length": EvalerFunc("native.length", i.FLength, i.ListArg, types.TypeInt),
		"native.nth":    EvalerFunc("native.nth", i.FNth, i.IntAndListArgs, types.TypeAny),
		"strtoint":      EvalerFunc("strtoint", i.FStrToInt, i.StrArg, types.TypeInt),
		"floattoint":    EvalerFunc("floattoint", i.FFloatToInt, AnyArgs, types.TypeInt),
		"strtofloat":    EvalerFunc("strtofloat", FStrToFloat, AnyArgs, types.TypeFloat),
		"inttofloat":    EvalerFunc("inttofloat", FIntToFloat, AnyArgs, types.TypeFloat),
		"open":          EvalerFunc("open", FOpen, i.StrArg, types.TypeStr),
		"type":          EvalerFunc("type", FType, SingleArg, types.TypeStr),
		"parse":         EvalerFunc("parse", i.FParse, i.StrArg, types.TypeList),
	}
	i.types = map[types.Type]types.Type{
		types.TypeUnknown: "",
		types.TypeAny:     "",
		types.TypeInt:     types.TypeAny,
		types.TypeFloat:   types.TypeAny,
		types.TypeStr:     "list[str]",
		types.TypeBool:    types.TypeAny,
		types.TypeFunc:    types.TypeAny,
		"list[a]":         types.TypeAny,
		"args[a]":         "list[a]",
	}
	i.typeAliases = map[types.Type]types.Type{
		types.TypeList: "list[any]",
	}
	return i
}

func (i *Interpret) UseBigInt(v bool) {
	if v {
		i.intMaker = &types.BigIntMaker{}
	} else {
		i.intMaker = &types.Int64Maker{}
	}
}

func (i *Interpret) loadLibrary(name string) error {
	foundFiles := false
	prefix := fmt.Sprintf("library/%s/", name)
	for _, file := range library.AssetNames() {
		if !strings.HasPrefix(file, prefix) {
			continue
		}
		err := func() error {
			data := library.MustAsset(file)
			if err := i.parse(file, bytes.NewReader(data)); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("Error whire loading %v: %w", file, err)
		}
		foundFiles = true
	}
	if !foundFiles {
		return fmt.Errorf("Library source files not found: %v", name)
	}
	return nil
}

func (i *Interpret) ParseInt(token string) (types.Int, bool) {
	return i.intMaker.ParseInt(token)
}

func (i *Interpret) ParseFloat(token string) (types.Float, bool) {
	return i.floatMaker.ParseFloat(token)
}

func (i *Interpret) parse(file string, input io.Reader) error {
	parser := NewParser(input, i)
L:
	for {
		val, err := parser.NextExpr(false)
		if err == io.EOF {
			break L
		}
		if err != nil {
			return err
		}
		switch a := val.E.(type) {
		case *types.Sexpr:
			if a.Quoted {
				return fmt.Errorf("Unexpected quoted s-expression: %v", a)
			}
			if a.Length() == 0 {
				return fmt.Errorf("Unexpected empty s-expression on top-level: %v", a)
			}
			head, _ := a.Head()
			if name, ok := head.E.(types.Ident); ok {
				switch name {
				case "func", "def", "func'", "def'":
					memo := false
					if name == "func'" || name == "def'" {
						memo = true
					}
					tail, _ := a.Tail()
					if err := i.defineFunc(file, tail.(*types.Sexpr), memo); err != nil {
						return err
					}
					continue L
				case "use":
					tail, _ := a.Tail()
					if err := i.use(file, tail.(*types.Sexpr).List); err != nil {
						return err
					}
					continue L
				case "deftype":
					tail, _ := a.Tail()
					if err := i.defineType(tail.(*types.Sexpr).List); err != nil {
						return err
					}
					continue L
				case "contract":
					tail, _ := a.Tail()
					if err := i.defineContract(tail.(*types.Sexpr).List); err != nil {
						return err
					}
					continue L
				}
			}
		}
		i.mainBody = append(i.mainBody, *val)
	}
	return nil
}

func (i *Interpret) Parse(file string, input io.Reader) error {
	if err := i.loadLibrary("builtin"); err != nil {
		return err
	}

	if err := i.parse(file, input); err != nil {
		return err
	}

	i.main = NewFuncInterpret(i, "__main__")
	if err := i.main.AddImpl(types.Ident("__main_args"), i.mainBody, false, types.TypeAny); err != nil {
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
	i.main.capturedVars["__stdin"] = &types.Value{E: stdin, T: types.TypeStr}
	params := []types.Value{}
	fargs := flag.Args()
	if len(fargs) > 0 {
		for _, arg := range fargs[1:] {
			params = append(params, types.Value{E: types.Str(arg), T: types.TypeStr})
		}
	}
	_, err := i.main.Eval(params)
	return err
}

// (func-name) args body...
func (i *Interpret) defineFunc(file string, se *types.Sexpr, memo bool) error {
	if se.Length() < 3 {
		return fmt.Errorf("Not enough arguments for function definition: %v", se)
	}
	name, ok := se.List[0].E.(types.Ident)
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
	returnType := types.TypeUnknown
	// Check if return type is specified
	if identType, ok := se.List[2].E.(types.Ident); ok {
		returnType, ok = types.ParseType(string(identType))
		if ok {
			if _, err := i.parseType(string(identType)); err != nil {
				return fmt.Errorf("%v: %v", fname, err)
			}
			bodyIndex++
		}
	}
	// TODO
	if err := fi.AddImpl(se.List[1].E, se.List[2:], memo, returnType); err != nil {
		return err
	}
	i.funcsOrigins[fname] = file
	return nil
}

func (i *Interpret) use(file string, args []types.Value) error {
	if len(args) < 1 {
		return fmt.Errorf("'use' expects arguments, none found.")
	}
	module := args[0]
	switch a := module.E.(type) {
	case types.Str:
		return i.useModule(string(a))
	case types.Ident:
		switch string(a) {
		case "bigmath":
			i.UseBigInt(true)
		case "std":
			if err := i.loadLibrary("std"); err != nil {
				return err
			}
		case "strict":
			i.strictTypes = true
		case "plugin":
			if len(args) < 2 {
				return fmt.Errorf("'use plugin' expects argument, none found")
			}
			name, ok := args[1].E.(types.Str)
			if !ok {
				return fmt.Errorf("'use plugin' expects string argument, found: %v", args[2])
			}
			return i.usePlugin(file, string(name))
		default:
			return fmt.Errorf("Unexpected argument to 'use': %v", string(a))
		}
		return nil
	}
	return fmt.Errorf("Unexpected argument type to 'use': %v (%T)", module, module)
}

func (in *Interpret) useModule(name string) error {
	includeDirs := append([]string{"."}, in.IncludeDirs...)
	for _, d := range includeDirs {
		filename := filepath.Join(d, name)
		f, err := os.Open(filename)
		if os.IsNotExist(err) {
			continue
		}
		defer f.Close()
		fpath, err := filepath.Abs(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot detect absolute path for %v: %v\n", filename, err)
			fpath = filename
		}
		return in.parse(fpath, f)
	}
	return fmt.Errorf("Module %v not found in %v", name, includeDirs)
}

func (in *Interpret) usePlugin(file, name string) (err error) {
	var (
		plug     *plugin.Plugin
		filename string
	)
	fdir := filepath.Dir(file)
	for _, dir := range []string{fdir, in.PluginDir} {
		// search "someplug" in "$dir/someplug/someplug.so"
		filename = filepath.Join(dir, name, name+".so")
		plug, err = plugin.Open(filename)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		break
	}
	if plug == nil {
		return fmt.Errorf("Plugin '%v' not found in directories: %v, %v", name, fdir, in.PluginDir)
	}

	sym, err := plug.Lookup("Types")
	if err != nil {
		return fmt.Errorf("Plugin %v does not define Types", filename)
	}
	tps := sym.(*map[types.Type]types.Type)
	for k, v := range *tps {
		if _, exist := in.types[k]; exist {
			return fmt.Errorf("Cannot redefine type %v from plugin %v: type already exist", k, filename)
		}
		in.types[k] = v
	}

	sym, err = plug.Lookup("Funcs")
	if err != nil {
		return fmt.Errorf("Plugin %v does not define Types", filename)
	}
	fncs := sym.(*map[string]types.Function)
	for k, v := range *fncs {
		if _, exist := in.funcs[k]; exist {
			return fmt.Errorf("Cannot redefine function %v from plugin %v: function already exist", k, filename)
		}
		in.funcs[k] = v
	}
	return nil
}

// (new-type) (old-type)
func (in *Interpret) defineType(args []types.Value) error {
	if len(args) != 2 {
		return fmt.Errorf("'deftype' expected two arguments, found: %v", args)
	}
	newId, ok := args[0].E.(types.Ident)
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}
	newType, ok := types.ParseType(string(newId))
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}

	if _, ok := in.types[newType]; ok {
		return fmt.Errorf("Cannot redefine type %v", newType)
	}

	oldId, ok := args[1].E.(types.Ident)
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[1])
	}
	oldType, ok := types.ParseType(string(oldId))
	if !ok {
		return fmt.Errorf("deftype expects first argument to be new type, found: %v", args[0])
	}
	oldType = in.UnaliasType(oldType)
	if _, ok := in.types[oldType.Canonical()]; !ok {
		return fmt.Errorf("Basic type does not exist: %v", oldType)
	}
	in.types[newType] = oldType
	return nil
}

// (contract (:a :b :c)
//   (fn1 ...) :return
//   (fn2 ...) :return
//   ...)
func (in *Interpret) defineContract(args []types.Value) error {
	if len(args) < 1 {
		return fmt.Errorf("Not enougn arguments to contract: %v", args)
	}
	switch cs := args[0].E.(type) {
	case types.Ident:
		t, ok := types.ParseType(string(cs))
		if !ok {
			return fmt.Errorf("Contract expect first argument to be type, found: %v", cs)
		}
		if _, ok := in.types[t]; ok {
			return fmt.Errorf("Cannot define contract %v: type already exist", string(cs))
		}
		in.types[t] = ""
		in.contracts[t] = struct{}{}
	}
	// TODO: implement contract-functions
	return nil
}

func (in *Interpret) canConvertType(from, to types.Type) (bool, error) {
	from = in.UnaliasType(from.Canonical())
	to = in.UnaliasType(to.Canonical())

	if from == types.TypeUnknown || to == types.TypeUnknown {
		return true, nil
	}

	if _, ok := in.types[to.Canonical()]; !ok {
		return false, fmt.Errorf("Cannot convert type %v into %v: %v is not defined", from, to, to)
	}
	if from == to {
		return true, nil
	}
	for {
		from = in.UnaliasType(from)
		if from == to {
			return true, nil
		}
		parent, ok := in.types[from.Canonical()]
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

func (in *Interpret) FPrint(args []types.Value) (*types.Value, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		e.E.Print(in.output)
	}
	fmt.Fprintf(in.output, "\n")
	return &types.Value{E: types.QEmpty, T: types.TypeList}, nil
}

// convert string into int
func (in *Interpret) FStrToInt(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FInt: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FInt: expected argument to be Str, found %v", args)
	}
	i, ok := in.intMaker.ParseInt(string(s))
	if !ok {
		return nil, fmt.Errorf("FInt: cannot convert argument into Int: %v", s)
	}
	return &types.Value{E: i, T: types.TypeInt}, nil
}

// convert float into int
func (in *Interpret) FFloatToInt(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FInt: expected exaclty one argument, found %v", args)
	}
	f, ok := args[0].E.(types.Float64)
	if !ok {
		return nil, fmt.Errorf("FInt: expected argument to be Float, found %v", args)
	}
	i := in.intMaker.MakeInt(int64(f))
	return &types.Value{E: i, T: types.TypeInt}, nil
}

func (in *Interpret) FLength(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FLength: expected exaclty one argument, found %v", args)
	}

	if a, ok := args[0].E.(Lenghter); ok {
		return &types.Value{E: in.intMaker.MakeInt(int64(a.Length())), T: types.TypeInt}, nil
	}

	a, ok := args[0].E.(types.List)
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
	return &types.Value{E: in.intMaker.MakeInt(l), T: types.TypeInt}, nil
}

func (in *Interpret) FNth(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FNth: expected exaclty two arguments, found %v", args)
	}
	bign, ok := args[0].E.(types.Int)
	if !ok {
		return nil, fmt.Errorf("FNth: expected first argument to be Int, fount: %v", args[0])
	}
	n := int(bign.Int64())
	if nther, ok := args[1].E.(Nther); ok {
		return nther.Nth(n)
	}
	a, ok := args[1].E.(types.List)
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

func (in *Interpret) FParse(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FParse: expected exaclty one argument, found %v", args)
	}
	var input io.Reader
	s, ok := args[0].E.(types.Str)
	if ok {
		input = strings.NewReader(string(s))
	} else if li, ok := args[0].E.(*LazyInput); ok {
		// TODO make non-destructing reading from lazy input
		input = li.input
	} else {
		return nil, fmt.Errorf("FParse: expected string argument, found %v", args)
	}
	parser := NewParser(input, in)
	list := &types.Sexpr{Quoted: true}
	for {
		val, err := parser.NextExpr(true)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		list.List = append(list.List, *val)
	}
	return &types.Value{T: types.TypeList, E: list}, nil
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
	mainArgs := map[string]types.Type{
		"__stdin": types.TypeStr,
		"__args":  types.Type("list[str]"),
	}
	for i := 1; i <= 9; i++ {
		mainArgs[fmt.Sprintf("_%d", i)] = types.TypeStr
	}
	_, err := i.evalBodyType("__main__", i.mainBody, mainArgs, nil)
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
				if impl.returnType == types.TypeUnknown {
					err := fmt.Errorf("%v : %v: return type should be specified in strict mode", i.funcsOrigins[fi.name], fi.name)
					errs = append(errs, err)
				}
				if impl.argfmt.Wildcard == "" {
					for _, a := range impl.argfmt.Args {
						if a.T == types.TypeUnknown {
							err := fmt.Errorf("%v : %v: arument type should be specified in strict mode: %v", i.funcsOrigins[fi.name], fi.name, a.Name)
							errs = append(errs, err)
						}
					}
				}
			}
			t, err := i.evalBodyType(fi.name, impl.body, impl.argfmt.Values(), nil)
			if err != nil {
				errs = append(errs, err)
			}
			if impl.returnType != types.TypeUnknown && !i.IsGeneric(impl.returnType) {
				if ok, err := i.canConvertType(t, impl.returnType); !ok || err != nil {
					err := fmt.Errorf("Incorrect return value in function %v(%v): expected %v actual %v (%v)", fi.name, impl.argfmt, impl.returnType, t, err)
					errs = append(errs, err)
				}
			}
		}
	}
	return
}

func (in *Interpret) evalBodyType(fname string, body []types.Value, vars map[string]types.Type, tps map[string]types.Type) (rt types.Type, err error) {
	if len(body) == 0 {
		// This should be possible only for __main__ function
		return types.TypeAny, err
	}

	u := types.TypeUnknown
L:
	for i, stt := range body[:len(body)-1] {
		_ = i
		switch a := stt.E.(type) {
		case types.Int, types.Str, types.Bool, types.Ident:
			continue L
		case *types.Sexpr:
			if a.Quoted || a.Empty() {
				continue L
			}
			ident, ok := a.List[0].E.(types.Ident)
			if !ok {
				return u, fmt.Errorf("Expected ident, found: %v", a.List[0])
			}
			switch name := string(ident); name {
			case "set", "set'":
				if i == len(body)-1 {
					return u, fmt.Errorf("Unexpected %v statement at the end of the function", name)
				}
				varname, ok := a.List[1].E.(types.Ident)
				if !ok {
					return u, fmt.Errorf("%v: second argument should be variable name, found: %v", name, a.List[1])
				}
				if len(a.List) == 4 {
					id, ok := a.List[3].E.(types.Ident)
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
				if _, err := in.exprType(fname, stt, vars); err != nil {
					return u, fmt.Errorf("%v: %v", fname, err)
				}
			}
		}
	}
	rt, err = in.exprType(fname, body[len(body)-1], vars)
	if err != nil {
		return u, err
	}
	rt = in.UnaliasType(rt)
	return rt.Expand(tps), nil
}

type ReturnTyper interface {
	ReturnType() types.Type
}

type Binder interface {
	TryBindAll([]types.Value) (types.Type, error)
}

var reArg = regexp.MustCompile(`^_[0-9]+$`)

func (i *Interpret) exprType(fname string, e types.Value, vars map[string]types.Type) (result types.Type, err error) {
	const u = types.TypeUnknown
	switch a := e.E.(type) {
	case types.Int, types.Float, types.Str, types.Bool:
		return e.T, nil
	case types.Ident:
		if t, ok := vars[string(a)]; ok {
			return t, nil
		} else if fe, ok := i.funcs[string(a)]; ok {
			if fu, ok := fe.(*FuncInterpret); ok {
				return fu.FuncType(), nil
			}
			return types.TypeFunc, nil
		} else if t, err := i.parseType(string(a)); err == nil {
			return t, nil
		}
		if string(a) == "__args" || reArg.MatchString(string(a)) {
			return types.TypeAny, nil
		}
		return u, fmt.Errorf("Undefined variable: %v", string(a))
	case *types.Sexpr:
		if a.Quoted || a.Empty() {
			return types.TypeList, nil
		}
		if a.Lambda {
			return types.TypeFunc, nil
		}
		ident, ok := a.List[0].E.(types.Ident)
		if !ok {
			return u, fmt.Errorf("%v: expected ident, found: %v", fname, a.List[0])
		}
		switch name := string(ident); name {
		case "set", "set'":
			return u, fmt.Errorf("%v: unexpected %v and the end of function", fname, ident)
		case "lambda":
			return types.TypeFunc, nil
		case "and", "or":
			return types.TypeBool, nil
		case "gen", "gen'":
			return types.TypeList, nil
		case "apply":
			if len(a.List) != 3 {
				return u, fmt.Errorf("%v: incorrect number of arguments to 'apply': %v", fname, a.List)
			}
			ftype, err := i.exprType(fname, a.List[1], vars)
			if err != nil {
				return u, err
			}
			if ftype.Basic() != "func" {
				return u, fmt.Errorf("%v: apply expects function on first place, found: %v", fname, a.List[1])
			}
			atype, err := i.exprType(fname, a.List[2], vars)
			if err != nil {
				return u, err
			}
			if atype.Basic() != "list" && atype.Basic() != "args" && atype != types.TypeUnknown {
				return u, fmt.Errorf("%v: apply expects list on second place, found: %v (%v)", fname, a.List[2], atype)
			}
			fi, ok := i.funcs[string(a.List[1].E.(types.Ident))]
			if !ok {
				return u, fmt.Errorf("%v: unknown function supplied to apply: %v", fname, a.List[1])
			}
			if returnTyper, ok := fi.(ReturnTyper); ok {
				return returnTyper.ReturnType(), nil
			}
			return u, nil
		case "if":
			if len(a.List) != 4 {
				return u, fmt.Errorf("%v: incorrect number of arguments to 'if': %v", fname, a.List)
			}
			condType, err := i.exprType(fname, a.List[1], vars)
			if err != nil {
				return u, err
			}
			if condType != types.TypeBool && condType != types.TypeUnknown {
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
			if t1 == types.TypeUnknown || t2 == types.TypeUnknown {
				return types.TypeUnknown, nil
			}
			t1 = i.UnaliasType(t1)
			t2 = i.UnaliasType(t2)
			if t1 != t2 {
				return types.TypeAny, nil
			}
			return t1, nil
		case "do":

			res, err := i.evalBodyType(fname, a.List[1:], vars, nil)
			return res, err
		default:
			// this is a function call
			if tvar, ok := vars[name]; ok {
				if tvar == types.TypeFunc || tvar == types.TypeUnknown {
					return types.TypeUnknown, nil
				}
				if tvar.Basic() == "func" {
					args := tvar.Arguments()
					if len(args) == 0 {
						return u, fmt.Errorf("%v: incorrect function type of %v: %v", fname, name, tvar)
					}
					rt := args[len(args)-1]
					for idx, item := range a.List[1:] {
						switch a := item.E.(type) {
						case types.Int:
							if ok, err := i.canConvertType(types.TypeInt, types.Type(args[idx])); !ok || err != nil {
								return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), types.TypeInt)
							}
						case types.Str:
							if ok, err := i.canConvertType(types.TypeStr, types.Type(args[idx])); !ok || err != nil {
								return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), types.TypeStr)
							}
						case types.Bool:
							if ok, err := i.canConvertType(types.TypeBool, types.Type(args[idx])); !ok || err != nil {
								return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), types.TypeBool)
							}
						case *types.Sexpr:
							if a.Empty() || a.Quoted {
								if ok, err := i.canConvertType(types.TypeList, types.Type(args[idx])); !ok || err != nil {
									return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), types.TypeList)
								}
							} else if a.Lambda {
								if ok, err := i.canConvertType(types.TypeFunc, types.Type(args[idx])); !ok || err != nil {
									return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), types.TypeFunc)
								}
							} else {
								itemType, err := i.exprType(fname, item, vars)
								if err != nil {
									return u, err
								}
								if !i.IsGeneric(itemType) {
									if ok, err := i.canConvertType(itemType, types.Type(args[idx])); !ok || err != nil {
										return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), itemType)
									}
								}
							}
						case types.Ident:
							itemType, err := i.exprType(fname, item, vars)
							if err != nil {
								return u, err
							}
							if !i.IsGeneric(itemType) {
								if ok, err := i.canConvertType(itemType, types.Type(args[idx])); !ok || err != nil {
									return u, fmt.Errorf("%v: cannot use %v as argument %d to %v: expected %v, found %v", fname, item, idx, name, types.Type(args[idx]), itemType)
								}
							}
						default:
							panic(fmt.Errorf("%v: unexpected type: %v", fname, item))
						}
					}
					return types.Type(rt), nil
				}
				return u, fmt.Errorf("%v: expected '%v' to be function, found: %v", fname, name, tvar)
			}
			f, ok := i.funcs[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "%v: cannot detect return type of function %v\n", fname, name)
				return types.TypeAny, nil
			}

			// check if we have matching func impl
			params := []types.Value{}
			for _, item := range a.List[1:] {
				switch a := item.E.(type) {
				case types.Int, types.Float, types.Str, types.Bool:
					params = append(params, item)
				case *types.Sexpr:
					if a.Empty() || a.Quoted {
						params = append(params, types.Value{T: types.TypeList, E: a})
					} else if a.Lambda {
						params = append(params, types.Value{T: types.TypeFunc})
					} else {
						itemType, err := i.exprType(fname, item, vars)
						if err != nil {
							return u, err
						}
						if i.IsGeneric(itemType) {
							itemType = types.TypeUnknown
						}
						params = append(params, types.Value{T: itemType})
					}
				case types.Ident:
					itemType, err := i.exprType(fname, item, vars)
					if err != nil {
						return u, err
					}
					if i.IsGeneric(itemType) {
						itemType = types.TypeUnknown
					}
					params = append(params, types.Value{T: itemType})
				default:
					panic(fmt.Errorf("%v: unexpected type: %v", fname, item))
				}
			}
			if binder, ok := f.(Binder); ok {
				t, err := binder.TryBindAll(params)
				if err != nil {
					return u, fmt.Errorf("%v: %v", fname, err)
				}

				return t, nil
			}

			return types.TypeUnknown, nil
		}
	}
	fmt.Fprintf(os.Stderr, "Unexpected return. (TypeAny)\n")
	return types.TypeAny, nil
}

func (in *Interpret) UnaliasType(t types.Type) types.Type {
	if tt, ok := in.typeAliases[t]; ok {
		return tt
	}
	return t
}

func (in *Interpret) parseType(token string) (types.Type, error) {
	t, ok := types.ParseType(token)
	if !ok {
		return types.TypeUnknown, fmt.Errorf("Token is not a type: %q", token)
	}
	t = in.UnaliasType(t)
	_, ok = in.types[t.Canonical()]
	if !ok {
		return "", fmt.Errorf("Cannot parse type %v: not defined", token)
	}
	return t, nil
}

func (in *Interpret) toParent(from, parent types.Type) (types.Type, error) {
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
			return types.TypeUnknown, fmt.Errorf("Cannot convert %v into %v", from, parent)
		}
		if f.Basic() == parent.Basic() {
			parent = f
			break
		}
		f = in.UnaliasType(f)
		if f.Basic() == parent.Basic() {
			parent = f
			break
		}
		par, ok := in.types[f.Canonical()]
		if !ok {
			return types.TypeUnknown, fmt.Errorf("Cannot convert type %v into %v: %v is not defined", from, parent, f)
		}
		f = par
	}
	res := parent.Basic()
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

	return types.Type(res), nil
}

func (in *Interpret) IsContract(t types.Type) bool {
	_, ok := in.contracts[t]
	return ok
}

func (in *Interpret) IsGeneric(t types.Type) bool {
	if in.IsContract(t) {
		return true
	}
	for _, a := range t.Arguments() {
		if in.IsGeneric(types.Type(a)) {
			return true
		}
	}
	return false
}
