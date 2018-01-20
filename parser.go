// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"bufio"
	"bytes"
	"io"
	"log"

	"github.com/pkg/errors"
	"github.com/szabba/ahm/assert"
	"github.com/szabba/ahm/internal/lexer"
	"github.com/szabba/ahm/internal/token"
)

type Parser struct {
	tokens tokenStream
}

func NewParser(r io.Reader) *Parser {
	p := new(Parser)
	p.tokens = *newStream(lexer.New(p.lexerSource(r)))
	return p
}

func (*Parser) lexerSource(r io.Reader) io.RuneScanner {
	switch r := r.(type) {
	case io.RuneScanner:
		return r
	default:
		return bufio.NewReader(r)
	}
}

func (p *Parser) ParseAll() ([]Node, error) {
	var nodes []Node
	for {
		node, err := p.Parse()

		if err == io.EOF {
			return nodes, nil
		} else if err != nil {
			return nodes, err
		}

		nodes = append(nodes, node)
	}
}

func (p *Parser) Parse() (Node, error) {
	tok, err := p.tokens.peek(0)
	if err != nil {
		return nil, err
	}

	switch tok.TokenType {

	case token.ProcMark:
		return p.parseProc()

	case token.Text:
		return p.parseText()

	default:
		return nil, p.unexpectedToken(tok, token.ProcMark, token.Text)
	}
}

func (p *Parser) parseProc() (Node, error) {
	first, err := p.tokens.peek(0)

	assert.That(err == nil, log.Panicf, "cannot parse proc: %s", err)
	assert.That(first.TokenType == token.ProcMark, log.Panicf, "cannot parse text: %s", p.unexpectedToken(first, token.ProcMark))

	p.tokens.accept(1)
	proc := new(Proc)

	tok, err := p.tokens.peek(0)
	if err != nil {
		return nil, err
	} else if tok.TokenType != token.ProcName {
		return nil, p.unexpectedToken(tok, token.ProcName)
	}
	proc.Name = tok.Text

	tok, err = p.tokens.peek(1)
	if err != nil {
		return nil, err
	} else if tok.TokenType != token.ProcArg {
		return nil, p.unexpectedToken(tok, token.ProcArg)
	}
	proc.Title = tok.Text

	p.tokens.accept(2)

	proc.Children, err = p.parseNestedNodes()
	if err == nil {
		_, err = p.tokens.peek(0)
	}
	return proc, err
}

func (p *Parser) parseNestedNodes() ([]Node, error) {
	tok, err := p.tokens.peek(0)
	if err != nil {
		return nil, nil
	} else if tok.TokenType == token.Dedent {
		return nil, nil
	} else if tok.TokenType != token.Newline {
		return nil, p.unexpectedToken(tok, token.Newline, token.Dedent)
	}
	p.tokens.accept(1)

	tok, err = p.tokens.peek(0)
	if err != nil {
		return nil, nil
	} else if tok.TokenType != token.Indent {
		return nil, nil
	}

	p.tokens.accept(1)

	nodes, err := p.parseNodesUntilDedent()
	if err != nil {
		return nil, err
	}

	tok, err = p.tokens.peek(0)
	if err != nil {
		return nil, nil
	} else if tok.TokenType != token.Dedent {
		return nil, p.unexpectedToken(tok, token.Dedent)
	}
	p.tokens.accept(1)

	return nodes, nil
}

func (p *Parser) parseNodesUntilDedent() ([]Node, error) {
	var nodes []Node
	for {
		node, err := p.Parse()
		nodes = append(nodes, node)
		if err != nil {
			return nodes, err
		}
		tok, err := p.tokens.peek(0)
		if err != nil {
			return nodes, err
		} else if tok.TokenType == token.Dedent {
			return nodes, nil
		} else if tok.TokenType == token.Newline {
			p.tokens.accept(1)
		}
	}
}

func (p *Parser) parseText() (Node, error) {
	first, err := p.tokens.peek(0)

	assert.That(err == nil, log.Panicf, "cannot parse text: %s", err)
	assert.That(first.TokenType == token.Text, log.Panicf, "cannot parse text: %s", p.unexpectedToken(first, token.Text))

	buf := bytes.NewBufferString(first.Text)
	makeNode := func() Node { return &Text{buf.String()} }
	p.tokens.accept(1)

	for {
		tok, err := p.tokens.peek(0)
		if err != nil {
			return makeNode(), err
		} else if tok.TokenType != token.Newline {
			return makeNode(), nil
		}

		newl := tok

		tok, err = p.tokens.peek(1)
		if err != nil {
			return nil, err
		} else if tok.TokenType == token.Dedent {
			p.tokens.accept(1)
			return makeNode(), err
		} else if tok.TokenType != token.Text {
			return makeNode(), err
		}

		io.WriteString(buf, newl.Text)
		io.WriteString(buf, tok.Text)
		p.tokens.accept(2)
	}
}

func (*Parser) unexpectedToken(tok token.Token, typs ...token.TokenType) error {
	return errors.Errorf("got %s, wanted one of the types %s", tok, typs)
}
