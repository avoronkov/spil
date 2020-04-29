package main

import (
	"fmt"
	"strings"
)

type Expr interface {
	fmt.Stringer
	Repr() string
}

type Int int64

func (i Int) String() string {
	return fmt.Sprintf("%d", int64(i))
}

func (i Int) Repr() string {
	return fmt.Sprintf("{Int: %d}", int64(i))
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

type Sexpr struct {
	List   []Expr
	Quoted bool
}

func (s *Sexpr) String() string {
	b := &strings.Builder{}
	if s.Quoted {
		fmt.Fprintf(b, "'(")
	} else {
		fmt.Fprintf(b, "(")
	}
	for _, item := range s.List {
		fmt.Fprintf(b, " %v", item.String())
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

// Mostly for equality of lists regardless of quoting.
func (s *Sexpr) ReprNoQuotes() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "{S:")
	for _, item := range s.List {
		fmt.Fprintf(b, " %v", item.Repr())
	}
	fmt.Fprintf(b, "}")
	return b.String()
}

func (s *Sexpr) Len() int {
	return len(s.List)
}

func (s *Sexpr) Head() Expr {
	if len(s.List) == 0 {
		panic(fmt.Errorf("Cannot perform Head() on empty list"))
	}
	return s.List[0]
}

func (s *Sexpr) Tail() *Sexpr {
	if len(s.List) == 0 {
		panic(fmt.Errorf("Cannot perform Tail() on empty list"))
	}
	return &Sexpr{
		List:   s.List[1:],
		Quoted: s.Quoted,
	}
}
