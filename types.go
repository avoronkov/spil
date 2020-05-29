package main

import (
	"fmt"
	"strings"
)

type Type string

const (
	TypeUnknown Type = "unknown"
	TypeAny     Type = "any"
	TypeInt     Type = "int"
	TypeStr     Type = "str"
	TypeBool    Type = "bool"
	TypeFunc    Type = "func"
	TypeList    Type = "list"
)

func (t Type) String() string {
	return ":" + string(t)
}

// ":list[a]" -> "list"
func (t Type) Basic() string {
	res := string(t)
	if p := strings.Index(res, "["); p >= 0 {
		res = res[:p]
	}
	return res
}

// ":tuple[a,b,c]" -> ["a", "b", "c"]
func (t Type) Arguments() []string {
	l := strings.Index(string(t), "[")
	if l < 0 {
		return nil
	}
	s := strings.TrimRight(string(t)[l+1:], "]")
	return strings.Split(s, ",")
}

// ":x[int,str,list]" -> "x[a,b,c]"
func (t Type) Canonical() Type {
	res := t.Basic()
	args := t.Arguments()
	if len(args) > 0 {
		res += "["
		for i := range args {
			if i > 0 {
				res += ","
			}
			res += string('a' + i)
		}
		res += "]"
	}
	return Type(res)
}

func (t Type) Expand(types map[string]Type) Type {
	if types == nil {
		return t
	}
	if newT, ok := types[t.Basic()]; ok {
		return newT
	}
	args := t.Arguments()
	if len(args) == 0 {
		return t
	}

	res := t.Basic() + "["
	for i, a := range args {
		if i > 0 {
			res += ","
		}
		res += string(Type(a).Expand(types))
	}
	res += "]"
	return Type(res)
}

func ParseType(token string) (Type, bool) {
	if strings.HasPrefix(token, ":") {
		return Type(token[1:]), true
	}
	return TypeUnknown, false
}

type ArgFmt struct {
	Args []Arg
	// list argument that matches everything
	Wildcard string
}

func (a *ArgFmt) Values() map[string]Type {
	m := make(map[string]Type)
	if a.Wildcard != "" {
		m[a.Wildcard] = TypeList
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
					result.Args = append(result.Args, Arg{string(r), TypeUnknown, nil})
				}
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("Expected arguments signature, found: %v", argfmt)
	}

}

type Param struct {
	T Type
	V Expr
}

func MakeParametersFromArgs(args []Expr) (res []Param) {
	for _, arg := range args {
		p := Param{V: arg}
		switch arg.(type) {
		case Int:
			p.T = TypeInt
		case Str:
			p.T = TypeStr
		case Bool:
			p.T = TypeBool
		case List:
			p.T = TypeList
		case Ident:
			// TODO maybe check that function exist
			p.T = TypeFunc
		default:
			panic(fmt.Errorf("Unexpected Expr type: %v (%v)", arg, arg))
		}
		res = append(res, p)
	}
	return
}
