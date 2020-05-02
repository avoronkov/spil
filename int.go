package main

import "fmt"

type Int interface {
	Expr
	Plus(Int) Int
	Minus(Int) Int
	Mult(Int) Int
	Div(Int) Int
	Less(Int) bool
	Eq(Int) bool
}

type Int64 int64

var _ Int = Int64(0)

func (i Int64) String() string {
	return fmt.Sprintf("%d", int64(i))
}

func (i Int64) Repr() string {
	return fmt.Sprintf("{Int64: %d}", int64(i))
}

func (i Int64) Plus(a Int) Int {
	return Int(i + a.(Int64))
}

func (i Int64) Minus(a Int) Int {
	return Int(i - a.(Int64))
}

func (i Int64) Mult(a Int) Int {
	return Int(i * a.(Int64))
}

func (i Int64) Div(a Int) Int {
	return Int(i / a.(Int64))
}

func (i Int64) Less(a Int) bool {
	return i < a.(Int64)
}

func (i Int64) Eq(a Int) bool {
	return i == a.(Int64)
}
