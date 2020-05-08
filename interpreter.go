package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Interpret struct {
	output   io.Writer
	funcs    map[string]Evaler
	mainBody []Expr

	builtinDir string

	parseInt func(token string) (Int, bool)

	lambdaCount int

	main *FuncInterpret
}

func NewInterpreter(w io.Writer, builtinDir string) *Interpret {
	i := &Interpret{
		output:     w,
		builtinDir: builtinDir,
		parseInt:   ParseInt64,
	}
	i.funcs = map[string]Evaler{
		"+":      EvalerFunc(FPlus, AllInts, TypeInt),
		"-":      EvalerFunc(FMinus, AllInts, TypeInt),
		"*":      EvalerFunc(FMultiply, AllInts, TypeInt),
		"/":      EvalerFunc(FDiv, AllInts, TypeInt),
		"mod":    EvalerFunc(FMod, TwoInts, TypeInt),
		"<":      EvalerFunc(FLess, TwoInts, TypeBool),
		"<=":     EvalerFunc(FLessEq, TwoInts, TypeBool),
		">":      EvalerFunc(FMore, TwoInts, TypeBool),
		">=":     EvalerFunc(FMoreEq, TwoInts, TypeBool),
		"=":      EvalerFunc(FEq, TwoArgs, TypeBool),
		"not":    EvalerFunc(FNot, OneBoolArg, TypeBool),
		"print":  EvalerFunc(i.FPrint, AnyArgs, TypeAny),
		"head":   EvalerFunc(FHead, ListArg, TypeAny),
		"tail":   EvalerFunc(FTail, ListArg, TypeList),
		"append": EvalerFunc(FAppend, AppenderArgs, TypeList),
		"list":   EvalerFunc(FList, AnyArgs, TypeList),
		"space":  EvalerFunc(FSpace, StrArg, TypeBool),
		"eol":    EvalerFunc(FEol, StrArg, TypeBool),
		"empty":  EvalerFunc(FEmpty, ListArg, TypeBool),
		"int":    EvalerFunc(i.FInt, StrArg, TypeInt),
		"open":   EvalerFunc(FOpen, StrArg, TypeStr),
	}
	return i
}

