package types

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Str string

var NotStr = errors.New("Token is not a string")

func ParseString(token string) (Str, error) {
	if !strings.HasPrefix(token, `"`) || !strings.HasSuffix(token, `"`) {
		return "", NotStr
	}
	str := token[1 : len(token)-1]
	str = strings.ReplaceAll(str, `\"`, `"`)
	str = strings.ReplaceAll(str, `\\`, `\`)
	str = strings.ReplaceAll(str, `\n`, "\n")
	return Str(str), nil
}

func (s Str) Head() (*Value, error) {
	if s == "" {
		return nil, fmt.Errorf("Cannot perform Head() on empty string")
	}
	return &Value{E: Str(string(s[0])), T: TypeStr}, nil
}

func (s Str) Tail() (List, error) {
	if s == "" {
		return nil, fmt.Errorf("Cannot perform Tail() on empty string")
	}
	return Str(s[1:]), nil
}

func (s Str) Empty() bool {
	return s == ""
}

func (s Str) String() string {
	return fmt.Sprintf("{Str: %q}", string(s))
}

func (s Str) Hash() (string, error) {
	return s.String(), nil
}

func (s Str) Print(w io.Writer) {
	io.WriteString(w, string(s))
}

func (s Str) Append(args []Value) (*Value, error) {
	result := string(s)
	for i, arg := range args {
		str, ok := arg.E.(Str)
		if !ok {
			return nil, fmt.Errorf("Str.Append() expect argument at position %v to be Str, found: %v", i, arg)
		}
		result += string(str)
	}
	return &Value{E: Str(result), T: TypeStr}, nil
}

func (s Str) Type() Type {
	return TypeStr
}
