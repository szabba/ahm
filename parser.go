// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
)

const (
	ProcPrefix rune = '@'
	EOL             = '\n'
)

type Parser struct {
	src io.RuneScanner

	pos     Position
	indents []string

	rangeMap *RangeMap
}

func NewParser(r io.Reader, options ...ParserOption) *Parser {
	p := new(Parser)
	p.initInput(r)
	p.applyOptions(options...)
	p.fillInDefaults()
	return p
}

func (p *Parser) initInput(r io.Reader) {
	switch r := r.(type) {
	case io.RuneScanner:
		p.src = r
	default:
		p.src = bufio.NewReader(r)
	}
}

func (p *Parser) applyOptions(options ...ParserOption) {
	for _, opt := range options {
		opt.applyTo(p)
	}
}

func (p *Parser) fillInDefaults() {
	p.fillInRangeMap()
}

func (p *Parser) fillInRangeMap() {
	if p.rangeMap != nil {
		return
	}
	p.rangeMap = newRangeMap()
}

func (p *Parser) ParseAll() ([]Node, error) {
	return nil, nil
}

func (p *Parser) Parse() (Node, error) {
	node, err := p.parseNode()
	if err == errStartOfSibling {
		err = nil
	}
	return node, p.wrapError(err)
}

func (p *Parser) wrapError(err error) error {
	if err == nil {
		return nil
	}
	return &locatedError{err, p.pos}
}

// Tries to parse a node.
// Expects the current indent to be already processed for the first line.
//
// May return errStartOfSibling for valid input.
func (p *Parser) parseNode() (Node, error) {
	r, err := p.peekRune()
	if err != nil {
		return nil, err
	}
	switch r {
	case ProcPrefix:
		return p.parseProc()
	default:
		return p.parseText()
	}
}

// Tries to parse a proc node.
// Expects the current indent to be already processed for the first line.
//
// May return errStartOfSibling for valid input.
func (p *Parser) parseProc() (Node, error) {
	proc := new(Proc)

	err := p.skipRune(ProcPrefix)
	if err != nil {
		return nil, err
	}

	proc.Name, err = p.parseProcName()
	if err != nil {
		return proc, err
	}

	err = p.skipIntralineSpace()
	if err != nil {
		return proc, err
	}

	proc.Title, err = p.parseProcTitle()
	if err != nil {
		return proc, err
	}

	proc.Children, err = p.parseProcChildren()
	switch dedErr := err.(type) {
	case *dedentError:
		if dedErr.Dedent() == 1 {
			err = errStartOfSibling
		} else {
			err = &dedentError{dedErr.dedent - 1}
		}
	}
	return proc, err
}

func (p *Parser) parseProcName() (string, error) {
	buf := new(bytes.Buffer)
	err := p.acceptWhileNot(unicode.IsSpace, buf)
	return buf.String(), err
}

func (p *Parser) parseProcTitle() (string, error) {
	buf := new(bytes.Buffer)
	err := p.acceptLine(buf)
	content := buf.String()
	content = strings.TrimSuffix(content, "\n")
	return content, err
}

func (p *Parser) parseProcChildren() ([]Node, error) {
	children := []Node{}
	err := p.skipCurrentIndent()
	if err != nil {
		return nil, err
	}
	err = p.acceptExtraIndent()
	if err == errNoNewIndent {
		return nil, errStartOfSibling
	}
	if err != nil {
		return nil, err
	}

	for {
		var child Node
		child, err = p.parseNode()
		if p.isErrNotGroundsToSkipChild(err) {
			children = append(children, child)
		}
		if p.isErrGroundsNotToParseFurtherChildren(err) {
			return children, err
		}
	}
}

func (*Parser) isErrNotGroundsToSkipChild(err error) bool {
	isNil := err == nil
	isEOF := err == io.EOF
	isStartOfSibling := err == errStartOfSibling
	_, isDedent := err.(*dedentError)

	return isNil || isEOF || isStartOfSibling || isDedent
}

func (*Parser) isErrGroundsNotToParseFurtherChildren(err error) bool {
	isNotNil := err != nil
	isNotStartOfSibling := err != errStartOfSibling

	return isNotNil && isNotStartOfSibling
}

// Tries to parse a text node.
// Expects the current indent to be already processed for the first line.
//
// May return errStartOfSibling for valid input.
// When that happens, calling parseNode will parse a proc node.
func (p *Parser) parseText() (Node, error) {
	builder := new(textBuilder)

	for {
		err := p.parseTextLine(builder)
		if err == io.EOF || err == errStartOfSibling {
			return builder.Build(), err
		}
		if err != nil {
			return nil, err
		}

		err = p.skipCurrentIndent()
		if err != nil {
			return nil, err
		}
	}
}

