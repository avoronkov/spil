package main

import (
	"fmt"
	"io"
)

type Interpret struct {
	parser *Parser
	output io.Writer
	vars   map[string]Expr
	funcs  map[string]Evaler
}

func NewInterpreter(r io.Reader, w io.Writer) *Interpret {
	i := &Interpret{
		parser: NewParser(r),
		output: w,
		vars:   make(map[string]Expr),
	}
	i.funcs = map[string]Evaler{
		"+":      EvalerFunc(FPlus),
		"-":      EvalerFunc(FMinus),
		"*":      EvalerFunc(FMultiply),
		"/":      EvalerFunc(FDiv),
		"<":      EvalerFunc(FLess),
		">":      EvalerFunc(FMore),
		"=":      EvalerFunc(FEq),
		"not":    EvalerFunc(FNot),
		"print":  EvalerFunc(i.FPrint),
		"head":   EvalerFunc(FHead),
		"tail":   EvalerFunc(FTail),
		"append": EvalerFunc(FAppend),
		"empty":  EvalerFunc(FEmpty),
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
			head, _ := a.Head()
			if name, ok := head.(Ident); ok {
				if name == "func" || name == "def" {
					tail, _ := a.Tail()
					if err := i.defineFunc(tail.(*Sexpr)); err != nil {
						return err
					}
					continue L
				}
			}
		}
		mainBody = append(mainBody, expr)
	}
	// Interpreter LOOP
	mainInterpret := NewFuncInterpret(i, "__main__")
	if err := mainInterpret.AddImpl(QEmpty, mainBody); err != nil {
		return err
	}
	_, err := mainInterpret.Eval(nil)
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
	return fi.AddImpl(se.List[1], se.List[2:])
}

func (in *Interpret) FPrint(args []Expr) (Expr, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Fprintf(in.output, " ")
		}
		fmt.Fprintf(in.output, "%v", e.String())
	}
	fmt.Fprintf(in.output, "\n")
	return QEmpty, nil
}
