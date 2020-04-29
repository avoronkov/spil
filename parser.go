package main

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type Parser struct {
	scanner *bufio.Scanner
	tokens  []string
}

func NewParser(r io.Reader) *Parser {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return &Parser{
		scanner: scanner,
	}
}

func (p *Parser) NextExpr() (Expr, error) {
	token, err := p.nextToken()
	if err != nil {
		return nil, err
	}
	if token == "(" || token == "'(" {
		item, err := p.nextSexpr()
		if err != nil {
			return nil, err
		}
		if token == "'(" {
			item.Quoted = true
		}
		return item, nil
	}
	if token == "'T" {
		return Bool(true), nil
	}
	if token == "'F" {
		return Bool(false), nil
	}
	if n, err := strconv.Atoi(token); err == nil {
		return Int(n), nil
	}
	if strings.HasPrefix(token, `"`) && strings.HasPrefix(token, `"`) {
		return Str(token[1 : len(token)-1]), nil
	}
	return Ident(token), nil
}

func (p *Parser) nextSexpr() (*Sexpr, error) {
	var list []Expr
	for {
		token, err := p.nextToken()
		if err != nil {
			return nil, err
		}
		if token == ")" {
			break
		}
		if token == "(" || token == "'(" {
			item, err := p.nextSexpr()
			if err != nil {
				return nil, err
			}
			if token == "'(" {
				item.Quoted = true
			}
			list = append(list, item)
			continue
		}
		if token == "'T" {
			list = append(list, Bool(true))
			continue
		}
		if token == "'F" {
			list = append(list, Bool(false))
			continue
		}
		if n, err := strconv.Atoi(token); err == nil {
			list = append(list, Int(n))
			continue
		}
		if strings.HasPrefix(token, `"`) && strings.HasPrefix(token, `"`) {
			list = append(list, Str(token[1:len(token)-1]))
			continue
		}
		list = append(list, Ident(token))
	}
	return &Sexpr{List: list}, nil
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
	if line == "" || line[0] == '#' {
		return p.prepareTokens()
	}
	fields := strings.Fields(line)
	var tokens []string
	for _, field := range fields {
		tokens = append(tokens, p.splitField(field)...)
	}
	p.tokens = tokens
	return nil
}

// '(foo' -> '(', 'foo'
// 'foo)' -> 'foo', ')'
// '(foo)' -> '(', 'foo', ')'
// `'(x` -> `'(`, `x`
func (p *Parser) splitField(field string) (tokens []string) {
	if strings.HasPrefix(field, "(") {
		tokens = append(tokens, "(")
		field = field[1:]
	} else if strings.HasPrefix(field, "'(") {
		tokens = append(tokens, "'(")
		field = field[2:]
	}
	if strings.HasSuffix(field, ")") {
		if len(field) > 1 {
			tokens = append(tokens, field[:len(field)-1], ")")
		} else {
			tokens = append(tokens, ")")
		}
	} else if len(field) > 0 {
		tokens = append(tokens, field)
	}
	return
}
