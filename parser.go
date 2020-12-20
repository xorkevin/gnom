package gnom

import (
	"errors"
	"fmt"
)

var (
	ErrGrammar = errors.New("grammar error")
)

type (
	GrammarSym struct {
		term bool
		kind int
	}

	GrammarSymGenerator struct {
		i int
	}

	GrammarRule struct {
		from int
		to   []GrammarSym
	}
)

func NewGrammarTerm(kind int) GrammarSym {
	return GrammarSym{
		term: true,
		kind: kind,
	}
}

func NewGrammarNonTerm(kind int) GrammarSym {
	return GrammarSym{
		term: false,
		kind: kind,
	}
}

func (s *GrammarSym) Term() bool {
	return s.term
}

func (s *GrammarSym) Kind() int {
	return s.kind
}

func NewGrammarRule(from GrammarSym, to ...GrammarSym) GrammarRule {
	return GrammarRule{
		from: from.Kind(),
		to:   to,
	}
}

func NewGrammarSymGenerator() *GrammarSymGenerator {
	return &GrammarSymGenerator{
		i: 0,
	}
}

func (g *GrammarSymGenerator) Term() GrammarSym {
	s := NewGrammarTerm(g.i)
	g.i++
	return s
}

func (g *GrammarSymGenerator) NonTerm() GrammarSym {
	s := NewGrammarNonTerm(g.i)
	g.i++
	return s
}

type (
	changeIntSet struct {
		change bool
		set    map[int]struct{}
	}
)

func newChangeIntSet() *changeIntSet {
	return &changeIntSet{
		change: false,
		set:    map[int]struct{}{},
	}
}

func (s *changeIntSet) resetChanged() {
	s.change = false
}

func (s *changeIntSet) isChanged() bool {
	return s.change
}

func (s *changeIntSet) iter() map[int]struct{} {
	return s.set
}

func (s *changeIntSet) contains(k int) bool {
	_, ok := s.set[k]
	return ok
}

func (s *changeIntSet) upsert(k int) {
	if _, ok := s.set[k]; !ok {
		s.change = true
		s.set[k] = struct{}{}
	}
}

func (s *changeIntSet) upsertAll(m map[int]struct{}) {
	for k := range m {
		s.upsert(k)
	}
}

type (
	changeIntIntSet struct {
		change bool
		set    map[int]map[int]struct{}
	}
)

func newChangeIntIntSet() *changeIntIntSet {
	return &changeIntIntSet{
		change: false,
		set:    map[int]map[int]struct{}{},
	}
}

func (s *changeIntIntSet) resetChanged() {
	s.change = false
}

func (s *changeIntIntSet) isChanged() bool {
	return s.change
}

func (s *changeIntIntSet) iter(a int) map[int]struct{} {
	if v, ok := s.set[a]; ok {
		return v
	}
	return map[int]struct{}{}
}

func (s *changeIntIntSet) upsert(a, b int) {
	if _, ok := s.set[a]; !ok {
		s.set[a] = map[int]struct{}{}
	}
	if _, ok := s.set[a][b]; !ok {
		s.change = true
		s.set[a][b] = struct{}{}
	}
}

func (s *changeIntIntSet) upsertAll(a int, m map[int]struct{}) {
	for k := range m {
		s.upsert(a, k)
	}
}

func isLL1Nullable(syms []GrammarSym, nullableSet *changeIntSet) bool {
	for _, i := range syms {
		if i.term {
			return false
		}
		if !nullableSet.contains(i.kind) {
			return false
		}
	}
	return true
}

func calcLL1NullableSet(rules []GrammarRule) *changeIntSet {
	set := newChangeIntSet()
	for {
		set.resetChanged()
		for _, i := range rules {
			if isLL1Nullable(i.to, set) {
				set.upsert(i.from)
			}
		}
		if !set.isChanged() {
			return set
		}
	}
}

func calcLL1First(syms []GrammarSym, firstSet *changeIntIntSet, nullableSet *changeIntSet) map[int]struct{} {
	set := newChangeIntSet()
	for _, i := range syms {
		if i.term {
			set.upsert(i.kind)
			return set.iter()
		}
		set.upsertAll(firstSet.iter(i.kind))
		if !nullableSet.contains(i.kind) {
			return set.iter()
		}
	}
	return set.iter()
}

func calcLL1FirstSet(rules []GrammarRule, nullableSet *changeIntSet) *changeIntIntSet {
	set := newChangeIntIntSet()
	for {
		set.resetChanged()
		for _, i := range rules {
			set.upsertAll(i.from, calcLL1First(i.to, set, nullableSet))
		}
		if !set.isChanged() {
			return set
		}
	}
}

func calcLL1FollowSet(rules []GrammarRule, start, eof int, firstSet *changeIntIntSet, nullableSet *changeIntSet) *changeIntIntSet {
	set := newChangeIntIntSet()
	set.upsert(start, eof)
	for {
		set.resetChanged()
		for _, i := range rules {
			for n, j := range i.to {
				if j.term {
					continue
				}
				b := i.to[n+1:]
				set.upsertAll(j.kind, calcLL1First(b, firstSet, nullableSet))
				if isLL1Nullable(b, nullableSet) {
					set.upsertAll(j.kind, set.iter(i.from))
				}
			}
		}
		if !set.isChanged() {
			return set
		}
	}
}

