package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Expr interface {
	fmt.Stringer
	// Write yourself into writer
	Print(io.Writer)
	Hash() (string, error)
	Type() Type
}

type List interface {
	Expr
	Head() (Expr, error)
	Tail() (List, error)
	Empty() bool
}

type Str string

var _ List = Str("")

var NotStr = errors.New("Token is not a string")

func ParseString(token string) (Str, error) {
	if !strings.HasPrefix(token, `"`) || !strings.HasSuffix(token, `"`) {
		return "", NotStr
	}
	str := token[1 : len(token)-1]
	str = strings.ReplaceAll(str, `\"`, `"`)
	str = strings.ReplaceAll(str, `\\`, `\`)
	str = strings.ReplaceAll(str, `\n`, "\n")
	return Str(str), nil
}

func (s Str) Head() (Expr, error) {
	if s == "" {
		return nil, fmt.Errorf("Cannot perform Head() on empty string")
	}
	return Str(string(s[0])), nil
}

func (s Str) Tail() (List, error) {
	if s == "" {
		return nil, fmt.Errorf("Cannot perform Tail() on empty string")
	}
	return Str(s[1:]), nil
}

func (s Str) Empty() bool {
	return s == ""
}

func (s Str) String() string {
	return fmt.Sprintf("{Str: %q}", string(s))
}

func (s Str) Hash() (string, error) {
	return s.String(), nil
}

func (s Str) Print(w io.Writer) {
	io.WriteString(w, string(s))
}

func (s Str) Append(args []Expr) (Expr, error) {
	result := string(s)
	for i, arg := range args {
		str, ok := arg.(Str)
		if !ok {
			return nil, fmt.Errorf("Str.Append() expect argument at position %v to be Str, found: %v", i, arg)
		}
		result += string(str)
	}
	return Str(result), nil
}

func (s Str) Type() Type {
	return TypeStr
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

func (i Ident) Type() Type {
	return TypeAny
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

func (i Bool) Type() Type {
	return TypeBool
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

func (s *Sexpr) Append(args []Expr) (Expr, error) {
	return &Sexpr{
		List:   append(s.List, args...),
		Quoted: s.Quoted,
	}, nil
}

func (s *Sexpr) Type() Type {
	return TypeList
}

func Equal(a, b Expr) bool {
	al, alist := a.(List)
	bl, blist := b.(List)
	if alist && blist && al.Empty() && bl.Empty() {
		return true
	}
	if a.Type() != b.Type() {
		return false
	}
	ha, aerr := a.Hash()
	hb, berr := b.Hash()
	if aerr != nil || berr != nil {
		return false
	}
	return ha == hb
}
