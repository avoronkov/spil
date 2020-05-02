package main

import (
	"fmt"
	"math/big"
	"strconv"
)

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

func ParseInt64(token string) (Int64, bool) {
	n, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		return Int64(0), false
	}
	return Int64(n), true
}

func (i Int64) String() string {
	return fmt.Sprintf("%d", int64(i))
}

func (i Int64) Repr() string {
	return fmt.Sprintf("{Int64: %d}", int64(i))
}

func (i Int64) Plus(a Int) Int {
	return Int64(i + a.(Int64))
}

func (i Int64) Minus(a Int) Int {
	return Int64(i - a.(Int64))
}

func (i Int64) Mult(a Int) Int {
	return Int64(i * a.(Int64))
}

func (i Int64) Div(a Int) Int {
	return Int64(i / a.(Int64))
}

func (i Int64) Less(a Int) bool {
	return i < a.(Int64)
}

func (i Int64) Eq(a Int) bool {
	return i == a.(Int64)
}

type BigInt struct {
	value *big.Int
}

var _ Int = (*BigInt)(nil)

func ParseBigInt(token string) (*BigInt, bool) {
	res := &big.Int{}
	_, ok := res.SetString(token, 10)
	if !ok {
		return nil, false
	}
	return &BigInt{res}, true
}

func (i *BigInt) String() string {
	return i.value.String()
}

func (i *BigInt) Repr() string {
	return fmt.Sprintf("{BigInt: %v}", i.value)
}

func (i *BigInt) Plus(a Int) Int {
	res := &big.Int{}
	res.Add(i.value, a.(*BigInt).value)
	return &BigInt{res}
}

func (i *BigInt) Minus(a Int) Int {
	res := &big.Int{}
	res.Sub(i.value, a.(*BigInt).value)
	return &BigInt{res}
}

func (i *BigInt) Mult(a Int) Int {
	res := &big.Int{}
	res.Mul(i.value, a.(*BigInt).value)
	return &BigInt{res}
}

func (i *BigInt) Div(a Int) Int {
	res := &big.Int{}
	res.Div(i.value, a.(*BigInt).value)
	return &BigInt{res}
}

func (i *BigInt) Less(a Int) bool {
	return i.value.Cmp(a.(*BigInt).value) < 0
}

func (i *BigInt) Eq(a Int) bool {
	return i.value.Cmp(a.(*BigInt).value) == 0
}
