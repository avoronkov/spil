package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/avoronkov/spil/types"
)

func TestMatchParameters(t *testing.T) {
	tests := []struct {
		name   string
		argfmt *ArgFmt
		args   []types.Value
		exp    bool
	}{
		{
			"no args",
			MakeArgFmt(),
			MakeParametersFromArgs([]types.Expr{}),
			true,
		},
		{
			"empty list",
			MakeArgFmt(Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QEmpty}),
			true,
		},
		{
			"empty list (2)",
			MakeArgFmt(Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1))}),
			false,
		},
		{
			"any + epty list",
			MakeArgFmt(Arg{Name: "x", T: types.TypeAny}, Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1)), types.QEmpty}),
			true,
		},
		{
			"str + epty list",
			MakeArgFmt(Arg{Name: "x", T: types.TypeStr}, Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1)), types.QEmpty}),
			false,
		},
		{
			"wildcard",
			MakeWildcard("args"),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1)), types.QEmpty}),
			true,
		},
		{
			"any + empty list (2)",
			MakeArgFmt(Arg{Name: "x", T: types.TypeAny}, Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1)), types.QEmpty}),
			true,
		},
		{
			"bool + empty list",
			MakeArgFmt(Arg{Name: "x", T: types.TypeBool}, Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{types.QList(Int64Param(1)), types.QEmpty}),
			false,
		},
		{
			"any x + x",
			MakeArgFmt(Arg{Name: "x", T: types.TypeAny}, Arg{Name: "x", T: types.TypeAny}),
			MakeParametersFromArgs([]types.Expr{types.Int64(1), types.Int64(1)}),
			true,
		},
		{
			"any x + x (2)",
			MakeArgFmt(Arg{Name: "x", T: types.TypeAny}, Arg{Name: "x", T: types.TypeAny}),
			MakeParametersFromArgs([]types.Expr{types.Int64(1), types.Int64(2)}),
			false,
		},
		{
			"str x + x",
			MakeArgFmt(Arg{Name: "x", T: types.TypeStr}, Arg{Name: "x", T: types.TypeStr}),
			MakeParametersFromArgs([]types.Expr{types.Int64(1), types.Int64(1)}),
			false,
		},
		{
			"str x + 1",
			MakeArgFmt(Arg{Name: "x", T: types.TypeStr}, Arg{T: types.TypeInt, V: types.Int64(1)}),
			MakeParametersFromArgs([]types.Expr{types.Int64(1), types.Int64(1)}),
			false,
		},
		{
			"str x + 1 (2)",
			MakeArgFmt(Arg{Name: "x", T: types.TypeStr}, Arg{T: types.TypeInt, V: types.Int64(1)}),
			MakeParametersFromArgs([]types.Expr{types.Int64(1), types.Int64(2)}),
			false,
		},
		{
			"int 1",
			MakeArgFmt(Arg{T: types.TypeInt, V: types.Int64(1)}),
			MakeParametersFromArgs([]types.Expr{types.Str("Hello")}),
			false,
		},
		{
			"ampty list vs empty lazy list",
			MakeArgFmt(Arg{T: types.TypeList, V: types.QEmpty}),
			MakeParametersFromArgs([]types.Expr{makeEmptyGen()}),
			true,
		},
		{
			"int 13 vs unknown 13",
			MakeArgFmt(Arg{T: types.TypeInt, V: types.Int64(13)}),
			[]types.Value{{T: types.TypeUnknown, E: types.Int64(13)}},
			true,
		},
		{
			"int 15 vs unknown 439",
			MakeArgFmt(Arg{Name: "n", T: types.TypeInt, V: types.Int64(15)}),
			[]types.Value{{T: types.TypeUnknown, E: types.Int64(439)}},
			false,
		},
		/*
			{
				MakeArgFmt(Arg{Name: "n", T: types.TypeInt, V: types.Int64(1)}),
				[]Param{Param{T: types.TypeAny, V: types.Int64(1)}},
				true,
			},
		*/
		{
			"list vs set",
			MakeArgFmt(Arg{Name: "n", T: types.TypeList}),
			[]types.Value{{T: types.Type("set"), E: types.QEmpty}},
			true,
		},
		{
			"list vs any",
			MakeArgFmt(Arg{Name: "n", T: types.TypeList}),
			[]types.Value{{T: types.TypeAny}},
			false,
		},
		{
			"any vs list",
			MakeArgFmt(Arg{Name: "n", T: types.TypeAny}),
			[]types.Value{{T: types.TypeList}},
			true,
		},
		{
			":a :a vs :int :int",
			MakeArgFmt(Arg{Name: "a", T: "a"}, Arg{Name: "b", T: "a"}),
			[]types.Value{{T: types.TypeInt}, {T: types.TypeInt}},
			true,
		},
		{
			":a :a vs :int :str",
			MakeArgFmt(Arg{Name: "a", T: "a"}, Arg{Name: "b", T: "a"}),
			[]types.Value{{T: types.TypeInt}, {T: types.TypeStr}},
			false,
		},
		{
			":a :list[a] vs :int :list[int]",
			MakeArgFmt(Arg{Name: "elem", T: "a"}, Arg{Name: "lst", T: "list[a]"}),
			[]types.Value{{T: types.TypeInt}, {T: "list[int]"}},
			true,
		},
		{
			":a :list[a] vs :int :list[any]",
			MakeArgFmt(Arg{Name: "elem", T: "a"}, Arg{Name: "lst", T: "list[a]"}),
			[]types.Value{{T: types.TypeInt}, {T: "list[any]"}},
			false,
		},
		{
			":list vs :list[any]",
			MakeArgFmt(Arg{Name: "lst", T: "list[any]"}),
			[]types.Value{{T: "list"}},
			true,
		},
		{
			":list[any] vs :list",
			MakeArgFmt(Arg{Name: "lst", T: "list"}),
			[]types.Value{{T: "list[any]"}},
			true,
		},
	}
	in := NewInterpreter(os.Stderr, getTestLibraryDir())
	fi := NewFuncInterpret(in, "__test__")
	in.types[types.Type("set")] = "list[any]"

	// contracts
	in.types["a"] = ""
	in.contracts["a"] = struct{}{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act, _ := fi.matchParameters(test.argfmt, test.args)
			if act != test.exp {
				t.Errorf("Incorrect matchArgs(%v, %v) != %v", test.argfmt, test.args, test.exp)
			}
		})
	}
}

