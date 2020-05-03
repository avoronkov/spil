package main

import "testing"

func BenchmarkFEq(b *testing.B) {
	args := []Expr{
		&Sexpr{
			List: []Expr{Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(1)},
		},
		&Sexpr{
			List: []Expr{Int64(1), Int64(1), Int64(1), Int64(1), Int64(1), Int64(2), Int64(1), Int64(1), Int64(1), Int64(1)},
		},
	}
	var err error
	for i := 0; i < b.N; i++ {
		_, err = FEq(args)
	}
	if err != nil {
		b.Fatal(err)
	}
}
