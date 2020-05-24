package main

import (
	"fmt"
	"os"
	"testing"
)

func TestMatchParameters(t *testing.T) {
	tests := []struct {
		name   string
		argfmt *ArgFmt
		args   []Param
		exp    bool
	}{
		{
			"no args",
			MakeArgFmt(),
			MakeParametersFromArgs([]Expr{}),
			true,
		},
		{
			"empty list",
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QEmpty}),
			true,
		},
		{
			"empty list (2)",
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1))}),
			false,
		},
		{
			"any + epty list",
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			"str + epty list",
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			false,
		},
		{
			"wildcard",
			MakeWildcard("args"),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			"any + empty list (2)",
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			true,
		},
		{
			"bool + empty list",
			MakeArgFmt(Arg{Name: "x", T: TypeBool}, Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{QList(Int64(1)), QEmpty}),
			false,
		},
		{
			"any x + x",
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{Name: "x", T: TypeAny}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			true,
		},
		{
			"any x + x (2)",
			MakeArgFmt(Arg{Name: "x", T: TypeAny}, Arg{Name: "x", T: TypeAny}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(2)}),
			false,
		},
		{
			"str x + x",
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{Name: "x", T: TypeStr}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			false,
		},
		{
			"str x + 1",
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(1)}),
			false,
		},
		{
			"str x + 1 (2)",
			MakeArgFmt(Arg{Name: "x", T: TypeStr}, Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Int64(1), Int64(2)}),
			false,
		},
		{
			"int 1",
			MakeArgFmt(Arg{T: TypeInt, V: Int64(1)}),
			MakeParametersFromArgs([]Expr{Str("Hello")}),
			false,
		},
		{
			"ampty list vs empty lazy list",
			MakeArgFmt(Arg{T: TypeList, V: QEmpty}),
			MakeParametersFromArgs([]Expr{makeEmptyGen()}),
			true,
		},
		{
			"int 13 vs unknown 13",
			MakeArgFmt(Arg{T: TypeInt, V: Int64(13)}),
			[]Param{Param{T: TypeUnknown, V: Int64(13)}},
			true,
		},
		{
			"int 15 vs unknown 439",
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
			"list vs set",
			MakeArgFmt(Arg{Name: "n", T: TypeList}),
			[]Param{Param{T: Type(":set"), V: QEmpty}},
			true,
		},
		{
			"list vs any",
			MakeArgFmt(Arg{Name: "n", T: TypeList}),
			[]Param{Param{T: TypeAny}},
			false,
		},
		{
			"any vs list",
			MakeArgFmt(Arg{Name: "n", T: TypeAny}),
			[]Param{Param{T: TypeList}},
			true,
		},
		{
			":a :a vs :int :int",
			MakeArgFmt(Arg{Name: "a", T: ":a"}, Arg{Name: "b", T: ":a"}),
			[]Param{{T: TypeInt}, {T: TypeInt}},
			true,
		},
		{
			":a :a vs :int :str",
			MakeArgFmt(Arg{Name: "a", T: ":a"}, Arg{Name: "b", T: ":a"}),
			[]Param{Param{T: TypeInt}, Param{T: TypeStr}},
			false,
		},
		{
			":a :list[a] vs :int :list[int]",
			MakeArgFmt(Arg{Name: "elem", T: ":a"}, Arg{Name: "lst", T: ":list[a]"}),
			[]Param{Param{T: TypeInt}, Param{T: ":list[int]"}},
			true,
		},
		{
			":a :list[a] vs :int :list[any]",
			MakeArgFmt(Arg{Name: "elem", T: ":a"}, Arg{Name: "lst", T: ":list[a]"}),
			[]Param{Param{T: TypeInt}, Param{T: ":list[any]"}},
			false,
		},
	}
	in := NewInterpreter(os.Stderr, "")
	fi := NewFuncInterpret(in, "__test__")
	in.types[Type(":set")] = TypeList
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			act, _ := fi.matchParameters(test.argfmt, test.args)
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
	tests := []struct {
		from, to Type
		res      bool
	}{
		{":int", ":any", true},
		{":list[a]", ":any", true},
		{":list[z]", ":any", true},
		{":list", ":any", true},
	}
	in := NewInterpreter(os.Stderr, "")
	for _, test := range tests {
		t.Run(fmt.Sprintf("%v %v", test.from, test.to), func(t *testing.T) {
			ok, err := in.canConvertType(test.from, test.to)
			if err != nil || ok != test.res {
				t.Errorf("Incorrect conversion from %v to %v: expected %v, actual %v (%v)", test.from, test.to, test.res, ok, err)
			}

		})
	}
}

func TestMatchType(t *testing.T) {
	tests := []struct {
		name   string
		arg    Type
		val    Type
		binds  *map[string]Type
		result bool
	}{
		{"int-int", ":int", ":int", &map[string]Type{}, true},
		{"any-int", ":any", ":int", &map[string]Type{}, true},
		{"a-int", ":a", ":int", &map[string]Type{}, true},
		{"a-int-str", ":a", ":str", &map[string]Type{"a": ":int"}, false},
		{"list[a]-list[int]", ":list[a]", ":list[int]", &map[string]Type{}, true},
		{"list[int]-tset[int]", ":list[int]", ":tset[int]", &map[string]Type{}, true},
		{"list[a]-tset[int]", ":list[a]", ":tset[int]", &map[string]Type{}, true},
		{"list[any]-set", ":list[any]", ":set", &map[string]Type{}, true},
		{"some[a,a]-some[x,y]", ":some[a,a]", ":some[x,y]", &map[string]Type{}, false},
		{"some[a,b]-some[x,y]", ":some[a,b]", ":some[x,y]", &map[string]Type{}, true},
		{"some[int,b]-some[int,y]", ":some[int,b]", ":some[int,y]", &map[string]Type{}, true},
		{"some[int,b]-some[x,y]", ":some[a,b]", ":some[int,y]", &map[string]Type{}, true},
		{"some[a,b]-intsome[a]", ":some[a,b]", ":intsome[a]", &map[string]Type{}, true},
		{"some[a,list[a]]-some[int,list[int]]", ":some[a,list[a]]", ":some[int,list[int]]", &map[string]Type{}, true},
		{"some[a,list[a]]-some[int,list[str]]", ":some[a,list[a]]", ":some[int,list[str]]", &map[string]Type{}, false},
	}

	in := NewInterpreter(os.Stderr, "")
	in.types[":some[a,b]"] = TypeAny
	in.types[":set"] = ":list[any]"
	in.types[":tset[a]"] = ":list[a]"
	in.types[":intsome[a]"] = ":some[int,a]"

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
		from, parent Type
		exp          Type
	}{
		{"set->list", ":set", "list", ":list[any]"},
		{"tset[bool]->list", ":tset[bool]", "list", ":list[bool]"},
		{"list[int]->any", ":list[int]", "any", ":any"},
		{"intsome[str]->some", ":intsome[str]", "some", ":some[int,str]"},
	}

	in := NewInterpreter(os.Stderr, "")
	in.types[":some[a,b]"] = TypeAny
	in.types[":set"] = ":list[any]"
	in.types[":tset[a]"] = ":list[a]"
	in.types[":intsome[a]"] = ":some[int,a]"

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
