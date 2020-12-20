package gnom

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
			syms: []GrammarSym{NewGrammarNonTerm(3), NewGrammarNonTerm(1), NewGrammarNonTerm(4)},
			exp:  true,
		},
	} {
		assert.Equalf(c.exp, isLL1Nullable(c.syms, nullableSet), "Fail case %d", n)
	}
}
