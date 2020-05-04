package main

import (
	"fmt"
	"io"
	"strings"
)

type Expr interface {
	fmt.Stringer
	// Write yourself into writer
	Print(io.Writer)
	Hash() (string, error)
}

type Str string

var _ Expr = Str("")

func (s Str) String() string {
	return fmt.Sprintf("{Str: %q}", string(s))
}

func (s Str) Hash() (string, error) {
	return s.String(), nil
}

func (s Str) Print(w io.Writer) {
	str := string(s)
	str = strings.ReplaceAll(str, `\"`, `"`)
	str = strings.ReplaceAll(str, `\\`, `\`)
	str = strings.ReplaceAll(str, `\n`, "\n")
	io.WriteString(w, str)
}

type Ident string

var _ Expr = Ident("")

func (i Ident) String() string {
	return fmt.Sprintf("{Ident: %v}", string(i))
}

func (i Ident) Hash() (string, error) {
	return i.String(), nil
}

func (i Ident) Print(w io.Writer) {
	io.WriteString(w, "_"+string(i))
}

type Bool bool

var _ Expr = Bool(false)

func (i Bool) String() string {
	if bool(i) {
		return "{Bool: 'T}"
	} else {
		return "{Bool: 'F}"
	}
}

func (i Bool) Hash() (string, error) {
	return i.String(), nil
}

func (i Bool) Print(w io.Writer) {
	if bool(i) {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
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
	Lambda bool
}

func QList(args ...Expr) *Sexpr {
	res := &Sexpr{Quoted: true}
	for _, arg := range args {
		res.List = append(res.List, arg)
	}
	return res
}

var QEmpty = &Sexpr{Quoted: true}

func (s *Sexpr) Print(w io.Writer) {
	if s.Quoted {
		fmt.Fprintf(w, "'(")
	} else {
		fmt.Fprintf(w, "(")
	}
	for i, item := range s.List {
		if i != 0 {
			fmt.Fprintf(w, " ")
		}
		item.Print(w)
	}
	fmt.Fprintf(w, ")")
}

func (s *Sexpr) String() string {
	b := &strings.Builder{}
	if s.Quoted {
		fmt.Fprintf(b, "{S':")
	} else {
		fmt.Fprintf(b, "{S:")
	}
	for _, item := range s.List {
		fmt.Fprintf(b, " %v", item)
	}
	fmt.Fprintf(b, "}")
	return b.String()
}

func (s *Sexpr) Hash() (string, error) {
	b := &strings.Builder{}
	if s.Quoted {
		fmt.Fprintf(b, "{S':")
	} else {
		fmt.Fprintf(b, "{S:")
	}
	for _, item := range s.List {
		hash, err := item.Hash()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(b, " %v", hash)
	}
	fmt.Fprintf(b, "}")
	return b.String(), nil
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
