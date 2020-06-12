package main

import (
	"fmt"
	"os"
	"unicode"

	"github.com/avoronkov/spil/types"
)

type nativeFunc struct {
	name   string
	fn     func([]types.Value) (*types.Value, error)
	ret    types.Type
	binder func([]types.Value) error
}

func (n *nativeFunc) Eval(args []types.Value) (*types.Value, error) {
	return n.fn(args)
}

func (n *nativeFunc) ReturnType() types.Type {
	return n.ret
}

func (n *nativeFunc) TryBind(params []types.Value) (int, types.Type, map[string]types.Type, error) {
	if err := n.binder(params); err != nil {
		return -1, types.TypeUnknown, nil, fmt.Errorf("%v: %v", n.name, err)
	}
	return 0, n.ret, nil, nil
}

func (n *nativeFunc) TryBindAll(params []types.Value) (types.Type, error) {
	if err := n.binder(params); err != nil {
		return "", fmt.Errorf("%v: %v", n.name, err)
	}
	return n.ret, nil
}

func EvalerFunc(name string, fn func([]types.Value) (*types.Value, error), binder func([]types.Value) error, ret types.Type) types.Function {
	return &nativeFunc{
		name:   name,
		fn:     fn,
		ret:    ret,
		binder: binder,
	}
}

func MakeIntOperation(name string, op func(x, y types.Int) types.Int) func([]types.Value) (*types.Value, error) {
	return func(args []types.Value) (*types.Value, error) {
		var result types.Int
		for i, arg := range args {
			a, ok := arg.E.(types.Int)
			if !ok {
				return nil, fmt.Errorf("%v: expected integer argument, found %v", name, arg)
			}
			if i == 0 {
				result = a
			} else {
				result = op(result, a)
			}
		}
		return &types.Value{E: result, T: types.TypeInt}, nil
	}
}

var FPlus = MakeIntOperation("+", func(x, y types.Int) types.Int {
	return x.Plus(y)
})

var FMinus = MakeIntOperation("-", func(x, y types.Int) types.Int {
	return x.Minus(y)
})

var FMultiply = MakeIntOperation("*", func(x, y types.Int) types.Int {
	return x.Mult(y)
})

var FDiv = MakeIntOperation("/", func(x, y types.Int) types.Int {
	return x.Div(y)
})

func FMod(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FMod: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].E.(types.Int)
	if !ok {
		return nil, fmt.Errorf("FMod: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].E.(types.Int)
	if !ok {
		return nil, fmt.Errorf("FMod: second argument should be integer, found %v", args[1])
	}
	return &types.Value{E: a.Mod(b), T: types.TypeInt}, nil
}

func MakeFloatOperation(name string, op func(x, y types.Float) types.Float) func([]types.Value) (*types.Value, error) {
	return func(args []types.Value) (*types.Value, error) {
		var result types.Float
		for i, arg := range args {
			a, ok := arg.E.(types.Float)
			if !ok {
				return nil, fmt.Errorf("%v: expected float argument, found %v", name, arg)
			}
			if i == 0 {
				result = a
			} else {
				result = op(result, a)
			}
		}
		return &types.Value{E: result, T: types.TypeFloat}, nil
	}
}

var FloatPlus = MakeFloatOperation("+", func(x, y types.Float) types.Float {
	return x.Plus(y)
})

var FloatMinus = MakeFloatOperation("-", func(x, y types.Float) types.Float {
	return x.Minus(y)
})

var FloatMult = MakeFloatOperation("*", func(x, y types.Float) types.Float {
	return x.Mult(y)
})
var FloatDiv = MakeFloatOperation("/", func(x, y types.Float) types.Float {
	return x.Div(y)
})

func FIntLess(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FIntLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].E.(types.Int)
	if !ok {
		return nil, fmt.Errorf("FIntLess: first argument should be integer, found %v", args[0])
	}
	b, ok := args[1].E.(types.Int)
	if !ok {
		return nil, fmt.Errorf("FIntLess: second argument should be integer, found %v", args[1])
	}
	return &types.Value{E: types.Bool(a.Less(b)), T: types.TypeBool}, nil
}

