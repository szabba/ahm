// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

import (
	"fmt"
	"log"

	"github.com/szabba/ahm/assert"
)

type PositionIn struct {
	Source string
	Position
}

func (pos PositionIn) String() string {
	return fmt.Sprintf("%s:%d,%d", pos.Source, pos.line, pos.column)
}

type Position struct {
	line, column int
}

func PositionOf(line, column int) Position {
	assert.That(line >= 1, log.Panicf, "line = %d < 1", line)
	assert.That(column >= 1, log.Panicf, "column = %d < 1", column)
	return Position{line - 1, column - 1}
}

func First() Position { return Position{} }

func (pos Position) GetLine() int   { return pos.line + 1 }
func (pos Position) GetColumn() int { return pos.column + 1 }

func (pos Position) String() string {
	return fmt.Sprintf("%d,%d", pos.line, pos.column)
}

func (pos Position) In(source string) PositionIn {
	return PositionIn{source, pos}
}

func (pos Position) StartSpan() Span {
	return Span{start: pos, end: pos}
}

func (pos Position) NextAfter(r rune) Position {
	if r == '\n' {
		return pos.nextLine()
	}
	return pos.nextColumn()
}

func (pos Position) nextLine() Position {
	return Position{line: pos.line + 1}
}

func (pos Position) nextColumn() Position {
	return Position{line: pos.line, column: pos.column + 1}
}
