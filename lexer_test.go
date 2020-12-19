package gnom

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMinInt(t *testing.T) {
	assert := assert.New(t)

	for _, c := range []struct {
		a int
		b int
	}{
		{
			a: 1,
			b: 2,
		},
	} {
		assert.Equalf(minInt(c.a, c.b), minInt(c.b, c.a), "Minimum is not the same: %d, %d", c.a, c.b)
		assert.LessOrEqualf(minInt(c.a, c.b), c.a, "Not minimum: %d, %d", c.a, c.b)
		assert.LessOrEqualf(minInt(c.a, c.b), c.b, "Not minimum: %d, %d", c.a, c.b)
	}
}

func TestToken(t *testing.T) {
	assert := assert.New(t)

	for _, c := range []struct {
		kind int
		val  string
	}{
		{
			kind: 1,
			val:  "abcd",
		},
	} {
		token := NewToken(c.kind, c.val)
		assert.Equalf(c.kind, token.Kind(), "Invalid kind: %d %s", c.kind, c.val)
		assert.Equalf(c.val, token.Val(), "Invalid val: %d %s", c.kind, c.val)
	}
}

func TestLexer_Tokenize(t *testing.T) {
	assert := assert.New(t)

	const (
		tokenDefault = iota
		tokenEOF
		tokenWSpace
		tokenNum
		tokenAdd
		tokenMul
		tokenValTypeInt
	)

	dfa := NewDfa(tokenEOF)
	wspace := NewDfa(tokenWSpace)
	dfa.AddDfa([]byte(" "), wspace)
	wspace.AddDfa([]byte(" "), wspace)
	num := NewDfa(tokenNum)
	dfa.AddDfa([]byte("0123456789"), num)
	num.AddDfa([]byte("0123456789"), num)
	dfa.AddPath([]byte("+"), tokenAdd, tokenDefault)
	dfa.AddPath([]byte("*"), tokenAdd, tokenDefault)
	dfa.AddPath([]byte("int"), tokenValTypeInt, tokenDefault)

	for _, c := range []struct {
		chars  string
		err    error
		tokens []Token
	}{
		{
			chars: "    314 +  1   int",
			tokens: []Token{
				NewToken(tokenNum, "314"),
				NewToken(tokenAdd, "+"),
				NewToken(tokenNum, "1"),
				NewToken(tokenValTypeInt, "int"),
				NewToken(tokenEOF, ""),
			},
		},
		{
			chars: "    314 -  1   ",
			err:   ErrLex,
		},
		{
			chars: "    314 in  ",
			err:   ErrLex,
		},
	} {
		lexer := NewLexer(dfa, tokenDefault, tokenEOF, map[int]struct{}{
			tokenWSpace: {},
		})
		tokens, err := lexer.Tokenize([]byte(c.chars))
		if c.err != nil {
			assert.Errorf(err, "Should fail to tokenize: %s", c.chars)
			assert.Truef(errors.Is(err, c.err), "Should fail to tokenize: %s", c.chars)
			continue
		}
		assert.NoErrorf(err, "Failed to tokenize: %s, %v", c.chars, err)
		assert.Equalf(c.tokens, tokens, "Failed to tokenize: %s", c.chars)
	}
}
