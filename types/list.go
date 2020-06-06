package types

type List interface {
	Expr
	Head() (*Value, error)
	Tail() (List, error)
	Empty() bool
}
