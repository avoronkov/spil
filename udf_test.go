package main

import (
	"os"
	"strconv"
	"testing"
)

func TestMatchArgs(t *testing.T) {
	tests := []struct {
		argfmt *ArgFmt
		args   []Param
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
		{
			MakeArgFmt(Arg{T: TypeInt, V: Int64(13)}),
			[]Param{Param{T: TypeUnknown, V: Int64(13)}},
			true,
		},
		{
			MakeArgFmt(Arg{Name: "n", T: TypeInt, V: Int64(15)}),
			[]Param{Param{T: TypeUnknown, V: Int64(439)}},
			false,
		},
		/*
			{
				MakeArgFmt(Arg{Name: "n", T: TypeInt, V: Int64(1)}),
				[]Param{Param{T: TypeAny, V: Int64(1)}},
				true,
			},
		*/
		{
			MakeArgFmt(Arg{Name: "n", T: TypeList}),
			[]Param{Param{T: Type(":set"), V: QEmpty}},
			true,
		},
		{
			MakeArgFmt(Arg{Name: "n", T: TypeList}),
			[]Param{Param{T: TypeAny}},
			false,
		},
		{
			MakeArgFmt(Arg{Name: "n", T: TypeAny}),
			[]Param{Param{T: TypeList}},
			true,
		},
	}
	in := NewInterpreter(os.Stderr, "")
	fi := NewFuncInterpret(in, "__test__")
	in.types[Type(":set")] = TypeList
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
	gen := func(args []Param) (*Param, error) {
		return &Param{V: QEmpty, T: TypeList}, nil
	}
	return NewLazyList(EvalerFunc("__gen__", gen, AnyArgs, TypeAny), []Param{{V: QEmpty, T: TypeList}}, false)
}

func TestCanConvertType(t *testing.T) {
	in := NewInterpreter(os.Stderr, "")
	ok, err := in.canConvertType(Type(":int"), Type(":any"))
	if err != nil || !ok {
		t.Errorf("Cannot convert :int into :any (%v)", err)
	}
}
