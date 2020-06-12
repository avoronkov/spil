package types

import (
	"fmt"
	"io"
	"strconv"
)

type Float interface {
	Expr
	Plus(Float) Float
	Minus(Float) Float
	Mult(Float) Float
	Div(Float) Float
	Less(Float) bool
	Eq(Float) bool
	Float64() float64
}

type FloatMaker interface {
	ParseFloat(token string) (Float, bool)
	MakeFloat(f float64) Float
}

//
// Float64
//

type Float64 float64

var _ Float = Float64(0.0)

func (i Float64) String() string {
	return fmt.Sprintf("{Float64: %v}", float64(i))
}
func (i Float64) Hash() (string, error) {
	return i.String(), nil
}

func (i Float64) Print(w io.Writer) {
	fmt.Fprintf(w, "%v", float64(i))
}

func (i Float64) Plus(a Float) Float {
	return Float64(i + a.(Float64))
}

func (i Float64) Minus(a Float) Float {
	return Float64(i - a.(Float64))
}

func (i Float64) Mult(a Float) Float {
	return Float64(i * a.(Float64))
}

func (i Float64) Div(a Float) Float {
	return Float64(i / a.(Float64))
}

func (i Float64) Less(a Float) bool {
	return i < a.(Float64)
}

func (i Float64) Eq(a Float) bool {
	return i == a.(Float64)
}

func (i Float64) Type() Type {
	return TypeInt
}

func (i Float64) Float64() float64 {
	return float64(i)
}

type Float64Maker struct{}

var _ FloatMaker = Float64Maker{}

func (Float64Maker) ParseFloat(token string) (Float, bool) {
	n, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return nil, false
	}
	return Float64(n), true
}

func (Float64Maker) MakeFloat(f float64) Float {
	return Float64(f)
}
