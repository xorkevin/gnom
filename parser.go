package gnom

import (
	"errors"
	"fmt"
)

var (
	ErrGrammar = errors.New("grammar error")
)

type (
	Parser struct {
		table map[int]map[int][]GrammarSym
	}

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

func NewGrammarRule(from int, to []GrammarSym) GrammarRule {
	return GrammarRule{
		from: from,
		to:   to,
	}
}

func NewParser(rules []GrammarRule, start, eof int) (*Parser, error) {
	// nullable set
	nonTerminals := map[int]struct{}{}
	nullableSet := map[int]struct{}{}
	for _, i := range rules {
		nonTerminals[i.from] = struct{}{}
		if len(i.to) == 0 {
			nullableSet[i.from] = struct{}{}
			continue
		}
	}
	for {
		change := false
	nullouter:
		for _, i := range rules {
			for _, j := range i.to {
				if j.term {
					continue nullouter
				}
				if _, ok := nullableSet[j.kind]; !ok {
					continue nullouter
				}
			}
			if _, ok := nullableSet[i.from]; !ok {
				change = true
				nullableSet[i.from] = struct{}{}
			}
		}
		if !change {
			break
		}
	}

	firstSet := map[int]map[int]struct{}{}
	followSet := map[int]map[int]struct{}{}
	parseTable := map[int]map[int][]GrammarSym{}
	for nt := range nonTerminals {
		firstSet[nt] = map[int]struct{}{}
		followSet[nt] = map[int]struct{}{}
		parseTable[nt] = map[int][]GrammarSym{}
	}

	// first set
	for _, i := range rules {
		if len(i.to) != 0 && i.to[0].term {
			firstSet[i.from][i.to[0].kind] = struct{}{}
		}
	}
	for {
		change := false
		for _, i := range rules {
			for _, j := range i.to {
				if j.term {
					if _, ok := firstSet[i.from][j.kind]; !ok {
						change = true
						firstSet[i.from][j.kind] = struct{}{}
					}
					break
				}
				for k := range firstSet[j.kind] {
					if _, ok := firstSet[i.from][k]; !ok {
						change = true
						firstSet[i.from][k] = struct{}{}
					}
				}
				if _, ok := nullableSet[j.kind]; !ok {
					break
				}
			}
		}
		if !change {
			break
		}
	}

	// follow set
	followSet[start][eof] = struct{}{}
	for {
		change := false
		for _, i := range rules {
		followouter:
			for n, j := range i.to {
				if j.term {
					continue
				}
				for _, k := range i.to[n+1:] {
					if k.term {
						if _, ok := followSet[j.kind][k.kind]; !ok {
							change = true
							followSet[j.kind][k.kind] = struct{}{}
						}
						continue followouter
					}
					for l := range firstSet[k.kind] {
						if _, ok := followSet[j.kind][l]; !ok {
							change = true
							followSet[j.kind][l] = struct{}{}
						}
					}
					if _, ok := nullableSet[k.kind]; !ok {
						continue followouter
					}
				}
				for k := range followSet[i.from] {
					if _, ok := followSet[j.kind][k]; !ok {
						change = true
						followSet[j.kind][k] = struct{}{}
					}
				}
			}
		}
		if !change {
			break
		}
	}

	// parsing table
	for _, i := range rules {
		for _, j := range i.to {
			if j.term {
				if _, ok := parseTable[i.from][j.kind]; ok {
					return nil, fmt.Errorf("Grammar is not LL1: %w", ErrGrammar)
				}
				parseTable[i.from][j.kind] = i.to
				break
			}
			for k := range firstSet[j.kind] {
				if _, ok := parseTable[i.from][k]; ok {
					return nil, fmt.Errorf("Grammar is not LL1: %w", ErrGrammar)
				}
				parseTable[i.from][k] = i.to
			}
			if _, ok := nullableSet[j.kind]; !ok {
				break
			}
		}
		if _, ok := nullableSet[i.from]; !ok {
			continue
		}
		for j := range followSet[i.from] {
			if _, ok := parseTable[i.from][j]; ok {
				return nil, fmt.Errorf("Grammar is not LL1: %w", ErrGrammar)
			}
			parseTable[i.from][j] = i.to
		}
	}

	return &Parser{
		table: parseTable,
	}, nil
}