func FStrLess(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FStrLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FStrLess: first argument should be string, found %v", args[0])
	}
	b, ok := args[1].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FStrLess: second argument should be string, found %v", args[1])
	}
	return &types.Value{E: types.Bool(string(a) < string(b)), T: types.TypeBool}, nil
}

func FFloatLess(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FFloatLess: expected 2 arguments, found %v", args)
	}
	a, ok := args[0].E.(types.Float)
	if !ok {
		return nil, fmt.Errorf("FFloatLess: first argument should be float, found %v", args[0])
	}
	b, ok := args[1].E.(types.Float)
	if !ok {
		return nil, fmt.Errorf("FFloatLess: second argument should be float, found %v", args[1])
	}
	return &types.Value{E: types.Bool(a.Less(b)), T: types.TypeBool}, nil
}

func FEq(args []types.Value) (*types.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("FEq: expected 2 arguments, found %v", args)
	}
	return &types.Value{E: types.Bool(types.Equal(args[0].E, args[1].E)), T: types.TypeBool}, nil
}

func FNot(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FNot: expected 1 argument, found %v", args)
	}
	a, ok := args[0].E.(types.Bool)
	if !ok {
		return nil, fmt.Errorf("FNot: expected argument to be Bool, found %v", args[0])
	}
	return &types.Value{E: types.Bool(!bool(a)), T: types.TypeBool}, nil
}

func FHead(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FHead: expected 1 argument, found %v", args)
	}
	a, ok := args[0].E.(types.List)
	if !ok {
		return nil, fmt.Errorf("FHead: expected argument to be List, found %v", args[0])
	}
	return a.Head()
}

func FTail(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FTail: expected 1 argument, found %v", args)
	}
	a, ok := args[0].E.(types.List)
	if !ok {
		return nil, fmt.Errorf("FTail: expected argument to be List, found %v", args[0])
	}
	t, err := a.Tail()
	if err != nil {
		return nil, err
	}
	return &types.Value{E: t, T: types.TypeList}, nil
}

func FEmpty(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEmpty: expected 1 argument, found %v", args)
	}
	a, ok := args[0].E.(types.List)
	if !ok {
		return nil, fmt.Errorf("FEmpty: expected argument to be List, found %v", args[0])
	}
	return &types.Value{E: types.Bool(a.Empty()), T: types.TypeBool}, nil
}

type Appender interface {
	Append([]types.Value) (*types.Value, error)
}

func FAppend(args []types.Value) (*types.Value, error) {
	if len(args) == 0 {
		return &types.Value{E: types.QEmpty, T: types.TypeList}, nil
	}
	if len(args) == 1 {
		return &args[0], nil
	}
	a, ok := args[0].E.(Appender)
	if !ok {
		return nil, fmt.Errorf("FAppend(1): expected first argument to be Appender, found %v", args[0])
	}
	return a.Append(args[1:])
}

func FList(args []types.Value) (*types.Value, error) {
	s := new(types.Sexpr)
	for _, a := range args {
		s.List = append(s.List, a)
	}
	s.Quoted = true
	return &types.Value{E: s, T: types.TypeList}, nil
}

// test if symbol is white-space
func FSpace(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FSpace: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FSpace: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FSpace: expected argument to be Str of length 1, found %v", s)
	}
	return &types.Value{E: types.Bool(unicode.IsSpace(rune(string(s)[0]))), T: types.TypeBool}, nil
}

// test if symbol is eol (\n)
func FEol(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FEol: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FEol: expected argument to be Str, found %v", args)
	}
	if len(s) != 1 {
		return nil, fmt.Errorf("FEol: expected argument to be Str of length 1, found %v", s)
	}
	return &types.Value{E: types.Bool(s == "\n"), T: types.TypeBool}, nil
}

func FOpen(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FOpen: expected exaclty one argument, found %v", args)
	}
	s, ok := args[0].E.(types.Str)
	if !ok {
		return nil, fmt.Errorf("FOpen: expected argument to be Str, found %v", args)
	}
	file, err := os.Open(string(s))
	if err != nil {
		return nil, err
	}
	return &types.Value{E: NewLazyInput(file), T: types.TypeStr}, nil
}

