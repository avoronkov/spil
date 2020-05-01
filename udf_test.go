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
		{QList(QEmpty), []Expr{QList(Int(1))}, false},
		{QList(Ident("x"), QEmpty), []Expr{QList(Int(1)), QEmpty}, true},
		{Ident("args"), []Expr{QList(Int(1)), QEmpty}, true},
		{QList(Ident("x"), QEmpty), []Expr{QList(Int(1)), QEmpty}, true},
		{QList(Ident("x"), Ident("x")), []Expr{Int(1), Int(1)}, true},
		{QList(Ident("x"), Ident("x")), []Expr{Int(1), Int(2)}, false},
		{QList(Ident("x"), Int(1)), []Expr{Int(1), Int(2)}, false},
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
