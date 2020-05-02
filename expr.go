package main

import (
	"fmt"
	"strings"
)

type Expr interface {
	fmt.Stringer
	Repr() string
}

type Str string

func (s Str) String() string {
	str := string(s)
	str = strings.ReplaceAll(str, `\"`, `"`)
	str = strings.ReplaceAll(str, `\\`, `\`)
	str = strings.ReplaceAll(str, `\n`, "\n")
	return str
}

func (s Str) Repr() string {
	return fmt.Sprintf("{Str: %q}", string(s))
}

type Ident string

func (i Ident) String() string {
	return "_" + string(i)
}

func (i Ident) Repr() string {
	return fmt.Sprintf("{Ident: %v}", string(i))
}

type Bool bool

func (i Bool) String() string {
	if bool(i) {
		return "true"
	} else {
		return "false"
	}
}

func (i Bool) Repr() string {
	if bool(i) {
		return "{Bool: 'T}"
	} else {
		return "{Bool: 'F}"
	}
}

type List interface {
	Expr
	Head() (Expr, error)
	Tail() (List, error)
	Empty() bool
}

var _ List = (*Sexpr)(nil)

type Sexpr struct {
	List   []Expr
	Quoted bool
}

func QList(args ...Expr) *Sexpr {
	res := &Sexpr{Quoted: true}
	for _, arg := range args {
		res.List = append(res.List, arg)
	}
	return res
}

var QEmpty = &Sexpr{Quoted: true}

func (s *Sexpr) String() string {
	b := &strings.Builder{}
	if s.Quoted {
		fmt.Fprintf(b, "'(")
	} else {
		fmt.Fprintf(b, "(")
	}
	for i, item := range s.List {
		if i != 0 {
			fmt.Fprintf(b, " ")
		}
		fmt.Fprintf(b, "%v", item)
	}
	fmt.Fprintf(b, ")")
	return b.String()
}

func (s *Sexpr) Repr() string {
	b := &strings.Builder{}
	if s.Quoted {
		fmt.Fprintf(b, "{S':")
	} else {
		fmt.Fprintf(b, "{S:")
	}
	for _, item := range s.List {
		fmt.Fprintf(b, " %v", item.Repr())
	}
	fmt.Fprintf(b, "}")
	return b.String()
}

func (s *Sexpr) Len() int {
	return len(s.List)
}

func (s *Sexpr) Head() (Expr, error) {
	if len(s.List) == 0 {
		return nil, fmt.Errorf("Cannot perform Head() on empty list")
	}
	return s.List[0], nil
}

func (s *Sexpr) Tail() (List, error) {
	if len(s.List) == 0 {
		return nil, fmt.Errorf("Cannot perform Tail() on empty list")
	}
	return &Sexpr{
		List:   s.List[1:],
		Quoted: s.Quoted,
	}, nil
}

func (s *Sexpr) Empty() bool {
	return len(s.List) == 0
}
