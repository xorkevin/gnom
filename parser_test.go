package gnom

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGrammarSym(t *testing.T) {
	assert := assert.New(t)

	for _, c := range []struct {
		term bool
		kind int
	}{
		{
			term: true,
			kind: 3,
		},
		{
			term: false,
			kind: 1,
		},
		{
			term: true,
			kind: 4,
		},
	} {
		var sym GrammarSym
		if c.term {
			sym = NewGrammarTerm(c.kind)
		} else {
			sym = NewGrammarNonTerm(c.kind)
		}
		assert.Equalf(c.term, sym.Term(), "Invalid term: %t %d", c.term, c.kind)
		assert.Equalf(c.kind, sym.Kind(), "Invalid kind: %t %d", c.term, c.kind)
	}
}

func TestIsLL1Nullable(t *testing.T) {
	assert := assert.New(t)

	nullableSet := newChangeIntSet()
	nullableSet.upsertAll(map[int]struct{}{
		3: {},
		1: {},
		4: {},
	})

	for n, c := range []struct {
		syms []GrammarSym
		exp  bool
	}{
		{
			syms: []GrammarSym{NewGrammarNonTerm(3)},
			exp:  true,
		},
		{
			syms: []GrammarSym{NewGrammarNonTerm(3), NewGrammarNonTerm(1), NewGrammarNonTerm(4)},
			exp:  true,
		},
		{
			syms: []GrammarSym{NewGrammarNonTerm(3), NewGrammarTerm(1), NewGrammarNonTerm(4)},
			exp:  false,
		},
		{
			syms: []GrammarSym{NewGrammarTerm(3)},
			exp:  false,
		},
	} {
		assert.Equalf(c.exp, isLL1Nullable(c.syms, nullableSet), "Fail case %d", n)
	}
}

func TestCalcLL1NullableSet(t *testing.T) {
	assert := assert.New(t)

	n1 := NewGrammarNonTerm(1)
	n2 := NewGrammarNonTerm(2)
	n3 := NewGrammarNonTerm(3)
	n4 := NewGrammarNonTerm(4)
	t1 := NewGrammarTerm(1)
	t2 := NewGrammarTerm(2)

	for n, c := range []struct {
		rules   []GrammarRule
		nullSet map[int]struct{}
	}{
		{
			rules: []GrammarRule{
				NewGrammarRule(n1.Kind(), []GrammarSym{t1}),
				NewGrammarRule(n1.Kind(), []GrammarSym{n2, t2}),
				NewGrammarRule(n1.Kind(), []GrammarSym{n3, n4}),
				NewGrammarRule(n2.Kind(), []GrammarSym{t1, t2}),
				NewGrammarRule(n3.Kind(), []GrammarSym{n4}),
				NewGrammarRule(n4.Kind(), []GrammarSym{}),
			},
			nullSet: map[int]struct{}{
				n3.Kind(): {},
				n1.Kind(): {},
				n4.Kind(): {},
			},
		},
	} {
		assert.Equalf(c.nullSet, calcLL1NullableSet(c.rules).iter(), "Fail case %d", n)
	}
}

func TestCalcLL1FirstSet(t *testing.T) {
	assert := assert.New(t)

	n1 := NewGrammarNonTerm(1)
	n2 := NewGrammarNonTerm(2)
	n3 := NewGrammarNonTerm(3)
	n4 := NewGrammarNonTerm(4)
	n5 := NewGrammarNonTerm(5)
	t1 := NewGrammarTerm(1)
	t2 := NewGrammarTerm(2)
	t3 := NewGrammarTerm(3)
	t4 := NewGrammarTerm(4)
	t5 := NewGrammarTerm(5)

	for n, c := range []struct {
		rules    []GrammarRule
		firstSet map[int]map[int]struct{}
	}{
		{
			rules: []GrammarRule{
				NewGrammarRule(n1.Kind(), []GrammarSym{t3}),
				NewGrammarRule(n1.Kind(), []GrammarSym{n3, t1}),
				NewGrammarRule(n2.Kind(), []GrammarSym{t1, t2}),
				NewGrammarRule(n2.Kind(), []GrammarSym{n4, n5, t2}),
				NewGrammarRule(n3.Kind(), []GrammarSym{n4}),
				NewGrammarRule(n3.Kind(), []GrammarSym{t4}),
				NewGrammarRule(n4.Kind(), []GrammarSym{}),
				NewGrammarRule(n5.Kind(), []GrammarSym{t5}),
			},
			firstSet: map[int]map[int]struct{}{
				n1.Kind(): {
					t3.Kind(): {},
					t1.Kind(): {},
					t4.Kind(): {},
				},
				n2.Kind(): {
					t1.Kind(): {},
					t5.Kind(): {},
				},
				n3.Kind(): {
					t4.Kind(): {},
				},
				n5.Kind(): {
					t5.Kind(): {},
				},
			},
		},
	} {
		assert.Equalf(c.firstSet, calcLL1FirstSet(c.rules, calcLL1NullableSet(c.rules)).set, "Fail case %d", n)
	}
}