func calcLL1ParseTable(rules []GrammarRule, nonTerminals *changeIntSet, nullableSet *changeIntSet, firstSet *changeIntIntSet, followSet *changeIntIntSet) (map[int]map[int][]GrammarSym, error) {
	table := map[int]map[int][]GrammarSym{}
	for nt := range nonTerminals.iter() {
		table[nt] = map[int][]GrammarSym{}
	}
	for n, i := range rules {
		for j := range calcLL1First(i.to, firstSet, nullableSet) {
			if _, ok := table[i.from][j]; ok {
				return nil, fmt.Errorf("Grammar is not LL1: duplicate rule: %d: %w", n, ErrGrammar)
			}
			table[i.from][j] = i.to
		}
		if isLL1Nullable(i.to, nullableSet) {
			for j := range followSet.iter(i.from) {
				if _, ok := table[i.from][j]; ok {
					return nil, fmt.Errorf("Grammar is not LL1: duplicate rule: %d: %w", n, ErrGrammar)
				}
				table[i.from][j] = i.to
			}
		}
	}
	return table, nil
}

type (
	TokenStack struct {
		tokens []Token
		buf    []Token
	}
)

func NewTokenStack(tokens []Token) *TokenStack {
	return &TokenStack{
		tokens: tokens,
		buf:    []Token{},
	}
}

func (s *TokenStack) Empty() bool {
	return len(s.tokens) == 0 && len(s.buf) == 0
}

func (s *TokenStack) Pop() (Token, bool) {
	if len(s.buf) == 0 {
		if len(s.tokens) == 0 {
			return Token{}, false
		}
		k := s.tokens[0]
		s.tokens = s.tokens[1:]
		return k, true
	}
	k := s.buf[len(s.buf)-1]
	s.buf = s.buf[:len(s.buf)-1]
	return k, true
}

func (s *TokenStack) Peek() (Token, bool) {
	if len(s.buf) == 0 {
		if len(s.tokens) == 0 {
			return Token{}, false
		}
		return s.tokens[0], true
	}
	return s.buf[len(s.buf)-1], true
}

func (s *TokenStack) Push(t Token) {
	s.buf = append(s.buf, t)
}

type (
	ParseTree struct {
		sym      GrammarSym
		token    Token
		children []ParseTree
	}
)

func NewParseTree(sym GrammarSym) *ParseTree {
	return &ParseTree{
		sym:      sym,
		children: []ParseTree{},
	}
}

func NewParseTreeLeaf(sym GrammarSym, token Token) *ParseTree {
	return &ParseTree{
		sym:   sym,
		token: token,
	}
}

func (t *ParseTree) Term() bool {
	return t.sym.Term()
}

func (t *ParseTree) Kind() int {
	return t.sym.Kind()
}

func (t *ParseTree) Val() string {
	return t.token.Val()
}

func (t *ParseTree) Children() []ParseTree {
	return t.children
}

func (t *ParseTree) AddChild(child ParseTree) {
	t.children = append(t.children, child)
}

type (
	LL1Parser struct {
		table map[int]map[int][]GrammarSym
		start GrammarSym
		eof   GrammarSym
	}
)

func NewLL1Parser(rules []GrammarRule, start, eof GrammarSym) (*LL1Parser, error) {
	nonTerminals := newChangeIntSet()
	for _, i := range rules {
		nonTerminals.upsert(i.from)
	}
	for _, i := range rules {
		for _, j := range i.to {
			if !j.term && !nonTerminals.contains(j.kind) {
				return nil, fmt.Errorf("Nonterminal lacks production rule: %d: %w", j.kind, ErrGrammar)
			}
		}
	}
	nullableSet := calcLL1NullableSet(rules)
	firstSet := calcLL1FirstSet(rules, nullableSet)
	followSet := calcLL1FollowSet(rules, start.Kind(), eof.Kind(), firstSet, nullableSet)

	parseTable, err := calcLL1ParseTable(rules, nonTerminals, nullableSet, firstSet, followSet)
	if err != nil {
		return nil, err
	}

	return &LL1Parser{
		table: parseTable,
		start: start,
		eof:   eof,
	}, nil
}

func (p *LL1Parser) getProduction(nt, t int) ([]GrammarSym, bool) {
	r, ok := p.table[nt]
	if !ok {
		return nil, false
	}
	prod, ok := r[t]
	if !ok {
		return nil, false
	}
	return prod, true
}

var (
	ErrParse = errors.New("parser error")
)

func (p *LL1Parser) parseSym(sym GrammarSym, ts *TokenStack) (*ParseTree, error) {
	if sym.Term() {
		token, ok := ts.Pop()
		if !ok {
			return nil, fmt.Errorf("Unexpected end of token stream: %w", ErrParse)
		}
		if sym.Kind() != token.Kind() {
			return nil, fmt.Errorf("Unexpected token: %s: %w", token.Val(), ErrParse)
		}
		return NewParseTreeLeaf(sym, token), nil
	}
	next, ok := ts.Peek()
	if !ok {
		return nil, fmt.Errorf("Unexpected end of token stream: %w", ErrParse)
	}
	prod, ok := p.getProduction(sym.Kind(), next.Kind())
	if !ok {
		return nil, fmt.Errorf("Unexpected token: %s: %w", next.Val(), ErrParse)
	}
	tree := NewParseTree(sym)
	for _, i := range prod {
		n, err := p.parseSym(i, ts)
		if err != nil {
			return nil, err
		}
		tree.AddChild(*n)
	}
	return tree, nil
}

func (p *LL1Parser) Parse(tokens []Token) (*ParseTree, error) {
	ts := NewTokenStack(tokens)
	root, err := p.parseSym(p.start, ts)
	if err != nil {
		return nil, err
	}
	last, ok := ts.Peek()
	if !ok {
		return nil, fmt.Errorf("Unexpected end of token stream: %w", ErrParse)
	}
	if p.eof.Kind() != last.Kind() {
		return nil, fmt.Errorf("Unexpected token: %s: %w", last.Val(), ErrParse)
	}
	return root, nil
}
