package types

import (
	"fmt"
	"io"
	"strings"
)

type Sexpr struct {
	List   []Value
	Quoted bool
	Lambda bool
}

func QList(args ...Value) *Sexpr {
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
		item.E.Print(w)
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
		hash, err := item.E.Hash()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(b, " %v", hash)
	}
	fmt.Fprintf(b, "}")
	return b.String(), nil
}

func (s *Sexpr) Length() int {
	return len(s.List)
}

func (s *Sexpr) Nth(n int) (*Value, error) {
	if n > len(s.List) {
		return nil, fmt.Errorf("Index is out of range: %v", n)
	}
	return &s.List[n-1], nil
}

func (s *Sexpr) Head() (*Value, error) {
	if len(s.List) == 0 {
		return nil, fmt.Errorf("Cannot perform Head() on empty list")
	}
	return &s.List[0], nil
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

func (s *Sexpr) Append(values []Value) (*Value, error) {
	return &Value{
		E: &Sexpr{
			List:   append(s.List, values...),
			Quoted: s.Quoted,
		},
		T: TypeList,
	}, nil
}

func (s *Sexpr) Type() Type {
	return TypeList
}
