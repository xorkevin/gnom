package gnom

import (
	"errors"
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

	g := NewGrammarSymGenerator()

	n3 := g.NonTerm()
	n1 := g.NonTerm()
	n4 := g.NonTerm()
	t1 := g.Term()
	t3 := g.Term()

	nullableSet := newChangeIntSet()
	nullableSet.upsertAll(map[int]struct{}{
		n3.Kind(): {},
		n1.Kind(): {},
		n4.Kind(): {},
	})

	for n, c := range []struct {
		syms []GrammarSym
		exp  bool
	}{
		{
			syms: []GrammarSym{n3},
			exp:  true,
		},
		{
			syms: []GrammarSym{n3, n1, n4},
			exp:  true,
		},
		{
			syms: []GrammarSym{n3, t1, n4},
			exp:  false,
		},
		{
			syms: []GrammarSym{t3},
			exp:  false,
		},
	} {
		assert.Equalf(c.exp, isLL1Nullable(c.syms, nullableSet), "Fail case %d", n)
	}
}

func TestCalcLL1NullableSet(t *testing.T) {
	assert := assert.New(t)

	g := NewGrammarSymGenerator()

	n1 := g.NonTerm()
	n2 := g.NonTerm()
	n3 := g.NonTerm()
	n4 := g.NonTerm()
	t1 := g.Term()
	t2 := g.Term()

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

	g := NewGrammarSymGenerator()

	n1 := g.NonTerm()
	n2 := g.NonTerm()
	n3 := g.NonTerm()
	n4 := g.NonTerm()
	n5 := g.NonTerm()
	t1 := g.Term()
	t2 := g.Term()
	t3 := g.Term()
	t4 := g.Term()
	t5 := g.Term()

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

func TestCalcLL1FollowSet(t *testing.T) {
	assert := assert.New(t)

	g := NewGrammarSymGenerator()

	n1 := g.NonTerm()
	n2 := g.NonTerm()
	n3 := g.NonTerm()
	n4 := g.NonTerm()
	n5 := g.NonTerm()
	eof := g.Term()
	t1 := g.Term()
	t2 := g.Term()
	t3 := g.Term()
	t4 := g.Term()
	t5 := g.Term()

	for n, c := range []struct {
		rules     []GrammarRule
		followSet map[int]map[int]struct{}
	}{
		{
			rules: []GrammarRule{
				NewGrammarRule(n1.Kind(), []GrammarSym{t3, n2, t1}),
				NewGrammarRule(n1.Kind(), []GrammarSym{t3, n2, n3}),
				NewGrammarRule(n2.Kind(), []GrammarSym{t1, t2}),
				NewGrammarRule(n3.Kind(), []GrammarSym{n4}),
				NewGrammarRule(n3.Kind(), []GrammarSym{t4}),
				NewGrammarRule(n4.Kind(), []GrammarSym{}),
				NewGrammarRule(n5.Kind(), []GrammarSym{n1, t5}),
			},
			followSet: map[int]map[int]struct{}{
				n1.Kind(): {
					eof.Kind(): {},
					t5.Kind():  {},
				},
				n2.Kind(): {
					t1.Kind():  {},
					t4.Kind():  {},
					t5.Kind():  {},
					eof.Kind(): {},
				},
				n3.Kind(): {
					eof.Kind(): {},
					t5.Kind():  {},
				},
				n4.Kind(): {
					eof.Kind(): {},
					t5.Kind():  {},
				},
			},
		},
	} {
		nullSet := calcLL1NullableSet(c.rules)
		assert.Equalf(c.followSet, calcLL1FollowSet(c.rules, n1.Kind(), eof.Kind(), calcLL1FirstSet(c.rules, nullSet), nullSet).set, "Fail case %d", n)
	}
}

func TestCalcNewLL1Parser(t *testing.T) {
	assert := assert.New(t)

	g := NewGrammarSymGenerator()

	S := g.NonTerm()
	eof := g.Term()

	B := g.NonTerm()
	C := g.NonTerm()
	w := g.Term()
	x := g.Term()
	y := g.Term()
	z := g.Term()

	EP := g.NonTerm()
	T := g.NonTerm()
	TP := g.NonTerm()
	F := g.NonTerm()
	id := g.Term()
	plus := g.Term()
	star := g.Term()
	lparen := g.Term()
	rparen := g.Term()

	for n, c := range []struct {
		rules      []GrammarRule
		err        error
		parseTable map[int]map[int][]GrammarSym
	}{
		{
			rules: []GrammarRule{
				NewGrammarRule(S.Kind(), []GrammarSym{B, C}),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S.Kind(), []GrammarSym{y, C, z}),
				NewGrammarRule(B.Kind(), []GrammarSym{}),
				NewGrammarRule(B.Kind(), []GrammarSym{z, y, S}),
				NewGrammarRule(C.Kind(), []GrammarSym{y, B, C, C}),
				NewGrammarRule(C.Kind(), []GrammarSym{}),
				NewGrammarRule(C.Kind(), []GrammarSym{w, z}),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S.Kind(), []GrammarSym{B, x}),
				NewGrammarRule(B.Kind(), []GrammarSym{}),
				NewGrammarRule(B.Kind(), []GrammarSym{z, y, S, B}),
				NewGrammarRule(B.Kind(), []GrammarSym{z, B, z}),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S.Kind(), []GrammarSym{B, w, S, S}),
				NewGrammarRule(S.Kind(), []GrammarSym{}),
				NewGrammarRule(S.Kind(), []GrammarSym{y}),
				NewGrammarRule(B.Kind(), []GrammarSym{x}),
				NewGrammarRule(B.Kind(), []GrammarSym{}),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S.Kind(), []GrammarSym{T, EP}),
				NewGrammarRule(EP.Kind(), []GrammarSym{plus, T, EP}),
				NewGrammarRule(EP.Kind(), []GrammarSym{}),
				NewGrammarRule(T.Kind(), []GrammarSym{F, TP}),
				NewGrammarRule(TP.Kind(), []GrammarSym{star, F, TP}),
				NewGrammarRule(TP.Kind(), []GrammarSym{}),
				NewGrammarRule(F.Kind(), []GrammarSym{id}),
				NewGrammarRule(F.Kind(), []GrammarSym{lparen, S, rparen}),
			},
			parseTable: map[int]map[int][]GrammarSym{
				S.Kind(): {
					id.Kind():     {T, EP},
					lparen.Kind(): {T, EP},
				},
				EP.Kind(): {
					plus.Kind():   {plus, T, EP},
					rparen.Kind(): {},
					eof.Kind():    {},
				},
				T.Kind(): {
					id.Kind():     {F, TP},
					lparen.Kind(): {F, TP},
				},
				TP.Kind(): {
					plus.Kind():   {},
					star.Kind():   {star, F, TP},
					rparen.Kind(): {},
					eof.Kind():    {},
				},
				F.Kind(): {
					id.Kind():     {id},
					lparen.Kind(): {lparen, S, rparen},
				},
			},
		},
	} {
		parser, err := NewLL1Parser(c.rules, S.Kind(), eof.Kind())
		if c.err != nil {
			assert.Errorf(err, "Should fail to create parser: case %d", n)
			assert.Truef(errors.Is(err, c.err), "Should fail to create parser: case %d", n)
			continue
		}
		assert.NoErrorf(err, "Failed to create parser: case %d", n)
		assert.Equalf(c.parseTable, parser.table, "Fail case %d", n)
	}
}
