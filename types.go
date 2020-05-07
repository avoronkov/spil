package main

import "fmt"

type Type int

const (
	TypeAny Type = iota
	TypeInt
	TypeStr
	TypeBool
	TypeFunc
	TypeList
)

func (t Type) String() string {
	switch t {
	case TypeAny:
		return ":any"
	case TypeInt:
		return ":int"
	case TypeStr:
		return ":str"
	case TypeBool:
		return ":bool"
	case TypeFunc:
		return ":func"
	case TypeList:
		return ":list"
	default:
		panic(fmt.Errorf("Unexpected type: %d", int(t)))
	}
}

func ParseType(token string) (Type, bool) {
	switch token {
	case ":any":
		return TypeAny, true
	case ":int":
		return TypeInt, true
	case ":str":
		return TypeStr, true
	case ":bool":
		return TypeBool, true
	case ":func":
		return TypeFunc, true
	case ":list":
		return TypeList, true
	default:
		return TypeAny, false
	}
}
