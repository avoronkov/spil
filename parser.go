package main

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"

	"github.com/avoronkov/spil/types"
)

var (
	UnexpectedEOF = errors.New("Unexpected EOF")
)

type IntParser interface {
	ParseInt(token string) (types.Int, bool)
}

type IntParserFn func(token string) (types.Int, bool)

func (f IntParserFn) ParseInt(token string) (types.Int, bool) {
	return f(token)
}

type Parser struct {
	scanner *bufio.Scanner
	tokens  []string

	intParser IntParser
}

func NewParser(r io.Reader, intParser IntParser) *Parser {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return &Parser{
		scanner:   scanner,
		intParser: intParser,
	}
}

func (p *Parser) NextExpr() (*types.Value, error) {
	token, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if token == "(" || token == "'(" || token == "\\(" {
		item, err := p.nextSexpr(token, token == "'(")
		if err != nil {
			return nil, err
		}
		return item, nil
	}
	if token == "'T" || token == "'F" || token == "true" || token == "false" {
		v := token == "'T" || token == "true"
		return &types.Value{E: types.Bool(v), T: types.TypeBool}, nil
	}
	if n, ok := p.intParser.ParseInt(token); ok {
		return &types.Value{E: n, T: types.TypeInt}, nil
	}
	if s, err := types.ParseString(token); err == nil {
		return &types.Value{E: s, T: types.TypeStr}, nil
	}
	// TODO
	return &types.Value{E: types.Ident(token), T: types.TypeUnknown}, nil
}

func (p *Parser) nextSexpr(leftBrace string, quoted bool) (*types.Value, error) {
	var list []types.Value
	for {
		token, err := p.nextToken()
		if err == io.EOF {
			return nil, UnexpectedEOF
		}
		if err != nil {
			return nil, err
		}
		if token == ")" {
			break
		}
		if token == "(" || token == "'(" || token == "\\(" {
			item, err := p.nextSexpr(token, quoted || token == "'(")
			if err != nil {
				return nil, err
			}
			list = append(list, *item)
			continue
		}
		par := p.tokenParam(token)
		list = append(list, *par)
	}

	return &types.Value{E: &types.Sexpr{
		List:   list,
		Quoted: quoted || leftBrace == "'(",
		Lambda: leftBrace == "\\(",
	}, T: types.TypeList}, nil
}

func (p *Parser) tokenParam(token string) *types.Value {
	if token == "'T" || token == "'F" || token == "true" || token == "false" {
		v := token == "'T" || token == "true"
		return &types.Value{E: types.Bool(v), T: types.TypeBool}
	}
	if n, ok := p.intParser.ParseInt(token); ok {
		return &types.Value{E: n, T: types.TypeInt}
	}
	if s, err := types.ParseString(token); err == nil {
		return &types.Value{E: s, T: types.TypeStr}
	}
	// TODO
	return &types.Value{E: types.Ident(token), T: types.TypeUnknown}
}

func (p *Parser) nextToken() (string, error) {
	if len(p.tokens) == 0 {
		if err := p.prepareTokens(); err != nil {
			return "", err
		}
	}
	token := p.tokens[0]
	p.tokens = p.tokens[1:]
	return token, nil
}

func (p *Parser) prepareTokens() error {
	if !p.scanner.Scan() {
		if err := p.scanner.Err(); err != nil {
			return err
		}
		return io.EOF
	}
	line := strings.TrimSpace(p.scanner.Text())
	if line == "" || line[0] == '#' || line[0] == ';' {
		return p.prepareTokens()
	}

	var token string
	var tokens []string
	inQuotes := false
	backslash := false
	for _, r := range line {
		if backslash {
			token += `\` + string(r)
			backslash = false
		} else if inQuotes {
			if r == '\\' {
				backslash = true
			} else {
				token += string(r)
				if r == '"' {
					inQuotes = false
					tokens = append(tokens, token)
					token = ""
				}
			}
		} else if r == '"' {
			token += string(r)
			inQuotes = true
		} else if unicode.IsSpace(r) {
			if inQuotes {
				token += string(r)
			} else if token != "" {
				tokens = append(tokens, token)
				token = ""
			}
		} else if r == '(' {
			if token == "'" {
				tokens = append(tokens, "'(")
			} else if token == "\\" {
				tokens = append(tokens, `\(`)
			} else if token != "" {
				tokens = append(tokens, token, "(")
			} else {
				tokens = append(tokens, "(")
			}
			token = ""
		} else if r == ')' {
			if token != "" {
				tokens = append(tokens, token)
				token = ""
			}
			tokens = append(tokens, ")")
		} else {
			token += string(r)
		}
	}
	if token != "" {
		tokens = append(tokens, token)
	}
	p.tokens = tokens
	return nil
}
