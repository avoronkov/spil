package types

import (
	"fmt"
	"io"
)

type Expr interface {
	fmt.Stringer
	// Write yourself into writer
	Print(io.Writer)
	Hash() (string, error)
	Type() Type
}

func Equal(a, b Expr) bool {
	al, alist := a.(List)
	if alist && al.Empty() {
		bl, blist := b.(List)
		return blist && bl.Empty()
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
