// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"log"

	"github.com/szabba/ahm/assert"
	"github.com/szabba/ahm/internal/lexer"
	"github.com/szabba/ahm/internal/token"
)

type tokenStream struct {
	lexer  *lexer.Lexer
	tokens []token.Token
	err    error
}

func newStream(lexer *lexer.Lexer) *tokenStream {
	stream := new(tokenStream)
	stream.lexer = lexer
	return stream
}

func (stream *tokenStream) accept(d int) {
	assert.That(d > 0, log.Panicf, "cannot accept %d tokens because %d <= 0", d, d)
	assert.That(
		d <= len(stream.tokens), log.Panicf,
		"cannot accept %d tokens when there are %d in the lookahead buffer",
		d, len(stream.tokens))

	rest := stream.tokens[d:]

	copy(stream.tokens, rest)
	stream.tokens = stream.tokens[:len(rest)]
}

func (stream *tokenStream) peek(d int) (token.Token, error) {
	stream.readForPeek(d)
	if d >= len(stream.tokens) {
		return token.Token{}, stream.err
	}
	return stream.tokens[d], nil
}

func (stream *tokenStream) readForPeek(d int) {
	if d < len(stream.tokens) {
		return
	}
	extraNeeded := len(stream.tokens) + 1 - d
	stream.readMore(extraNeeded)
}

func (stream *tokenStream) readMore(n int) {
	for i := 0; stream.err == nil && i < n; i++ {
		stream.readOne()
	}
}

func (stream *tokenStream) readOne() {
	if stream.err != nil {
		return
	}
	var tok token.Token
	tok, stream.err = stream.lexer.Next()
	if tok.TokenType == token.Invalid {
		return
	}
	stream.tokens = append(stream.tokens, tok)
}
