package types

import (
	"fmt"
	"io"
)

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