func FType(args []types.Value) (*types.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf(": expected exaclty one argument, found %v", args)
	}
	return &types.Value{E: types.Str(args[0].T.String()), T: types.TypeStr}, nil
}

type Lenghter interface {
	Length() int
}

type Nther interface {
	Nth(n int) (*types.Value, error)
}

// Binders
func SingleArg(params []types.Value) error {
	if len(params) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", params)
	}
	return nil
}

func (in *Interpret) AllInts(params []types.Value) error {
	for i, p := range params {
		if p.T == types.TypeUnknown || in.IsContract(p.T) {
			continue
		}
		ok, err := in.canConvertType(p.T, types.TypeInt)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("Expected all integer arguments, found %v at position %v", p, i)
		}
	}
	return nil
}

func (in *Interpret) TwoInts(params []types.Value) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	return in.AllInts(params)
}

func (in *Interpret) TwoStrs(params []types.Value) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	for i := 0; i < 2; i++ {
		ok, err := in.matchType(types.TypeStr, params[i].T, &map[string]types.Type{})
		if err != nil || !ok {
			return fmt.Errorf("Cannot convert argument %d to List: %v, %w", i, params[i], err)
		}
	}
	return nil
}

func TwoArgs(params []types.Value) error {
	if len(params) != 2 {
		return fmt.Errorf("expected 2 arguments, found %v", params)
	}
	return nil
}

func (in *Interpret) OneBoolArg(params []types.Value) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}
	ok, err := in.canConvertType(params[0].T, types.TypeBool)
	if err != nil {
		return err
	}
	if !ok && params[0].T != types.TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be Bool, found %v", params[0])
	}
	return nil
}

func AnyArgs(params []types.Value) error {
	return nil
}

func (in *Interpret) ListArg(params []types.Value) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 argument, found %v", params)
	}

	ok, err := in.matchType("list[a]", params[0].T, &map[string]types.Type{})
	if err != nil {
		return fmt.Errorf("Cannot convert first argument to List: %w", err)
	}
	if ok {
		return nil
	}

	if params[0].T != types.TypeList && params[0].T != types.TypeStr && params[0].T != types.TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be List, found %v", params[0])
	}
	return nil
}

func (in *Interpret) AppenderArgs(params []types.Value) error {
	if len(params) <= 1 {
		return nil
	}
	ok, err := in.matchType(types.Type("list[a]"), params[0].T, &map[string]types.Type{})
	if ok {
		return nil
	}
	if err != nil {
		return err
	}
	ok, err = in.canConvertType(params[0].T, types.TypeList)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	ok, err = in.canConvertType(params[0].T, types.TypeStr)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	if params[0].T != types.TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("AppenderArgs: expected first argument to be Appender, found %v", params[0])
	}
	return nil
}

func (in *Interpret) StrArg(params []types.Value) error {
	if len(params) != 1 {
		return fmt.Errorf("expected exaclty one argument, found %v", params)
	}
	ok, err := in.canConvertType(params[0].T, types.TypeStr)
	if err != nil {
		return err
	}
	if !ok && params[0].T != types.TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected argument to be Str, found %v", params)
	}
	return nil
}

func (in *Interpret) IntAndListArgs(params []types.Value) error {
	if len(params) != 2 {
		return fmt.Errorf("expected two arguments, found %v", params)
	}

	ok, err := in.canConvertType(params[0].T, types.TypeInt)
	if err != nil {
		return err
	}
	if !ok && params[0].T != types.TypeUnknown && !in.IsContract(params[0].T) {
		return fmt.Errorf("expected first argument to be Int, found %v", params)
	}

	ok, err = in.matchType("list[a]", params[1].T, &map[string]types.Type{})
	if err != nil {
		return fmt.Errorf("Second argument is not a list: %w", err)
	}
	if !ok && params[1].T != types.TypeUnknown && !in.IsContract(params[1].T) {
		return fmt.Errorf("expected second argument to be List, found %v", params)
	}
	return nil
}
