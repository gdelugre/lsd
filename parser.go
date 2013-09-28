// Copyright (c) 2013 Guillaume Delugré.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package selfml

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Tokens for opening and closing a S-expr.
const sexprOpen = '('
const sexprClose = ')'
const endOfLine = '\n'

// Structure returned when a parsing error occurs.
type parserError struct {
	message string
	lineNumber uint
}

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
	root bool
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

// Generic type function for parsing selfValue.
type parseFunc func() (selfValue, error)

// Error printing.
func (err *parserError) Error() string {
	return fmt.Sprintf("Error while parsing self-ml: %s (line %d)", err.message, err.lineNumber)
}

// Error generator.
func (p *selfParser) newError(str string) error {
	return &parserError{message: str, lineNumber: p.lineNumber}
}

// Error generator.
// Overrides current line number of parser.
func (p *selfParser) newErrorAtLine(str string, lineNum uint) error {
	return &parserError{message: str, lineNumber: lineNum}
}

// Converts a selfString into a printable string.
func (s selfString) String(_ int) string {
	if len(s) == 0 {
		return "[]"
	} else if strings.ContainsAny(string(s), "`#;\\([{}])") {
		return "`" + strings.Replace(string(s), "`", "``", -1) + "`"
	} else {
		return string(s)
	}
}

// Root node has special properties.
// It can only contain subnodes and must not start or end with S-expr delimitors.
func (node selfNode) isRoot() bool {
	return node.root
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

func (node *selfNode) getNodeByName(name string) *selfNode {
	for _, n := range node.values {
		switch n.(type) {
		case selfNode:
			subNode := n.(selfNode)
			if string(subNode.head) == name {
				return &subNode
			}
		}
	}
	return nil
}

func (p *selfParser) next() {
	p.pos += p.runeWidth
	if p.pos >= len(p.input) {
		p.eod = true
		return
	}

	if p.r == endOfLine {
		p.lineNumber++
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
		str  string = ""
		prev rune   = -1
		lineNum uint = p.lineNumber
	)

	for !p.eod {
		if p.r != '`' && prev == '`' {
			break
		}

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

	if p.eod {
		return "", p.newErrorAtLine("unexpected end of data while parsing string", lineNum)
	} else {
		return selfString(str), nil
	}
}

func (p *selfParser) parseBracketedString() (selfString, error) {
	level := 1
	str := ""
	lineNum := p.lineNumber

	for !p.eod {
		if p.r == ']' {
			level--
			if level == 0 {
				p.next()
				break
			}
		}

		if p.r == '[' {
			level++
		}

		str += string(p.r)
		p.next()
	}

	if p.eod {
		return "", p.newErrorAtLine("unexpected end of data while parsing string", lineNum)
	} else {
		return selfString(str), nil
	}
}

func (p *selfParser) parseString() (selfString, error) {
	var str string = ""

	if p.eod {
		return "", p.newError("unexpected end of data")
	}

	switch p.r {
	case '`':
		p.next()
		return p.parseBacktickString()
	case '[':
		p.next()
		return p.parseBracketedString()
	default:
		for isStringChar(p.r) {
			str += string(p.r)
			p.next()
		}
	}

	return selfString(str), nil
}

func (p *selfParser) parseNodeBody(rootNode bool) (values []selfValue, err error) {
	var (
		v          selfValue
		parseValue parseFunc
	)
	values = make([]selfValue, 0)

	p.skipSpaces()
	for !p.eod && p.r != sexprClose {
		if p.r == sexprOpen {
			parseValue = func() (selfValue, error) { return p.parseNode() }
		} else if !rootNode {
			parseValue = func() (selfValue, error) { return p.parseString() }
		} else {
			return nil, p.newError("Unexpected string in root node")
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
	var nodeName selfString

	p.skipSpaces()
	if p.r != sexprOpen {
		return nil, p.newError("expected `(` rune at start of list")
	}
	p.next()

	nodeName, err = p.parseString()
	if err != nil {
		return nil, err
	}

	node = &selfNode{head: nodeName}
	if node.values, err = p.parseNodeBody(false); err != nil {
		return nil, err
	}

	return
}
