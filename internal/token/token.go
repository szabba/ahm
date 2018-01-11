// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package token

import (
	"fmt"

	"github.com/szabba/ahm/position"
)

//go:generate stringer -type TokenType

type TokenType int

const (
	Invalid TokenType = iota
	Indent
	Dedent
	Misdent
	Newline
	ProcMark
	ProcName
	ProcArg
	Text
)

type Token struct {
	TokenType TokenType
	Span      position.Span
	Text      string
}

func (tok Token) String() string {
	return fmt.Sprintf("%s:%s:%q", tok.Span, tok.TokenType, tok.Text)
}
