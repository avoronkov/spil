package main

import (
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	tests := []struct {
		arg Type
		exp string
	}{
		{":int", "int"},
		{":list[a,b]", "list"},
	}
	for _, test := range tests {
		t.Run(string(test.arg), func(t *testing.T) {
			act := test.arg.Basic()
			if act != test.exp {
				t.Errorf("%q.Basic() failed: expected %v, actual %v", test.arg, test.exp, act)
			}
		})
	}
}

func TestArguments(t *testing.T) {
	tests := []struct {
		arg Type
		exp []string
	}{
		{":int", nil},
		{":list[a]", []string{"a"}},
		{":list[a,b,c]", []string{"a", "b", "c"}},
	}
	for _, test := range tests {
		t.Run(string(test.arg), func(t *testing.T) {
			act := test.arg.Arguments()
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf("%q.Arguments() failed: expected %v, actual %v", test.arg, test.exp, act)
			}
		})
	}
}
