package main

import (
	"fmt"
	"os"
	"unicode"
)

type Evaler interface {
	Eval([]Param) (*Param, error)
	ReturnType() Type
	TryBind(params []Param) (int, Type, map[string]Type, error)
}

type nativeFunc struct {
	name   string
	fn     func([]Param) (*Param, error)
	ret    Type
	binder func([]Param) error
}

func (n *nativeFunc) Eval(args []Param) (*Param, error) {
	return n.fn(args)
}

func (n *nativeFunc) ReturnType() Type {
	return n.ret
}

func (n *nativeFunc) TryBind(params []Param) (int, Type, map[string]Type, error) {
	if err := n.binder(params); err != nil {
		return -1, TypeUnknown, nil, fmt.Errorf("%v: %v", n.name, err)
	}
	return 0, n.ret, nil, nil
}

func EvalerFunc(name string, fn func([]Param) (*Param, error), binder func([]Param) error, ret Type) Evaler {
	return &nativeFunc{
		name:   name,
		fn:     fn,
		ret:    ret,
		binder: binder,
	}
}

func FPlus(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FPlus: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Plus(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMinus(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMinus: expected integer argument in position %v, found %v", i, arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Minus(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMultiply(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Mult(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FDiv(args []Param) (*Param, error) {
	var result Int
	for i, arg := range args {
		a, ok := arg.V.(Int)
		if !ok {
			return nil, fmt.Errorf("FMultiply: expected integer argument, found %v", arg)
		}
		if i == 0 {
			result = a
		} else {
			result = result.Div(a)
		}
	}
	return &Param{V: result, T: TypeInt}, nil
}

func FMod(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMod: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FMod: second argument should be integer, found %v", args[1])
	}
	return &Param{V: a.Mod(b), T: TypeInt}, nil
}

func FIntLess(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FIntLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FIntLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].V.(Int)
	if !ok {
		return nil, fmt.Errorf("FIntLess: second argument should be integer, found %v", args[1])
	}
	return &Param{V: Bool(a.Less(b)), T: TypeBool}, nil
}

func FStrLess(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FStrLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FStrLess: first argument should be string, found %v", args[0])
	}
	b, ok := args[1].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FStrLess: second argument should be string, found %v", args[1])
	}
	return &Param{V: Bool(string(a) < string(b)), T: TypeBool}, nil
}

func FEq(args []Param) (*Param, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FEq: expected 2 arguments, found %v", args)
	}
	return &Param{V: Bool(Equal(args[0].V, args[1].V)), T: TypeBool}, nil
}

func FNot(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FNot: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(Bool)
	if !ok {
		return nil, fmt.Errorf("FNot: expected argument to be Bool, found %v", args[0])
	}
	return &Param{V: Bool(!bool(a)), T: TypeBool}, nil
}

func FHead(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FHead: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FHead: expected argument to be List, found %v", args[0])
	}
	return a.Head()
}

func FTail(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FTail: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FTail: expected argument to be List, found %v", args[0])
	}
	t, err := a.Tail()
	if err != nil {
		return nil, err
	}
	return &Param{V: t, T: TypeList}, nil
}

func FEmpty(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEmpty: expected 1 argument, found %v", args)
	}
	a, ok := args[0].V.(List)
	if !ok {
		return nil, fmt.Errorf("FEmpty: expected argument to be List, found %v", args[0])
	}
	return &Param{V: Bool(a.Empty()), T: TypeBool}, nil
}

type Appender interface {
	Append([]Param) (*Param, error)
}

func FAppend(args []Param) (*Param, error) {
	if len(args) == 0 {
		return &Param{V: QEmpty, T: TypeList}, nil
	}
	if len(args) == 1 {
		return &args[0], nil
	}
	a, ok := args[0].V.(Appender)
	if !ok {
		return nil, fmt.Errorf("FAppend(1): expected first argument to be Appender, found %v", args[0])
	}
	return a.Append(args[1:])
}

func FList(args []Param) (*Param, error) {
	s := new(Sexpr)
	for _, a := range args {
		s.List = append(s.List, a.V)
	}
	s.Quoted = true
	return &Param{V: s, T: TypeList}, nil
}

