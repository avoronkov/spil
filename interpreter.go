package main

import (
	"fmt"
	"io"
)

type Interpret struct {
	parser   *Parser
	output   io.Writer
	vars     map[string]Expr
	deffuncs map[string]Func
	funcs    map[string]*FuncInterpet
}

func NewInterpreter(r io.Reader, w io.Writer) *Interpret {
	i := &Interpret{
		parser: NewParser(r),
		output: w,
		vars:   make(map[string]Expr),
		funcs:  make(map[string]*FuncInterpet),
	}
	i.deffuncs = map[string]Func{
		"+":      FPlus,
		"-":      FMinus,
		"*":      FMultiply,
		"/":      FDiv,
		"<":      FLess,
		">":      FMore,
		"=":      FEq,
		"not":    FNot,
		"print":  i.FPrint,
		"head":   FHead,
		"tail":   FTail,
		"append": FAppend,
	}
	return i
}

func (i *Interpret) Run() error {
	var mainBody []Expr
	// Parsing LOOP
L:
	for {
		expr, err := i.parser.NextExpr()
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
				return fmt.Errorf("Unexpected empty s-expression: %v", a)
			}
			head := a.Head()
			if name, ok := head.(Ident); ok {
				if name == "func" || name == "def" {
					if err := i.defineFunc(a.Tail()); err != nil {
						return err
					}
					continue L
				}
			}
		}
		mainBody = append(mainBody, expr)
	}
	// Interpreter LOOP
	mainInterpret, err := NewFuncInterpret(i, "__main__", &Sexpr{Quoted: true}, mainBody)
	if err != nil {
		return err
	}
	mainRuntime, err := mainInterpret.Bind(nil)
	if err != nil {
		return err
	}
	_, err = mainRuntime.Eval()
	return err
}

// (func-name) args body...
func (i *Interpret) defineFunc(se *Sexpr) error {
	if se.Len() < 3 {
		return fmt.Errorf("Not enough arguments for function definition: %v", se.Repr())
	}
	name, ok := se.List[0].(Ident)
	if !ok {
		return fmt.Errorf("func expected identifier first, found %v", se.List[0].Repr())
	}

	fi, err := NewFuncInterpret(i, string(name), se.List[1], se.List[2:])
	if err != nil {
		return err
	}
	i.funcs[string(name)] = fi
	return nil
}

func (in *Interpret) FPrint(args []Expr) (Expr, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		fmt.Fprintf(in.output, "%v", e.String())
	}
	fmt.Fprintf(in.output, "\n")
	return &Sexpr{Quoted: true}, nil
}
