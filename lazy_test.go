package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/avoronkov/spil/types"
)

func TestLazyList(t *testing.T) {
	var ll types.List = NewLazyList(EvalerFunc("__func__", testCounter, AnyArgs, types.TypeList), []types.Value{{E: types.Int64(0), T: types.TypeInt}}, false)
	res := make([]types.Expr, 0, 10)
	for !ll.Empty() {
		val, err := ll.Head()
		if err != nil {
			t.Fatalf("Head() failed: %v", err)
		}
		res = append(res, val.E)
		ll, err = ll.Tail()
		if err != nil {
			t.Fatalf("Tail() failed: %v", err)
		}
	}
	exp := []types.Expr{
		types.Int64(1),
		types.Int64(2),
		types.Int64(3),
		types.Int64(4),
		types.Int64(5),
		types.Int64(6),
		types.Int64(7),
		types.Int64(8),
		types.Int64(9),
		types.Int64(10),
	}
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Incorrect sequence generated by LazyList: expected %v, actual %v", exp, res)
	}
}

// 1..10 generator
func testCounter(args []types.Value) (*types.Value, error) {
	prev := int64(args[0].E.(types.Int64))
	if prev >= 10 {
		return &types.Value{E: types.QEmpty, T: types.TypeList}, nil
	}
	s := &types.Sexpr{
		List:   []types.Value{{E: types.Int64(prev + 1), T: types.TypeInt}},
		Quoted: true,
	}
	return &types.Value{E: s, T: types.TypeList}, nil
}

func Int64Param(x int64) types.Value {
	return types.Value{
		E: types.Int64(x),
		T: types.TypeInt,
	}
}

func TestLazyListState(t *testing.T) {
	var ll types.List = NewLazyList(EvalerFunc("fibGen", fibGen, AnyArgs, types.TypeList), []types.Value{{E: types.QList(Int64Param(1), Int64Param(1)), T: types.TypeList}}, false)
	res := make([]types.Expr, 0, 10)
	for i := 0; i < 6; i++ {
		val, err := ll.Head()
		if err != nil {
			t.Fatalf("Head() failed: %v", err)
		}
		res = append(res, val.E)
		ll, err = ll.Tail()
		if err != nil {
			t.Fatalf("Tail() failed: %v", err)
		}
	}
	exp := []types.Expr{
		types.Int64(1),
		types.Int64(2),
		types.Int64(3),
		types.Int64(5),
		types.Int64(8),
		types.Int64(13),
	}
	if !reflect.DeepEqual(exp, res) {
		t.Errorf("Incorrect sequence generated by LazyList: expected %v, actual %v", exp, res)
	}
}

func fibGen(args []types.Value) (*types.Value, error) {
	state := args[0].E.(*types.Sexpr)
	a := int64(state.List[0].E.(types.Int64))
	b := int64(state.List[1].E.(types.Int64))
	a, b = b, a+b
	s := &types.Sexpr{
		List: []types.Value{
			Int64Param(a),
			types.Value{E: types.QList(Int64Param(a), Int64Param(b)), T: types.TypeList},
		},
		Quoted: true,
	}
	return &types.Value{E: s, T: types.TypeList}, nil
}

func TestLazyListFiniteString(t *testing.T) {
	ll := NewLazyList(
		EvalerFunc("__func__", testCounter, AnyArgs, types.TypeAny),
		[]types.Value{{E: types.Int64(0), T: types.TypeInt}},
		false,
	)

	var buffer strings.Builder
	ll.Print(&buffer)
	act := buffer.String()
	exp := "'(1 2 3 4 5 6 7 8 9 10)"
	if act != exp {
		t.Errorf("Incorrect string representation of LazyList: expected %q, actual %q", exp, act)
	}
}
