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
	}

	for _, test := range testdata {
		for _, bigint := range []bool{false, true} {
			name := test.input
			if bigint {
				name += "-big"
			}
			t.Run(name, func(t *testing.T) {
				parseInt := ParseInt64
				if bigint {
					parseInt = ParseBigInt
				}
				p := NewParser(strings.NewReader(test.input), IntParserFn(parseInt))
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
