// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package lexer

import (
	"log"

	"github.com/szabba/ahm/assert"
)

type indentStack struct {
	indents []string
}

func (stack *indentStack) pushIndent(extraIndent string) {
	stack.indents = append(stack.indents, extraIndent)
}

func (stack *indentStack) popIndent() {
	assert.That(len(stack.indents) >= 1, log.Panicf, "stack underflow")
	stack.indents = stack.indents[:len(stack.indents)-1]
}
