// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT 
// license that can be found in the LICENSE file.

package selfml

import (
	"errors"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokens for opening and closing a S-expr.
const sexprOpen = '('
const sexprClose = ')'
const endOfLine = '\n'

// Interface for representing a generic element in a S-expr.
type selfValue interface {
	String(int) string
}

// String value in a S-expr.
type selfString string

// S-expr value in a S-expr, must start with a selfString.
type selfNode struct {
	head   selfString
	values []selfValue
}

// Holds the parser state.
type selfParser struct {
	input      string
	pos        int
	lineNumber uint
	r          rune
	runeWidth  int
	eod        bool
}

// Returned value after parsing a self-ml string.
type Tree map[string]interface{}

type parseFunc func() (selfValue, error)

// Converts a selfString into a printable string.
func (s selfString) String(_ int) string {
	return string(s)
}

// Root node has no head by convention.
func (node selfNode) isRoot() bool {
	return len(node.head) == 0
}

// Converts a selfNode into a printable string with indentation.
func (node selfNode) String(indent int) (str string) {
	// Root node needs no delimitors
	if !node.isRoot() {
		str += string(sexprOpen) + node.head.String(indent)
	} else {
		indent -= 1
	}

	if len(node.values) > 0 {
		for _, v := range node.values {
			str += "\n" + strings.Repeat("    ", indent+1) + v.String(indent+1)
		}
	}

	if !node.isRoot() {
		str += string(sexprClose) + "\n"
	}
	return
}

func (p *selfParser) next() {
	p.pos += p.runeWidth
	if p.pos >= len(p.input) {
		p.eod = true
		return
	}

	if p.r == endOfLine {
		p.lineNumber += 1
	}

	if p.r, p.runeWidth = utf8.DecodeRuneInString(p.input[p.pos:]); p.r == utf8.RuneError {
		panic("bad rune")
	}
}

func isComment(r rune) bool {
	return r == ';' || r == '#'
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r)
}

func isStringChar(r rune) bool {
	if isSpace(r) {
		return false
	}

	switch r {
	case '[', ']', '(', ')', '{', '}', '\\':
		return false
	default:
		return true
	}
}

func (p *selfParser) skipLine() {
	//println("skipLine")
	for !p.eod && p.r != endOfLine {
		p.next()
	}
}

func (p *selfParser) skipSpaces() {
	for !p.eod && (isComment(p.r) || isSpace(p.r)) {
		if isComment(p.r) {
			p.skipLine()
		} else {
			p.next()
		}
	}
}

func (p *selfParser) parseBacktickString() (selfString, error) {
	var (
		str string = ""
		prev rune = -1
	)

	for {
		if p.r != '`' && prev == '`' { break }

		if p.r == '`' {
			if prev == '`' {
				str += "`"
				prev = -1
			}
		} else {
			str += string(p.r)
		}

		prev = p.r
		p.next()
	}

	return selfString(str), nil
}

func (p *selfParser) parseString() (selfString, error) {
	//println("parseString")
	var str string = ""

	if p.eod {
		return "", errors.New("eod")
	}

	switch p.r {
	case '`':
		p.next()
		return p.parseBacktickString()
	case '[':
		return "", nil
	default:
		for isStringChar(p.r) {
			str += string(p.r)
			p.next()
		}
	}

	return selfString(str), nil
}

func (p *selfParser) parseNodeBody() (values []selfValue, err error) {
	//println("parseNodeBody")
	var (
		v          selfValue
		parseValue parseFunc
	)
	values = make([]selfValue, 0)

	p.skipSpaces()
	for !p.eod && p.r != sexprClose {
		if p.r == sexprOpen {
			parseValue = func() (selfValue, error) { return p.parseNode() }
		} else {
			parseValue = func() (selfValue, error) { return p.parseString() }
		}

		if v, err = parseValue(); err != nil {
			return nil, err
		} else {
			values = append(values, v)
		}

		p.skipSpaces()
	}

	if p.r == sexprClose {
		p.next()
	}

	return
}

func (p *selfParser) parseNode() (node *selfNode, err error) {
	//println("parseNode")
	var nodeName selfString

	p.skipSpaces()
	if p.r != sexprOpen {
		return nil, errors.New("expected ( char")
	}
	p.next()

	nodeName, err = p.parseString()
	if err != nil {
		return nil, err
	}

	node = &selfNode{head: nodeName}
	if node.values, err = p.parseNodeBody(); err != nil {
		return nil, err
	}

	return
}

func (node *selfNode) reduce() (tree *Tree, err error) {
	return
}

// Parses a self-ml string and returns its Tree structure.
func LoadString(data string) (tree *Tree, err error) {
	p := selfParser{input: data, r: '\n'}
	rootNode := selfNode{head: ""}
	if rootNode.values, err = p.parseNodeBody(); err != nil {
		return
	}

	println(rootNode.String(0))
	tree, err = rootNode.reduce()
	return
}

// Parses a self-ml file on disk and returns its Tree structure.
func Load(path string) (tree *Tree, err error) {
	var bytes []byte
	if bytes, err = ioutil.ReadFile(path); err != nil {
		return
	}

	tree, err = LoadString(string(bytes))
	return
}
