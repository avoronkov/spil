package main

import (
	"fmt"
	"strings"
)

type Type int

const (
	TypeAny Type = iota
	TypeInt
	TypeStr
	TypeBool
	TypeFunc
	TypeList
)

func (t Type) String() string {
	switch t {
	case TypeAny:
		return ":any"
	case TypeInt:
		return ":int"
	case TypeStr:
		return ":str"
	case TypeBool:
		return ":bool"
	case TypeFunc:
		return ":func"
	case TypeList:
		return ":list"
	default:
		panic(fmt.Errorf("Unexpected type: %d", int(t)))
	}
}

func ParseType(token string) (Type, bool) {
	switch token {
	case ":any":
		return TypeAny, true
	case ":int":
		return TypeInt, true
	case ":str":
		return TypeStr, true
	case ":bool":
		return TypeBool, true
	case ":func":
		return TypeFunc, true
	case ":list":
		return TypeList, true
	default:
		return TypeAny, false
	}
}

type ArgFmt struct {
	Args []Arg
	// list argument that matches everything
	Wildcard string
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
	T    Type
	V    Expr
}

func ParseArgFmt(argfmt Expr) (*ArgFmt, error) {
	switch a := argfmt.(type) {
	case Ident:
		// pass arguments as list with specified name
		return &ArgFmt{Wildcard: string(a)}, nil
	case *Sexpr:
		// bind arguments
		result := &ArgFmt{}
		for _, arg := range a.List {
			switch r := arg.(type) {
			case Int:
				result.Args = append(result.Args, Arg{"", TypeInt, arg})
			case Str:
				result.Args = append(result.Args, Arg{"", TypeStr, arg})
			case Bool:
				result.Args = append(result.Args, Arg{"", TypeBool, arg})
			case *Sexpr:
				if !r.Empty() {
					return nil, fmt.Errorf("Unexpected non-empty list in a list of arguments")
				}
				result.Args = append(result.Args, Arg{"", TypeList, arg})
			case Ident:
				if colon := strings.Index(string(r), ":"); colon >= 0 {
					tp, ok := ParseType(string(r)[colon:])
					if !ok {
						return nil, fmt.Errorf("Unknown type is specified in argument %v", arg)
					}
					result.Args = append(result.Args, Arg{string(r)[:colon], tp, nil})
				} else {
					result.Args = append(result.Args, Arg{string(r), TypeAny, nil})
				}
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}

}
