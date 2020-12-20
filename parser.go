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

func NewGrammarRule(from int, to []GrammarSym) GrammarRule {
	return GrammarRule{
		from: from,
		to:   to,
	}
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

func (s *changeIntIntSet) contains(a, b int) bool {
	v, ok := s.set[a]
	if !ok {
		return false
	}
	_, ok = v[b]
	return ok
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
	LL1Parser struct {
		table map[int]map[int][]GrammarSym
	}
)

func NewLL1Parser(rules []GrammarRule, start, eof int) (*LL1Parser, error) {
	nonTerminals := newChangeIntSet()
	for _, i := range rules {
		nonTerminals.upsert(i.from)
	}
	for _, i := range rules {
		for _, j := range i.to {
			if !j.term && nonTerminals.contains(j.kind) {
				return nil, fmt.Errorf("Nonterminal lacks production rule: %d: %w", j.kind, ErrGrammar)
			}
		}
	}
	nullableSet := calcLL1NullableSet(rules)
	firstSet := calcLL1FirstSet(rules, nullableSet)
	followSet := calcLL1FollowSet(rules, start, eof, firstSet, nullableSet)

	parseTable, err := calcLL1ParseTable(rules, nonTerminals, nullableSet, firstSet, followSet)
	if err != nil {
		return nil, err
	}

	return &LL1Parser{
		table: parseTable,
	}, nil
}
