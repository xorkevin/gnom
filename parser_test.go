package gnom

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
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
				NewGrammarRule(n1, t1),
				NewGrammarRule(n1, n2, t2),
				NewGrammarRule(n1, n3, n4),
				NewGrammarRule(n2, t1, t2),
				NewGrammarRule(n3, n4),
				NewGrammarRule(n4),
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
				NewGrammarRule(n1, t3),
				NewGrammarRule(n1, n3, t1),
				NewGrammarRule(n2, t1, t2),
				NewGrammarRule(n2, n4, n5, t2),
				NewGrammarRule(n3, n4),
				NewGrammarRule(n3, t4),
				NewGrammarRule(n4),
				NewGrammarRule(n5, t5),
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
				NewGrammarRule(n1, t3, n2, t1),
				NewGrammarRule(n1, t3, n2, n3),
				NewGrammarRule(n2, t1, t2),
				NewGrammarRule(n3, n4),
				NewGrammarRule(n3, t4),
				NewGrammarRule(n4),
				NewGrammarRule(n5, n1, t5),
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

func TestNewLL1Parser(t *testing.T) {
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
				NewGrammarRule(S, B, C),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S, y, C, z),
				NewGrammarRule(B),
				NewGrammarRule(B, z, y, S),
				NewGrammarRule(C, y, B, C, C),
				NewGrammarRule(C),
				NewGrammarRule(C, w, z),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S, B, x),
				NewGrammarRule(B),
				NewGrammarRule(B, z, y, S, B),
				NewGrammarRule(B, z, B, z),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S, B, w, S, S),
				NewGrammarRule(S),
				NewGrammarRule(S, y),
				NewGrammarRule(B, x),
				NewGrammarRule(B),
			},
			err: ErrGrammar,
		},
		{
			rules: []GrammarRule{
				NewGrammarRule(S, T, EP),
				NewGrammarRule(EP, plus, T, EP),
				NewGrammarRule(EP),
				NewGrammarRule(T, F, TP),
				NewGrammarRule(TP, star, F, TP),
				NewGrammarRule(TP),
				NewGrammarRule(F, id),
				NewGrammarRule(F, lparen, S, rparen),
			},
			parseTable: map[int]map[int][]GrammarSym{
				S.Kind(): {
					id.Kind():     {T, EP},
					lparen.Kind(): {T, EP},
				},
				EP.Kind(): {
					plus.Kind():   {plus, T, EP},
					rparen.Kind(): nil,
					eof.Kind():    nil,
				},
				T.Kind(): {
					id.Kind():     {F, TP},
					lparen.Kind(): {F, TP},
				},
				TP.Kind(): {
					plus.Kind():   nil,
					star.Kind():   {star, F, TP},
					rparen.Kind(): nil,
					eof.Kind():    nil,
				},
				F.Kind(): {
					id.Kind():     {id},
					lparen.Kind(): {lparen, S, rparen},
				},
			},
		},
	} {
		parser, err := NewLL1Parser(c.rules, S, eof)
		if c.err != nil {
			assert.Errorf(err, "Should fail to create parser: case %d", n)
			assert.Truef(errors.Is(err, c.err), "Should fail to create parser: case %d", n)
			continue
		}
		assert.NoErrorf(err, "Failed to create parser: case %d, %v", n, err)
		assert.Equalf(c.parseTable, parser.table, "Fail case %d", n)
	}
}

