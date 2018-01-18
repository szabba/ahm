// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lexer

import (
	"bytes"

	"github.com/szabba/ahm/internal/token"
	"github.com/szabba/ahm/position"
)

type tokenBuilder struct {
	typ                token.TokenType
	buf                bytes.Buffer
	span               position.Span
	lastRuneWasNewline bool
}

func (builder *tokenBuilder) startToken(typ token.TokenType) {
	builder.typ = typ
	builder.buf.Reset()
	builder.span = builder.span.EndsBefore().StartSpan()
}

func (builder *tokenBuilder) startSkipping() {
	builder.startToken(token.Invalid)
}

func (builder *tokenBuilder) isNotBuilding() bool { return builder.typ == token.Invalid }

func (builder *tokenBuilder) isNotEmpty() bool { return !builder.isEmpty() }

func (builder *tokenBuilder) isEmpty() bool {
	return builder.span.StartsAt() == builder.span.EndsBefore()
}

func (builder *tokenBuilder) build() token.Token {
	tok := token.Token{TokenType: builder.typ, Span: builder.span, Text: builder.buf.String()}
	builder.startSkipping()
	return tok
}

func (builder *tokenBuilder) acceptRune(r rune) {
	builder.advancePosition(r)

	if builder.isNotBuilding() {
		return
	}

	builder.buf.WriteRune(r)
}

func (builder *tokenBuilder) advancePosition(r rune) {
	builder.span = builder.nextSpan(r)
}

func (builder *tokenBuilder) nextSpan(r rune) position.Span {
	span := builder.span.Add(r)
	if builder.isNotBuilding() {
		span = span.EndsBefore().StartSpan()
	}
	return span
}
