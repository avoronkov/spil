package types

import "strings"

type Type string

const (
	TypeUnknown Type = "unknown"
	TypeAny     Type = "any"
	TypeInt     Type = "int"
	TypeStr     Type = "str"
	TypeBool    Type = "bool"
	TypeFunc    Type = "func"
	TypeList    Type = "list"
)

func ParseType(token string) (Type, bool) {
	if strings.HasPrefix(token, ":") {
		return Type(token[1:]), true
	}
	return TypeUnknown, false
}

func (t Type) String() string {
	return ":" + string(t)
}

// ":list[a]" -> "list"
func (t Type) Basic() string {
	res := string(t)
	if p := strings.Index(res, "["); p >= 0 {
		res = res[:p]
	}
	return res
}

// ":tuple[a,b,c]" -> ["a", "b", "c"]
func (t Type) Arguments() []string {
	l := strings.Index(string(t), "[")
	if l < 0 {
		return nil
	}
	s := strings.TrimRight(string(t)[l+1:], "]")
	return strings.Split(s, ",")
}

// ":x[int,str,list]" -> "x[a,b,c]"
func (t Type) Canonical() Type {
	res := t.Basic()
	args := t.Arguments()
	if len(args) > 0 {
		res += "["
		for i := range args {
			if i > 0 {
				res += ","
			}
			res += string('a' + i)
		}
		res += "]"
	}
	return Type(res)
}

func (t Type) Expand(types map[string]Type) Type {
	if types == nil {
		return t
	}
	if newT, ok := types[t.Basic()]; ok {
		return newT
	}
	args := t.Arguments()
	if len(args) == 0 {
		return t
	}

	res := t.Basic() + "["
	for i, a := range args {
		if i > 0 {
			res += ","
		}
		res += string(Type(a).Expand(types))
	}
	res += "]"
	return Type(res)
}