func TestLL1Parser_Parse(t *testing.T) {
	assert := assert.New(t)

	g := NewGrammarSymGenerator()

	def := g.Term()
	eof := g.Term()
	wspace := g.Term()
	S := g.NonTerm()

	T := g.NonTerm()
	SP := g.NonTerm()
	TP := g.NonTerm()
	F := g.NonTerm()
	num := g.Term()
	plus := g.Term()
	star := g.Term()
	lparen := g.Term()
	rparen := g.Term()

	dfa := NewDfa(def.Kind())
	wspaceNode := NewDfa(wspace.Kind())
	dfa.AddDfa([]rune(" "), wspaceNode)
	wspaceNode.AddDfa([]rune(" "), wspaceNode)
	numNode := NewDfa(num.Kind())
	dfa.AddDfa([]rune("0123456789"), numNode)
	numNode.AddDfa([]rune("0123456789"), numNode)
	dfa.AddPath([]rune("+"), plus.Kind(), def.Kind())
	dfa.AddPath([]rune("*"), star.Kind(), def.Kind())
	dfa.AddPath([]rune("("), lparen.Kind(), def.Kind())
	dfa.AddPath([]rune(")"), rparen.Kind(), def.Kind())
	lexer := NewDfaLexer(dfa, def.Kind(), eof.Kind(), map[int]struct{}{
		wspace.Kind(): {},
	})

	parser, err := NewLL1Parser([]GrammarRule{
		NewGrammarRule(S, T, SP),
		NewGrammarRule(SP, plus, T, SP),
		NewGrammarRule(SP),
		NewGrammarRule(T, F, TP),
		NewGrammarRule(TP, star, F, TP),
		NewGrammarRule(TP),
		NewGrammarRule(F, num),
		NewGrammarRule(F, lparen, S, rparen),
	}, S, eof)
	assert.NoErrorf(err, "Failed to create parser: %v", err)

	var evalTree func(node ParseTree) (int, error)
	evalTree = func(node ParseTree) (int, error) {
		if node.Term() {
			if node.Kind() == num.Kind() {
				v, err := strconv.Atoi(node.Val())
				if err != nil {
					return 0, err
				}
				return v, nil
			}
			return 0, fmt.Errorf("Cannot eval")
		}
		children := node.Children()
		switch node.Kind() {
		case S.Kind():
			v1, err := evalTree(children[0])
			if err != nil {
				return 0, err
			}
			v2, err := evalTree(children[1])
			if err != nil {
				return 0, err
			}
			return v1 + v2, nil
		case SP.Kind():
			if len(children) == 0 {
				return 0, nil
			}
			v1, err := evalTree(children[1])
			if err != nil {
				return 0, err
			}
			v2, err := evalTree(children[2])
			if err != nil {
				return 0, err
			}
			return v1 + v2, nil
		case T.Kind():
			v1, err := evalTree(children[0])
			if err != nil {
				return 0, err
			}
			v2, err := evalTree(children[1])
			if err != nil {
				return 0, err
			}
			return v1 * v2, nil
		case TP.Kind():
			if len(children) == 0 {
				return 1, nil
			}
			v1, err := evalTree(children[1])
			if err != nil {
				return 0, err
			}
			v2, err := evalTree(children[2])
			if err != nil {
				return 0, err
			}
			return v1 * v2, nil
		case F.Kind():
			if len(children) == 1 {
				return evalTree(children[0])
			}
			return evalTree(children[1])
		default:
			return 0, fmt.Errorf("Cannot eval")
		}
	}

	for _, c := range []struct {
		text string
		err  error
		exp  int
	}{
		{
			text: "1 + 2 + 3",
			exp:  6,
		},
		{
			text: "3 * 2 + 3",
			exp:  9,
		},
		{
			text: "3 * (2 + 3)",
			exp:  15,
		},
	} {
		tokens, err := lexer.Tokenize([]rune(c.text))
		assert.NoErrorf(err, "Failed to tokenize %s: %v", c.text, err)
		tree, err := parser.Parse(tokens)
		if c.err != nil {
			assert.Errorf(err, "Should fail to parse %s", c.text)
			continue
		}
		assert.NoErrorf(err, "Failed to parse %s: %v", c.text, err)
		v, err := evalTree(*tree)
		assert.NoErrorf(err, "Failed to eval %s: %v", c.text, err)
		assert.Equal(c.exp, v)
	}
}
