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
}

func NewInterpreter(w io.Writer, builtinDir string) *Interpret {
	i := &Interpret{
		output:     w,
		builtinDir: builtinDir,
		parseInt:   ParseInt64,
	}
	i.funcs = map[string]Evaler{
		"+":      EvalerFunc(FPlus),
		"-":      EvalerFunc(FMinus),
		"*":      EvalerFunc(FMultiply),
		"/":      EvalerFunc(FDiv),
		"mod":    EvalerFunc(FMod),
		"<":      EvalerFunc(FLess),
		"<=":     EvalerFunc(FLessEq),
		">":      EvalerFunc(FMore),
		">=":     EvalerFunc(FMoreEq),
		"=":      EvalerFunc(FEq),
		"not":    EvalerFunc(FNot),
		"print":  EvalerFunc(i.FPrint),
		"head":   EvalerFunc(FHead),
		"tail":   EvalerFunc(FTail),
		"append": EvalerFunc(FAppend),
		"list":   EvalerFunc(FList),
		"space":  EvalerFunc(FSpace),
		"eol":    EvalerFunc(FEol),
		"empty":  EvalerFunc(FEmpty),
		"int":    EvalerFunc(i.FInt),
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

func (i *Interpret) Run(input io.Reader) error {
	if err := i.parse(input); err != nil {
		return err
	}
	// load builtin last
	if i.builtinDir != "" {
		if err := i.loadBuiltin(i.builtinDir); err != nil {
			return err
		}
	}

	// Interpreter LOOP
	mainInterpret := NewFuncInterpret(i, "__main__")
	if err := mainInterpret.AddImpl(QList(Ident("__stdin")), i.mainBody, false); err != nil {
		return err
	}
	stdin := NewLazyInput()
	_, err := mainInterpret.Eval([]Expr{stdin})
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
	// TODO
	return fi.AddImpl(se.List[1], se.List[2:], memo)
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
		i.UseBigInt(true)
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
