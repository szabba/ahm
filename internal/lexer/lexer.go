// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lexer

import (
	"bufio"
	"io"
	"unicode"

	"github.com/szabba/ahm/ahmerr"
	"github.com/szabba/ahm/internal/token"
)

const (
	NewlineRune  = '\n'
	ProcMarkRune = '@'
)

type Lexer struct {
	src       io.RuneScanner
	nextToken tokenBuilder
	indents   indentStack
	next      func() error
	err       error
}

func NewFromReader(input io.Reader) *Lexer {
	return New(bufio.NewReader(input))
}

func New(input io.RuneScanner) *Lexer {
	lex := new(Lexer)
	lex.src = input
	lex.next = lex.scanLineAfterIndent
	return lex
}

func (lex *Lexer) Next() (token.Token, error) {
	if lex.err != nil {
		lex.next = nil
		return token.Token{}, lex.err
	}
	lex.err = lex.next()
	tok := lex.nextToken.build()
	lex.adjustCurrentIndent(tok)
	return tok, lex.err
}

func (lex *Lexer) adjustCurrentIndent(tok token.Token) {
	switch tok.TokenType {
	case token.Indent:
		lex.indents.pushIndent(tok.Text)
	case token.Dedent:
		lex.indents.popIndent()
	}
}

func (lex *Lexer) tryToScanIndent() error {
	lex.next = lex.scanLineAfterIndent

	indents := lex.indents.indents
	levelsMatched, err := lex.acceptStrings(indents...)
	if err != nil && err != io.EOF {
		return err
	}

	if levelsMatched < len(indents) {
		return lex.produceDedents(len(indents) - levelsMatched)()
	}

	lex.startToken(token.Indent)
	err = lex.acceptWhile(isIntralineSpace)
	if err != nil && err != io.EOF {
		return err
	}

	if lex.nextToken.isNotEmpty() {
		return nil
	}

	return lex.scanLineAfterIndent()
}

func (lex *Lexer) produceDedents(n int) func() error {
	return func() error {
		lex.next = lex.nextWhenDedentsLeft(n)
		lex.startToken(token.Dedent)
		return nil
	}
}

func (lex *Lexer) nextWhenDedentsLeft(n int) func() error {
	if n > 1 {
		return lex.produceDedents(n - 1)
	}
	return lex.scanLineAfterIndent
}

func (lex *Lexer) scanLineAfterIndent() error {
	r, err := lex.peekRune()
	if err != nil && err != io.EOF {
		return err
	}
	if r == ProcMarkRune {
		return lex.scanProcMark()
	}
	return lex.scanText()
}

func (lex *Lexer) scanText() error {
	lex.next = lex.scanNewline
	lex.startToken(token.Text)
	return lex.acceptWhile(isNotNewline)
}

func (lex *Lexer) scanProcMark() error {
	lex.next = lex.scanProcName
	lex.startToken(token.ProcMark)
	return lex.acceptOne(ProcMarkRune)
}

func (lex *Lexer) scanProcName() error {
	lex.next = lex.scanProcRest
	lex.startToken(token.ProcName)
	err := lex.acceptWhile(isNotSpace)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (lex *Lexer) scanProcRest() error {
	err := lex.skipWhile(isIntralineSpace)
	if err != nil && err != io.EOF {
		return err
	}
	lex.next = lex.scanNewline
	lex.startToken(token.ProcArg)
	return lex.acceptWhile(isNotNewline)
}

func (lex *Lexer) scanNewline() error {
	lex.next = lex.tryToScanIndent
	lex.startToken(token.Newline)
	return lex.acceptOne(NewlineRune)
}

func (lex *Lexer) skipWhile(p func(rune) bool) error {
	lex.nextToken.startSkipping()
	err := lex.acceptWhile(p)
	if err != nil {
		return err
	}
	return nil
}

func (lex *Lexer) startToken(typ token.TokenType) { lex.nextToken.startToken(typ) }

func (lex *Lexer) acceptWhile(p func(rune) bool) error {
	for {
		r, _, err := lex.src.ReadRune()

		if err != nil {
			return err
		}
		if !p(r) {
			return lex.src.UnreadRune()
		}

		lex.nextToken.acceptRune(r)
	}
}

// acceptStrings tries to consume excactly the runes in each element of strs, one after another.
//
// n is the number of strs elements that have been accepted succesfully.
// err is an error, if any.
func (lex *Lexer) acceptStrings(strs ...string) (n int, err error) {
	// FIXME: translate certain conditions into io.UnexpectedEOF ?
	for level, s := range strs {
		for i, r := range s {

			err := lex.acceptOne(r)
			if i == 0 && ahmerr.IsUnexpectedRune(err) {
				err = nil
				err = lex.src.UnreadRune()
				return level, err
			}
			if err != nil {
				return level, err
			}
		}
	}
	return len(strs), nil
}

func (lex *Lexer) acceptOne(want rune) error {
	r, _, err := lex.src.ReadRune()
	if err != nil {
		return err
	}
	if r != want {
		return ahmerr.NewUnexpectedRuneError(r, want)
	}
	lex.nextToken.acceptRune(r)
	return nil
}

func (lex *Lexer) peekRune() (rune, error) {
	r, _, err := lex.src.ReadRune()
	if err != nil {
		return r, err
	}
	err = lex.src.UnreadRune()
	return r, err
}

func isNotSpace(r rune) bool { return !unicode.IsSpace(r) }

func isIntralineSpace(r rune) bool { return r != NewlineRune && unicode.IsSpace(r) }

func isNotNewline(r rune) bool { return r != NewlineRune }
