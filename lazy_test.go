package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLazyList(t *testing.T) {
	var ll List = NewLazyList(EvalerFunc(testCounter, AnyArgs, TypeList), Int64(0), false)
	res := make([]Expr, 0, 10)
	for !ll.Empty() {
		val, err := ll.Head()
		if err != nil {
			t.Fatalf("Head() failed: %v", err)
		}
		res = append(res, val)
		ll, err = ll.Tail()
		if err != nil {
			t.Fatalf("Tail() failed: %v", err)
		}
	}
	exp := []Expr{
		Int64(1),
		Int64(2),
		Int64(3),
		Int64(4),
		Int64(5),
		Int64(6),
		Int64(7),
		Int64(8),
		Int64(9),
		Int64(10),
	}
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Incorrect sequence generated by LazyList: expected %v, actual %v", exp, res)
	}
}

// 1..10 generator
func testCounter(args []Expr) (Expr, error) {
	prev := int64(args[0].(Int64))
	if prev >= 10 {
		return QEmpty, nil
	}
	return &Sexpr{
		List:   []Expr{Int64(prev + 1)},
		Quoted: true,
	}, nil
}

func TestLazyListState(t *testing.T) {
	var ll List = NewLazyList(EvalerFunc(fibGen, AnyArgs, TypeList), QList(Int64(1), Int64(1)), false)
	res := make([]Expr, 0, 10)
	for i := 0; i < 6; i++ {
		val, err := ll.Head()
		if err != nil {
			t.Fatalf("Head() failed: %v", err)
		}
		res = append(res, val)
		ll, err = ll.Tail()
		if err != nil {
			t.Fatalf("Tail() failed: %v", err)
		}
	}
	exp := []Expr{
		Int64(1),
		Int64(2),
		Int64(3),
		Int64(5),
		Int64(8),
		Int64(13),
	}
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Incorrect sequence generated by LazyList: expected %v, actual %v", exp, res)
	}
}

func fibGen(args []Expr) (Expr, error) {
	state := args[0].(*Sexpr)
	a := int64(state.List[0].(Int64))
	b := int64(state.List[1].(Int64))
	a, b = b, a+b
	return &Sexpr{
		List: []Expr{
			Int64(a),
			QList(Int64(a), Int64(b)),
		},
		Quoted: true,
	}, nil
}

func TestLazyListFiniteString(t *testing.T) {
	ll := NewLazyList(EvalerFunc(testCounter, AnyArgs, TypeAny), Int64(0), false)
	var buffer strings.Builder
	ll.Print(&buffer)
	act := buffer.String()
	exp := "'(1 2 3 4 5 6 7 8 9 10)"
	if act != exp {
		t.Errorf("Incorrect string representation of LazyList: expected %q, actual %q", exp, act)
	}
}
