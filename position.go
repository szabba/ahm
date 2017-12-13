// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

type Position struct {
	line, column int
	isNewline    bool
	srcName      string
}

func InSource(srcName string) Position {
	return Position{srcName: srcName}
}

func (pos Position) SourceName() string { return pos.srcName }

func (pos Position) Line() int   { return pos.line }
func (pos Position) Column() int { return pos.column }

func (pos Position) IsValid() bool { return pos.line >= 1 && pos.column >= 1 }

func (pos Position) Next(r rune) Position {
	next := pos
	if pos.line == 0 {
		next.line++
	}
	next.column++
	if pos.isNewline {
		next.line++
		next.column = 1
	}
	next.isNewline = r == '\n'
	return next
}

func (pos Position) StartRange() Range {
	return Range{pos, pos}
}
