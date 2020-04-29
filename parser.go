package main

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"
)

var (
	UnexpectedEOF = errors.New("Unexpected EOF")
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
		if err == io.EOF {
			return nil, UnexpectedEOF
		}
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

	var token string
	var tokens []string
	inQuotes := false
	backslash := false
	for _, r := range line {
		if unicode.IsSpace(r) {
			if inQuotes {
				token += string(r)
			} else if token != "" {
				tokens = append(tokens, token)
				token = ""
			}
		} else if r == '(' {
			if token == "'" {
				tokens = append(tokens, "'(")
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
		} else if r == '"' {
			if !inQuotes {
				token += string(r)
				inQuotes = true
			} else if backslash {
				token += `\"`
				backslash = false
			} else {
				token += string(r)
				inQuotes = false
				tokens = append(tokens, token)
				token = ""
			}
		} else if r == '\\' {
			if backslash {
				token += `\\`
				backslash = false
			} else {
				backslash = true
			}
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
