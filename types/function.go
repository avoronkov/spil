package types

type Function interface {
	Eval([]Value) (*Value, error)
	ReturnType() Type
	TryBindAll(params []Value) (Type, error)
}
