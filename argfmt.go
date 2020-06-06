package main

import (
	"fmt"
	"strings"

	"github.com/avoronkov/spil/types"
)

type ArgFmt struct {
	Args []Arg
	// list argument that matches everything
	Wildcard string
}

func (a *ArgFmt) Values() map[string]types.Type {
	m := make(map[string]types.Type)
	if a.Wildcard != "" {
		m[a.Wildcard] = types.TypeList
	} else {
		for _, arg := range a.Args {
			if arg.Name != "" {
				m[arg.Name] = arg.T
			}
		}
	}
	return m
}

func MakeArgFmt(args ...Arg) (a *ArgFmt) {
	a = &ArgFmt{}
	for _, arg := range args {
		a.Args = append(a.Args, arg)
	}
	return
}

func MakeWildcard(name string) *ArgFmt {
	return &ArgFmt{Wildcard: name}
}

type Arg struct {
	Name string
	T    types.Type
	V    types.Expr
}

func ParseArgFmt(argfmt types.Expr) (*ArgFmt, error) {
	switch a := argfmt.(type) {
	case types.Ident:
		// pass arguments as list with specified name
		return &ArgFmt{Wildcard: string(a)}, nil
	case *types.Sexpr:
		// bind arguments
		result := &ArgFmt{}
		for _, arg := range a.List {
			switch r := arg.E.(type) {
			case types.Int:
				result.Args = append(result.Args, Arg{"", types.TypeInt, arg.E})
			case types.Str:
				result.Args = append(result.Args, Arg{"", types.TypeStr, arg.E})
			case types.Bool:
				result.Args = append(result.Args, Arg{"", types.TypeBool, arg.E})
			case *types.Sexpr:
				if !r.Empty() {
					return nil, fmt.Errorf("Unexpected non-empty list in a list of arguments")
				}
				result.Args = append(result.Args, Arg{"", types.TypeList, arg.E})
			case types.Ident:
				if colon := strings.Index(string(r), ":"); colon >= 0 {
					tp, ok := types.ParseType(string(r)[colon:])
					if !ok {
						return nil, fmt.Errorf("Unknown type is specified in argument %v", arg)
					}
					result.Args = append(result.Args, Arg{string(r)[:colon], tp, nil})
				} else {
					result.Args = append(result.Args, Arg{string(r), types.TypeUnknown, nil})
				}
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}

}

type Param struct {
	T types.Type
	V types.Expr
}

func MakeParametersFromArgs(args []types.Expr) (res []types.Value) {
	for _, arg := range args {
		p := types.Value{E: arg}
		switch arg.(type) {
		case types.Int:
			p.T = types.TypeInt
		case types.Str:
			p.T = types.TypeStr
		case types.Bool:
			p.T = types.TypeBool
		case types.List:
			p.T = types.TypeList
		case types.Ident:
			// TODO maybe check that function exist
			p.T = types.TypeFunc
		default:
			panic(fmt.Errorf("Unexpected Expr type: %v (%v)", arg, arg))
		}
		res = append(res, p)
	}
	return
}
