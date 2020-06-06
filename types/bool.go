package types

import "io"

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