func makeEmptyGen() types.List {
	gen := func(args []types.Value) (*types.Value, error) {
		return &types.Value{E: types.QEmpty, T: types.TypeList}, nil
	}
	return NewLazyList(
		EvalerFunc("__gen__", gen, AnyArgs, types.TypeAny),
		[]types.Value{{E: types.QEmpty, T: types.TypeList}},
		false,
	)
}

func TestCanConvertType(t *testing.T) {
	tests := []struct {
		from, to types.Type
		res      bool
	}{
		{"int", "any", true},
		{"list[a]", "any", true},
		{"list[z]", "any", true},
		{"list", "any", true},
	}
	in := NewInterpreter(os.Stderr, getTestLibraryDir())
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v", test.from, test.to), func(t *testing.T) {
			ok, err := in.canConvertType(test.from, test.to)
			if err != nil || ok != test.res {
				t.Errorf("Incorrect conversion from %v to %v: expected %v, actual %v (%v)", test.from, test.to, test.res, ok, err)
			}

		})
	}
}

func emptyStringTypeMap() *map[string]types.Type {
	m := make(map[string]types.Type)
	return &m
}

func TestMatchType(t *testing.T) {
	tests := []struct {
		name   string
		arg    types.Type
		val    types.Type
		binds  *map[string]types.Type
		result bool
	}{
		{"int-int", "int", "int", emptyStringTypeMap(), true},
		{"any-int", "any", "int", emptyStringTypeMap(), true},
		{"a-int", "a", "int", emptyStringTypeMap(), true},
		{"a-int-str", "a", "str", &map[string]types.Type{"a": "int"}, false},
		{"list[a]-list[int]", "list[a]", "list[int]", emptyStringTypeMap(), true},
		{"list[int]-tset[int]", "list[int]", "tset[int]", emptyStringTypeMap(), true},
		{"list[a]-tset[int]", "list[a]", "tset[int]", emptyStringTypeMap(), true},
		{"list[any]-set", "list[any]", "set", emptyStringTypeMap(), true},
		{"some[a,a]-some[x,y]", "some[a,a]", "some[x,y]", emptyStringTypeMap(), false},
		{"some[a,b]-some[x,y]", "some[a,b]", "some[x,y]", emptyStringTypeMap(), true},
		{"some[int,b]-some[int,y]", "some[int,b]", "some[int,y]", emptyStringTypeMap(), true},
		{"some[int,b]-some[x,y]", "some[a,b]", "some[int,y]", emptyStringTypeMap(), true},
		{"some[a,b]-intsome[a]", "some[a,b]", "intsome[a]", emptyStringTypeMap(), true},
		{"some[a,list[a]]-some[int,list[int]]", "some[a,list[a]]", "some[int,list[int]]", emptyStringTypeMap(), true},
		{"some[a,list[a]]-some[int,list[str]]", "some[a,list[a]]", "some[int,list[str]]", emptyStringTypeMap(), false},
		{"list-list[any]", "list", "list[any]", emptyStringTypeMap(), true},
		{"list[any]-list", "list[any]", "list", emptyStringTypeMap(), true},
		{"list-any", "any", "list", emptyStringTypeMap(), true},
		{"list[a]-list", "list[a]", "list", emptyStringTypeMap(), true},
		{"list[a]-list[any]", "list[a]", "list[any]", emptyStringTypeMap(), true},
		{"func[a]-func", "func[a]", "func", emptyStringTypeMap(), true},
		{"func-func[a]", "func", "func[a]", emptyStringTypeMap(), true},
	}

	in := NewInterpreter(os.Stderr, getTestLibraryDir())
	in.types["some[a,b]"] = types.TypeAny
	in.types["set"] = "list[any]"
	in.types["tset[a]"] = "list[a]"
	in.types["intsome[a]"] = "some[int,a]"

	// contracts
	in.types["a"] = ""
	in.types["b"] = ""
	in.contracts["a"] = struct{}{}
	in.contracts["b"] = struct{}{}

	for i, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			binds := *test.binds
			t.Logf("%v: binds = %v", i, binds)
			res, err := in.matchType(test.arg, test.val, &binds)
			if err != nil {
				t.Fatalf("matchType(%v, %v, %v) failed: %v", test.arg, test.val, test.binds, err)
			}
			if res != test.result {
				t.Errorf("matchType(%v, %v, %v) incorrect: expected %v, actual %v", test.arg, test.val, test.binds, test.result, res)
			}
		})
	}
}

func TestToParent(t *testing.T) {
	tests := []struct {
		name         string
		from, parent types.Type
		exp          types.Type
	}{
		{"set->list", "set", "list", "list[any]"},
		{"tset[bool]->list", "tset[bool]", "list", "list[bool]"},
		{"list[int]->any", "list[int]", "any", "any"},
		{"intsome[str]->some", "intsome[str]", "some", "some[int,str]"},
	}

	in := NewInterpreter(os.Stderr, getTestLibraryDir())
	in.types["some[a,b]"] = types.TypeAny
	in.types["set"] = "list[any]"
	in.types["tset[a]"] = "list[a]"
	in.types["intsome[a]"] = "some[int,a]"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act, err := in.toParent(test.from, test.parent)
			if err != nil {
				t.Fatalf("toParent(%v, %v) failed: %v", test.from, test.parent, err)
			}
			if act != test.exp {
				t.Errorf("toParent(%v, %v) failed: expected %v, actual %v", test.from, test.parent, test.exp, act)
			}
		})
	}
}
