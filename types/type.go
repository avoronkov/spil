package types

import "strings"

type Type string

const (
	TypeUnknown Type = "unknown"
	TypeAny     Type = "any"
	TypeInt     Type = "int"
	TypeFloat   Type = "float"
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
	s := string(t)[l+1 : len(t)-1]
	return splitArguments(s)
}

func splitArguments(argStr string) (args []string) {
	start := 0
	braces := 0
	for i := 0; i < len(argStr); i++ {
		switch c := argStr[i]; c {
		case ',':
			if braces == 0 {
				args = append(args, argStr[start:i])
				start = i + 1
			}
		case '[':
			braces++
		case ']':
			braces--
		}
	}
	if start < len(argStr) {
		args = append(args, argStr[start:])
	}
	return
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
