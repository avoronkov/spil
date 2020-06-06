package types

import "testing"

func TestStrString(t *testing.T) {
	s := Str(`hello " world`)
	if act, exp := s.String(), `{Str: "hello \" world"}`; act != exp {
		t.Errorf("Incorrect string representation of Str:\nexpected %q,\n  actual %q", exp, act)
	}
}
