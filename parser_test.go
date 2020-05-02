package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestNextToken(t *testing.T) {
	testdata := []struct {
		input  string
		result []string
	}{
		{`foo`, []string{"foo"}},
		{`( hello )`, []string{"(", "hello", ")"}},
		{`(hello )`, []string{"(", "hello", ")"}},
		{`( hello)`, []string{"(", "hello", ")"}},
		{`(hello)`, []string{"(", "hello", ")"}},
		{`(print (+ 1 2))`, []string{"(", "print", "(", "+", "1", "2", ")", ")"}},
		{`(print "hello world" )`, []string{"(", "print", `"hello world"`, ")"}},
		{`"\""`, []string{`"\""`}},
		{`"foo \" bar"`, []string{`"foo \" bar"`}},
		{`(print "hello \"world\"" )`, []string{"(", "print", `"hello \"world\""`, ")"}},
		{"(hello)\ntrue\n#vim ft=lisp", []string{"(", "hello", ")", "true"}},
	}

	for _, test := range testdata {
		t.Run(test.input, func(t *testing.T) {
			p := NewParser(strings.NewReader(test.input), false)
			var tokens []string
			for {
				tok, err := p.nextToken()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("nextToken() failed; %v", err)
				}
				tokens = append(tokens, tok)
			}
			if !reflect.DeepEqual(tokens, test.result) {
				t.Errorf("Tokens are parsed incorrectly:\nexpected %s,\n  actual %s", test.result[0], tokens[0])
			}
		})
	}
}
