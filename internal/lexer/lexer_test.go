// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lexer_test

import (
	"io"
	"strings"
	"testing"

	"github.com/szabba/ahm/assert"
	"github.com/szabba/ahm/internal/lexer"
	"github.com/szabba/ahm/internal/token"
	"github.com/szabba/ahm/position"
)

func TestEmptyInput(t *testing.T) {
	expectTokens(t, "", token.Token{token.Text, span(1, 1, 1, 1), ""})
}

func TestLineOfText(t *testing.T) {
	expectTokens(t, "A line of text", token.Token{token.Text, span(1, 1, 1, 15), "A line of text"})
}

func TestMultipleLinesOfText(t *testing.T) {
	expectTokens(
		t, multiline(
			"First line",
			"Second line",
		),

		token.Token{token.Text, span(1, 1, 1, 11), "First line"},
		token.Token{token.Newline, span(1, 11, 2, 1), "\n"},
		token.Token{token.Text, span(2, 1, 2, 12), "Second line"},
	)
}

func TestSingleLineProc(t *testing.T) {
	expectTokens(
		t, "@proc arg",

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 6), "proc"},
		token.Token{token.ProcArg, span(1, 7, 1, 10), "arg"},
	)
}

func TestSingleLineProcWithoutArg(t *testing.T) {
	expectTokens(
		t, "@proc   ",

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 6), "proc"},
		token.Token{token.ProcArg, span(1, 9, 1, 9), ""},
	)
}

func TestProcWithTextChild(t *testing.T) {
	expectTokens(
		t, multiline(
			"@proc",
			"  Child text",
		),

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 6), "proc"},
		token.Token{token.ProcArg, span(1, 6, 1, 6), ""},
		token.Token{token.Newline, span(1, 6, 2, 1), "\n"},
		token.Token{token.Indent, span(2, 1, 2, 3), "  "},
		token.Token{token.Text, span(2, 3, 2, 13), "Child text"},
		token.Token{token.Dedent, span(2, 13, 2, 13), ""},
	)
}

func TestProcWithProcChild(t *testing.T) {
	expectTokens(
		t, multiline(
			"@parent",
			"  @child",
		),

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 8), "parent"},
		token.Token{token.ProcArg, span(1, 8, 1, 8), ""},
		token.Token{token.Newline, span(1, 8, 2, 1), "\n"},
		token.Token{token.Indent, span(2, 1, 2, 3), "  "},
		token.Token{token.ProcMark, span(2, 3, 2, 4), "@"},
		token.Token{token.ProcName, span(2, 4, 2, 9), "child"},
		token.Token{token.ProcArg, span(2, 9, 2, 9), ""},
		token.Token{token.Dedent, span(2, 9, 2, 9), ""},
	)
}

func TestProcWithTextGrandchild(t *testing.T) {
	expectTokens(
		t, multiline(
			"@grandparent",
			"  @parent",
			"      child",
		),

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 13), "grandparent"},
		token.Token{token.ProcArg, span(1, 13, 1, 13), ""},
		token.Token{token.Newline, span(1, 13, 2, 1), "\n"},
		token.Token{token.Indent, span(2, 1, 2, 3), "  "},
		token.Token{token.ProcMark, span(2, 3, 2, 4), "@"},
		token.Token{token.ProcName, span(2, 4, 2, 10), "parent"},
		token.Token{token.ProcArg, span(2, 10, 2, 10), ""},
		token.Token{token.Newline, span(2, 10, 3, 1), "\n"},
		token.Token{token.Indent, span(3, 3, 3, 7), "    "},
		token.Token{token.Text, span(3, 7, 3, 12), "child"},
		token.Token{token.Dedent, span(3, 12, 3, 12), ""},
		token.Token{token.Dedent, span(3, 12, 3, 12), ""},
	)
}

func TestProcWithChildThenTextSibling(t *testing.T) {
	expectTokens(
		t, multiline(
			"@parent",
			"  child",
			"aunt",
		),

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 8), "parent"},
		token.Token{token.ProcArg, span(1, 8, 1, 8), ""},
		token.Token{token.Newline, span(1, 8, 2, 1), "\n"},
		token.Token{token.Indent, span(2, 1, 2, 3), "  "},
		token.Token{token.Text, span(2, 3, 2, 8), "child"},
		token.Token{token.Newline, span(2, 8, 3, 1), "\n"},
		token.Token{token.Dedent, span(3, 1, 3, 1), ""},
		token.Token{token.Text, span(3, 1, 3, 5), "aunt"},
	)
}
func TestProcWithTextGrandchildThenTextSibling(t *testing.T) {
	expectTokens(
		t, multiline(
			"@grandparent",
			"  @parent",
			"      child",
			"grandaunt",
		),

		token.Token{token.ProcMark, span(1, 1, 1, 2), "@"},
		token.Token{token.ProcName, span(1, 2, 1, 13), "grandparent"},
		token.Token{token.ProcArg, span(1, 13, 1, 13), ""},
		token.Token{token.Newline, span(1, 13, 2, 1), "\n"},
		token.Token{token.Indent, span(2, 1, 2, 3), "  "},
		token.Token{token.ProcMark, span(2, 3, 2, 4), "@"},
		token.Token{token.ProcName, span(2, 4, 2, 10), "parent"},
		token.Token{token.ProcArg, span(2, 10, 2, 10), ""},
		token.Token{token.Newline, span(2, 10, 3, 1), "\n"},
		token.Token{token.Indent, span(3, 3, 3, 7), "    "},
		token.Token{token.Text, span(3, 7, 3, 12), "child"},
		token.Token{token.Newline, span(3, 12, 4, 1), "\n"},
		token.Token{token.Dedent, span(4, 1, 4, 1), ""},
		token.Token{token.Dedent, span(4, 1, 4, 1), ""},
		token.Token{token.Text, span(4, 1, 4, 10), "grandaunt"},
	)
}

func multiline(lines ...string) string { return strings.Join(lines, "\n") }

func expectTokens(t *testing.T, rawInput string, tokens ...token.Token) {

	lexer := lexer.New(strings.NewReader(rawInput))

	var (
		tok token.Token
		err error
	)

	for i, want := range tokens {
		tok, err = lexer.Next()
		t.Logf("token %d: token = %s", i, tok)
		assert.That(err == nil, t.Logf, "token %d: error: %s", i, err)

		assert.That(tok == want, t.Fatalf, "token %d: got %s, wanted %s", i, tok, want)

		isLast := i+1 == len(tokens)

		if isLast {
			assert.That(err == io.EOF, t.Fatalf, "token %d: got error %q, wanted %q", i, err, io.EOF)
		} else {
			assert.That(err == nil, t.Fatalf, "token %d: unexpected error: %q", i, err)
		}
	}

	want := token.Token{TokenType: token.Invalid}
	tok, err = lexer.Next()
	assert.That(tok == want, t.Fatalf, "after expected output: got token %s, wanted %s", tok, want)
	assert.That(err == io.EOF, t.Fatalf, "got error %q, wanted %q", err, io.EOF)
}

func span(startLine, startCol, endLine, endCol int) position.Span {
	return position.SpanFromTo(
		position.PositionOf(startLine, startCol),
		position.PositionOf(endLine, endCol))
}