func (i *Interpret) UseBigInt(v bool) {
	if v {
		i.parseInt = ParseBigInt
	} else {
		i.parseInt = ParseInt64
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
			if err := i.parse(f); err != nil {
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
	return i.parseInt(token)
}

func (i *Interpret) parse(input io.Reader) error {
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
			if a.Len() == 0 {
				return fmt.Errorf("Unexpected empty s-expression on top-level: %v", a)
			}
			head, _ := a.Head()
			if name, ok := head.(Ident); ok {
				switch name {
				case "func", "def", "func'", "def'":
					memo := false
					if name == "func'" || name == "def'" {
						memo = true
					}
					tail, _ := a.Tail()
					if err := i.defineFunc(tail.(*Sexpr), memo); err != nil {
						return err
					}
					continue L
				case "use":
					tail, _ := a.Tail()
					if err := i.use(tail.(*Sexpr).List); err != nil {
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

func (i *Interpret) Parse(input io.Reader) error {
	if err := i.parse(input); err != nil {
		return err
	}
	// load builtin last
	if i.builtinDir != "" {
		if err := i.loadBuiltin(i.builtinDir); err != nil {
			return err
		}
	}

	i.main = NewFuncInterpret(i, "__main__")
	if err := i.main.AddImpl(QList(Ident("__stdin")), i.mainBody, false, TypeAny); err != nil {
		return err
	}
	return nil
}

// type-checking
func (i *Interpret) Check() error {
	return i.CheckReturnTypes()
}

func (i *Interpret) Run() error {
	stdin := NewLazyInput(os.Stdin)
	_, err := i.main.Eval([]Expr{stdin})
	return err
}

// (func-name) args body...
func (i *Interpret) defineFunc(se *Sexpr, memo bool) error {
	if se.Len() < 3 {
		return fmt.Errorf("Not enough arguments for function definition: %v", se)
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("func expected identifier first, found %v", se.List[0])
	}

	fname := string(name)
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
	returnType := TypeAny
	// Check if return type is specified
	if identType, ok := se.List[2].(Ident); ok {
		returnType, ok = ParseType(string(identType))
		if ok {
			bodyIndex++
		}
	}
	// TODO
	return fi.AddImpl(se.List[1], se.List[2:], memo, returnType)
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
		return i.parse(f)
	case Ident:
		switch string(a) {
		case "bigmath":
			i.UseBigInt(true)
		default:
			return fmt.Errorf("Unknown use-directive: %v", string(a))
		}
		return nil
	}
	return fmt.Errorf("Unexpected argument type to 'use': %v (%T)", module, module)
}

func (in *Interpret) FPrint(args []Expr) (Expr, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		e.Print(in.output)
	}
	fmt.Fprintf(in.output, "\n")
	return QEmpty, nil
}

// convert string into int
func (in *Interpret) FInt(args []Expr) (Expr, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FInt: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].(Str)
	if !ok {
		return nil, fmt.Errorf("FInt: expected argument to be Str, found %v", args)
	}
	i, ok := in.parseInt(string(s))
	if !ok {
		return nil, fmt.Errorf("FInt: cannot convert argument into Int: %v", s)
	}
	return i, nil
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

func (i *Interpret) CheckReturnTypes() error {
	for _, fn := range i.funcs {
		fi, ok := fn.(*FuncInterpret)
		if !ok {
			// native function
			continue
		}
		/*
			if fi.returnType == TypeAny {
				continue
			}
		*/
		for _, impl := range fi.bodies {
			t, err := i.evalBodyType(fi.name, impl.body, impl.argfmt.Values())
			if err != nil {
				return err
			}
			if fi.returnType != TypeAny {
				if t != fi.returnType {
					return fmt.Errorf("Incorrect return value in function %v(%v): expected %v actual %v", fi.name, impl.argfmt, fi.returnType, t)
				}
			}
		}
	}
	return nil
}

func (in *Interpret) evalBodyType(fname string, body []Expr, vars map[string]Type) (rt Type, err error) {

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
				return 0, fmt.Errorf("Expected ident, found: %v", a.List[0])
			}
			switch name := string(ident); name {
			case "set", "set'":
				if i == len(body)-1 {
					return 0, fmt.Errorf("Unexpected %v statement at the end of the function", name)
				}
				varname, ok := a.List[1].(Ident)
				if !ok {
					return 0, fmt.Errorf("%v: second argument should be variable name, found: %v", name, a.List[1])
				}
				if len(a.List) == 4 {
					id, ok := a.List[3].(Ident)
					if !ok {
						return 0, fmt.Errorf("Fourth statement of %v should be type identifier, found: %v", name, a.List[3])
					}
					tp, ok := ParseType(string(id))
					if !ok {
						return 0, fmt.Errorf("Fourth statement of %v should be type identifier, found: %v", name, a.List[3])
					}
					vars[string(varname)] = tp
				} else if len(a.List) == 3 {
					tp, err := in.exprType(fname, a.List[2], vars)
					if err != nil {
						return 0, err
					}
					vars[string(varname)] = tp
				} else {
					return 0, fmt.Errorf("%v: incorrect number of arguments %v: %v", fname, name, a.List)
				}
			case "print":
				for i, arg := range a.List[1:] {
					_, err := in.exprType(fname, arg, vars)
					if err != nil {
						return 0, fmt.Errorf("%v: incorrect argument to print at posision %v: %v", fname, i, err)
					}
				}
			default:
				if _, err := in.exprType(fname, a, vars); err != nil {
					return 0, fmt.Errorf("%v: %v", fname, err)
				}
			}
		}
	}
	return in.exprType(fname, body[len(body)-1], vars)
}

func (i *Interpret) exprType(fname string, e Expr, vars map[string]Type) (Type, error) {
	switch a := e.(type) {
	case Int:
		return TypeInt, nil
	case Str:
		return TypeStr, nil
	case Bool:
		return TypeBool, nil
	case Ident:
		if t, ok := vars[string(a)]; ok {
			return t, nil
		} else if _, ok := i.funcs[string(a)]; ok {
			return TypeFunc, nil
		}
		return 0, fmt.Errorf("Undefined variable: %v", string(a))
	case *Sexpr:
		if a.Quoted || a.Empty() {
			return TypeList, nil
		}
		if a.Lambda {
			return TypeFunc, nil
		}
		ident, ok := a.List[0].(Ident)
		if !ok {
			return 0, fmt.Errorf("%v: expected ident, found: %v", fname, a.List[0])
		}
		switch name := string(ident); name {
		case "set", "set'":
			return 0, fmt.Errorf("%v: unexpected %v and the end of function", fname, ident)
		case "lambda":
			return TypeFunc, nil
		case "and", "or":
			return TypeBool, nil
		case "gen", "gen'":
			return TypeList, nil
		case "apply":
			tail, _ := a.Tail()
			return i.exprType(fname, tail, vars)
		case "if":
			if len(a.List) != 4 {
				return 0, fmt.Errorf("%v: incorrect number of arguments to 'if': %v", fname, a.List)
			}
			condType, err := i.exprType(fname, a.List[1], vars)
			if err != nil {
				return 0, err
			}
			if condType != TypeBool {
				return 0, fmt.Errorf("%v: condition in if-statement should return :bool, found: %v", fname, condType)
			}
			t1, err := i.exprType(fname, a.List[2], vars)
			if err != nil {
				return 0, err
			}
			t2, err := i.exprType(fname, a.List[3], vars)
			if err != nil {
				return 0, err
			}
			if t1 == TypeAny || t2 == TypeAny {
				return TypeAny, nil
			}
			if t1 != t2 {
				return 0, fmt.Errorf("Different types returned by if-statement: %v != %v", t1, t2)
			}
			return t1, nil
		default:
			// this is a function call
			f, ok := i.funcs[name]
			if !ok {
				fmt.Fprintf(os.Stderr, "%v: cannot detect return type of function %v\n", fname, name)
				return TypeAny, nil
			}

			// check if we have matching func impl
			args := []Expr{}
			for _, item := range a.List[1:] {
				switch a := item.(type) {
				case Int, Str, Bool:
					args = append(args, a)
				case *Sexpr:
					if a.Empty() || a.Quoted {
						args = append(args, a)
					} else if a.Lambda {
						args = append(args, fakeArg(TypeFunc))
					} else {
						itemType, err := i.exprType(fname, item, vars)
						if err != nil {
							return 0, err
						}
						args = append(args, fakeArg(itemType))
					}
				case Ident:
					itemType, err := i.exprType(fname, item, vars)
					if err != nil {
						return 0, err
					}
					args = append(args, fakeArg(itemType))
				default:
					panic(fmt.Errorf("%v: unexpected type: %v", fname, item))
				}
			}
			err := f.TryBind(args)
			if err != nil {
				return 0, fmt.Errorf("%v: %v", fname, err)
			}

			return f.ReturnType(), nil
		}
	}
	fmt.Fprintf(os.Stderr, "Unexpected return. (TypeAny)\n")
	return TypeAny, nil
}

func fakeArg(t Type) Expr {
	switch t {
	case TypeInt:
		return Int64(-1339)
	case TypeStr:
		return Str("asdfasdlakdfa'adfask$%@")
	case TypeBool:
		return Bool(false)
	case TypeList:
		return QList(QEmpty, Int64(-1339))
	case TypeFunc, TypeAny:
		return Ident("__fake_function_name_adfaf12312f")
	default:
		panic(fmt.Errorf("Unexpected Type: %v", t))
	}
}
