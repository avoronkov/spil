package main

import (
	"os"
	"strconv"
	"testing"
)

func TestMatchArgs(t *testing.T) {
	tests := []struct {
		argfmt *ArgFmt
		args   []Parameter
		exp    bool
	}{
		{
			MakeArgFmt(),
			MakeParametersFromArgs([]Expr{}),
			true,
		},
		{
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QEmpty}),
			true,
		},
		{
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1))}),
			false,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			false,
		},
		{
			MakeWildcard("args"),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeBool}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			false,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{Name: "x", T: TypeAny}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			true,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{Name: "x", T: TypeAny}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(2)}),
			false,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{Name: "x", T: TypeStr}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			false,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			false,
		},
		{
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(2)}),
			false,
		},
		{
			MakeArgFmt(Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Str("Hello")}),
			false,
		},
		{
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{makeEmptyGen()}),
			true,
		},
	}
	in := NewInterpreter(os.Stderr, "")
	fi := NewFuncInterpret(in, "__test__")
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			act := fi.matchParameters(test.argfmt, test.args)
			if act != test.exp {
				t.Errorf("Incorrect matchArgs(%v, %v) != %v", test.argfmt, test.args, test.exp)
			}
		})
	}
}

func makeEmptyGen() List {
	gen := func(args []Expr) (Expr, error) {
		return QEmpty, nil
	}
	return NewLazyList(EvalerFunc("__gen__", gen, AnyArgs, TypeAny), QEmpty, false)
}