// test if symbol is white-space
func FSpace(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FSpace: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FSpace: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FSpace: expected argument to be Str of length 1, found %v", s)
	}
	return &Param{V: Bool(unicode.IsSpace(rune(string(s)[0]))), T: TypeBool}, nil
}

// test if symbol is eol (\n)
func FEol(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEol: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FEol: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FEol: expected argument to be Str of length 1, found %v", s)
	}
	return &Param{V: Bool(s == "\n"), T: TypeBool}, nil
}

func FOpen(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FOpen: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].V.(Str)
	if !ok {
		return nil, fmt.Errorf("FOpen: expected argument to be Str, found %v", args)
	}
	file, err := os.Open(string(s))
	if err != nil {
		return nil, err
	}
	return &Param{V: NewLazyInput(file), T: TypeStr}, nil
}

func FType(args []Param) (*Param, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(": expected exaclty one argument, found %v", args)
	}
	return &Param{V: Str(args[0].T.String()), T: TypeStr}, nil
}

type Lenghter interface {
	Length() int
}

type Nther interface {
	Nth(n int) (*Param, error)
}

// Binders
func SingleArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", params)
	}
	return nil
}

func (in *Interpret) AllInts(params []Param) error {
	for i, p := range params {
		if p.T == TypeUnknown || in.IsContract(p.T) {
			continue
		}
		ok, err := in.canConvertType(p.T, TypeInt)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("Expected all integer arguments, found %v at position %v", p, i)
		}
	}
	return nil
}

func (in *Interpret) TwoInts(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	return in.AllInts(params)
}

func (in *Interpret) TwoStrs(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	for i := 0; i < 2; i++ {
		ok, err := in.matchType(TypeStr, params[i].T, &map[string]Type{})
		if err != nil || !ok {
			return fmt.Errorf("Cannot convert argument %d to List: %v, %w", i, params[i], err)
		}
	}
	return nil
}

func TwoArgs(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	return nil
}

func (in *Interpret) OneBoolArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}
	ok, err := in.canConvertType(params[0].T, TypeBool)
	if err != nil {
		return err
	}
	if !ok && params[0].T != TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be Bool, found %v", params[0])
	}
	return nil
}

func AnyArgs(params []Param) error {
	return nil
}

func (in *Interpret) ListArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}

	ok, err := in.matchType("list[a]", params[0].T, &map[string]Type{})
	if err != nil {
		return fmt.Errorf("Cannot convert first argument to List: %w", err)
	}
	if ok {
		return nil
	}

	if params[0].T != TypeList && params[0].T != TypeStr && params[0].T != TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be List, found %v", params[0])
	}
	return nil
}

func (in *Interpret) AppenderArgs(params []Param) error {
	if len(params) <= 1 {
		return nil
	}
	ok, err := in.matchType(Type("list[a]"), params[0].T, &map[string]Type{})
	if ok {
		return nil
	}
	if err != nil {
		return err
	}
	ok, err = in.canConvertType(params[0].T, TypeList)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	ok, err = in.canConvertType(params[0].T, TypeStr)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	if params[0].T != TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("AppenderArgs: expected first argument to be Appender, found %v", params[0])
	}
	return nil
}

func (in *Interpret) StrArg(params []Param) error {
	if len(params) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", params)
	}
	ok, err := in.canConvertType(params[0].T, TypeStr)
	if err != nil {
		return err
	}
	if !ok && params[0].T != TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be Str, found %v", params)
	}
	return nil
}

func (in *Interpret) IntAndListArgs(params []Param) error {
	if len(params) != 2 {
		return fmt.Errorf("expected two arguments, found %v", params)
	}

	ok, err := in.canConvertType(params[0].T, TypeInt)
	if err != nil {
		return err
	}
	if !ok && params[0].T != TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected first argument to be Int, found %v", params)
	}

	ok, err = in.canConvertType(params[1].T, TypeList)
	if err != nil {
		return fmt.Errorf("Second argument is not a list: %w", err)
	}
	if !ok && params[1].T != TypeUnknown && !in.IsContract(params[1].T) {
		return fmt.Errorf("expected second argument to be List, found %v", params)
	}
	return nil
}
