package gnom

import (
	"errors"
	"fmt"
	"strings"
)

type (
	Dfa struct {
		kind  int
		nodes map[rune]*Dfa
	}
)

func NewDfa(kind int) *Dfa {
	return &Dfa{
		kind:  kind,
		nodes: map[rune]*Dfa{},
	}
}

func (d *Dfa) AddDfa(s []rune, dfa *Dfa) {
	for _, c := range s {
		d.nodes[c] = dfa
	}
}

func (d *Dfa) AddPath(path []rune, kind int, def int) *Dfa {
	if len(path) == 0 {
		d.kind = kind
		return d
	}
	c := path[0]
	path = path[1:]
	if _, ok := d.nodes[c]; !ok {
		d.nodes[c] = NewDfa(def)
	}
	return d.nodes[c].AddPath(path, kind, def)
}

func (d *Dfa) Match(c rune) (*Dfa, bool) {
	next, ok := d.nodes[c]
	if !ok {
		return nil, false
	}
	return next, true
}

func (d *Dfa) Kind() int {
	return d.kind
}

type (
	Token struct {
		kind int
		val  string
	}
)

func NewToken(kind int, val string) Token {
	return Token{
		kind: kind,
		val:  val,
	}
}

func (t *Token) Kind() int {
	return t.kind
}

func (t *Token) Val() string {
	return t.val
}

type (
	DfaLexer struct {
		dfa     *Dfa
		def     int
		eof     int
		ignored map[int]struct{}
	}
)

func NewDfaLexer(dfa *Dfa, def, eof int, ignored map[int]struct{}) *DfaLexer {
	return &DfaLexer{
		dfa:     dfa,
		def:     def,
		eof:     eof,
		ignored: ignored,
	}
}

var (
	ErrLex = errors.New("lexer error")
)

func minInt(a, b int) int {
	if b < a {
		return b
	}
	return a
}

func (l *DfaLexer) Next(chars []rune) (*Token, []rune, error) {
	s := &strings.Builder{}
	n := l.dfa
	for {
		if len(chars) == 0 {
			break
		}
		c := chars[0]
		next, ok := n.Match(c)
		if !ok {
			break
		}
		n = next
		s.WriteRune(c)
		chars = chars[1:]
	}
	if n.Kind() == l.def {
		if s.Len() == 0 && len(chars) == 0 {
			return &Token{kind: l.eof, val: ""}, chars, nil
		}
		return nil, nil, fmt.Errorf("Invalid tokens: %s: %w", s.String()+string(chars[:minInt(8, len(chars))]), ErrLex)
	}
	return &Token{kind: n.Kind(), val: s.String()}, chars, nil
}

func (l *DfaLexer) Tokenize(chars []rune) ([]Token, error) {
	tokens := []Token{}
	for {
		t, next, err := l.Next(chars)
		if err != nil {
			return nil, err
		}
		if _, ok := l.ignored[t.Kind()]; !ok {
			tokens = append(tokens, *t)
		}
		chars = next
		if t.Kind() == l.eof {
			return tokens, nil
		}
	}
}
