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
		{`\(foo bar)`, []string{`\(`, "foo", "bar", ")"}},
		{`"(set n (get-int) :int)"`, []string{`"(set n (get-int) :int)"`}},
	}

	for _, test := range testdata {
		for _, bigint := range []bool{false, true} {
			name := test.input
			if bigint {
				name += "-big"
			}
			t.Run(name, func(t *testing.T) {
				var numberParser NumberParser = defaultNumberParser{}
				if bigint {
					numberParser = bigNumberParser{}
				}
				p := NewParser(strings.NewReader(test.input), numberParser)
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
					t.Errorf("Tokens are parsed incorrectly:\nexpected %v,\n  actual %v", test.result, tokens)
				}
			})
		}
	}
}

func TestNextExpr(t *testing.T) {
	testdata := []struct {
		input  string
		result string
	}{
		{`'((1 2 3) (4 5 6))`, `{S': {S': {Int64: 1} {Int64: 2} {Int64: 3}} {S': {Int64: 4} {Int64: 5} {Int64: 6}}}`},
	}
	for _, test := range testdata {
		name := test.input
		t.Run(name, func(t *testing.T) {
			var numberParser NumberParser = defaultNumberParser{}
			if bigint {
				numberParser = bigNumberParser{}
			}
			p := NewParser(strings.NewReader(test.input), numberParser)
			res, err := p.NextExpr(false)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := res.E.Hash()
			if err != nil {
				t.Fatal(err)
			}
			if act, exp := hash, test.result; act != exp {
				t.Errorf("Incorrect NextToken() result:\nexpected %v,\n  actual %v", exp, act)
			}
		})
	}
}
