package main

import "fmt"

type Func func([]Expr) (Expr, error)

var Funcs = map[string]Func{
	"print": FPrint,
	"+":     FPlus,
}

func FPrint(args []Expr) (Expr, error) {
	for i, e := range args {
		if i > 0 {
			fmt.Printf(" ")
		}
		fmt.Printf("%v", e.String())
	}
	fmt.Printf("\n")
	return &Sexpr{Quoted: true}, nil
}

func FPlus(args []Expr) (Expr, error) {
	var result int
	for _, arg := range args {
		a, ok := arg.(Int)
		if !ok {
			return nil, fmt.Errorf("FPlus: expected integer argument, found %v", arg.Repr())
		}
		result += int(a)
	}
	return Int(result), nil
}