// Tries to read a line of a text node.
// Expects the current indent to be already processed for the first line.
func (p *Parser) parseTextLine(builder *textBuilder) error {

	r, err := p.peekRune()
	if err != nil {
		return err
	}

	switch r {
	case ProcPrefix:
		return errStartOfSibling

	default:
		err = p.acceptLine(builder)
		if err == io.EOF {
			return io.EOF
		}
		if err != nil {
			return err
		}
		return nil
	}
}

func (p *Parser) skipCurrentIndent() error {
	return p.acceptCurrentIndent(ioutil.Discard)
}

func (p *Parser) acceptCurrentIndent(w io.Writer) error {
	for n := range p.indents {
		err := p.acceptIndentLevel(w, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) acceptIndentLevel(w io.Writer, n int) error {
	r, err := p.peekRune()
	if err != nil {
		return err
	}
	// TODO: extra handling for newlines
	if !unicode.IsSpace(r) || r == '\n' {
		dedent := p.dedentForLevel(n)
		p.dropToIndentLevel(n)
		return &dedentError{dedent}
	}
	return p.skipString(p.indents[n])
}

func (p *Parser) dedentForLevel(lvl int) int {
	return len(p.indents) - lvl + 1
}

func (p *Parser) acceptExtraIndent() error {
	extraIndent := new(bytes.Buffer)
	err := p.acceptIntralineSpace(extraIndent)
	if err != nil {
		return err
	}
	if extraIndent.Len() == 0 {
		return errNoNewIndent
	}
	p.indents = append(p.indents, extraIndent.String())
	return nil
}

func (p *Parser) dropToIndentLevel(n int) {
	p.indents = p.indents[:n]
}

func (p *Parser) skipIntralineSpace() error {
	return p.acceptIntralineSpace(ioutil.Discard)
}

func (p *Parser) acceptIntralineSpace(w io.Writer) error {
	isIntralineSpace := func(r rune) bool {
		return unicode.IsSpace(r) && r != EOL
	}
	return p.acceptWhile(isIntralineSpace, w)
}

func (p *Parser) acceptLine(w io.Writer) error {
	err := p.acceptUntilEOL(w)
	if err != nil {
		return err
	}
	err = writeRune(w, EOL)
	if err != nil {
		return err
	}
	return p.moveForward()
}

func (p *Parser) acceptUntilEOL(w io.Writer) error {
	notEOL := func(r rune) bool {
		return r != EOL
	}
	return p.acceptWhile(notEOL, w)
}

func (p *Parser) acceptWhileNot(pred func(r rune) bool, w io.Writer) error {
	notPred := func(r rune) bool {
		return !pred(r)
	}
	return p.acceptWhile(notPred, w)
}

func (p *Parser) acceptWhile(pred func(r rune) bool, w io.Writer) error {
	for {
		r, err := p.peekRune()
		if err != nil {
			return err
		}

		if !pred(r) {
			return nil
		}

		err = writeRune(w, r)
		if err != nil {
			return err
		}

		err = p.moveForward()
		if err != nil {
			return err
		}
	}
}

func (p *Parser) acceptStrings(w io.Writer, chunks ...string) error {
	for _, chunk := range chunks {
		err := p.acceptString(w, chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) skipString(s string) error {
	return p.acceptString(ioutil.Discard, s)
}

func (p *Parser) acceptString(w io.Writer, s string) error {
	wantedRunes := []rune(s)
	for _, want := range wantedRunes {
		err := p.acceptRune(w, want)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Parser) skipRune(r rune) error {
	return p.acceptRune(ioutil.Discard, r)
}

func (p *Parser) acceptRune(w io.Writer, want rune) error {
	got, err := p.peekRune()
	if err != nil {
		return err
	}
	if want != got {
		return &wrongRuneError{got: got, wanted: want}
	}
	err = writeRune(w, want)
	if err != nil {
		return err
	}
	return p.moveForward()
}

func (p *Parser) peekRune() (rune, error) {
	r, _, err := p.src.ReadRune()
	if err != nil {
		return r, err
	}
	err = p.src.UnreadRune()
	if err != nil {
		return r, err
	}
	return r, err
}

func (p *Parser) moveForward() error {
	r, _, err := p.src.ReadRune()
	if err == nil {
		p.pos = p.pos.Next(r)
	}
	return err
}
