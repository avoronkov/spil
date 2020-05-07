package main

import (
	"strconv"
	"testing"
)

func TestMatchArgs(t *testing.T) {
	tests := []struct {
		argfmt Expr
		args   []Expr
		exp    bool
	}{
		{QEmpty, []Expr{}, true},
		{QList(QEmpty), []Expr{QEmpty}, true},
		{QList(QEmpty), []Expr{QList(Int64(1))}, false},
		{QList(Ident("x"), QEmpty), []Expr{QList(Int64(1)), QEmpty}, true},
		{Ident("args"), []Expr{QList(Int64(1)), QEmpty}, true},
		{QList(Ident("x"), QEmpty), []Expr{QList(Int64(1)), QEmpty}, true},
		{QList(Ident("x"), Ident("x")), []Expr{Int64(1), Int64(1)}, true},
		{QList(Ident("x"), Ident("x")), []Expr{Int64(1), Int64(2)}, false},
		{QList(Ident("x"), Int64(1)), []Expr{Int64(1), Int64(2)}, false},
		{QList(Int64(1)), []Expr{Str("hello")}, false},
		{QList(QEmpty), []Expr{makeEmptyGen()}, true},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			act := matchArgs(test.argfmt, test.args)
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
	return NewLazyList(EvalerFunc(gen, TypeAny), QEmpty, false)
}
