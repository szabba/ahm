// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

import "fmt"

type PositionIn struct {
	Source string
	Position
}

func (pos PositionIn) String() string {
	return fmt.Sprintf("%s:%d,%d", pos.Source, pos.Line, pos.Column)
}

type Position struct {
	Line, Column int
}

func First() Position { return Position{1, 1} }

func (pos Position) String() string {
	return fmt.Sprintf("%d,%d", pos.Line, pos.Column)
}

func (pos Position) In(source string) PositionIn {
	return PositionIn{source, pos}
}

func (pos Position) IsValid() bool { return pos.Line >= 1 && pos.Column >= 1 }

func (pos Position) NextLine() Position {
	return Position{pos.Line + 1, 1}
}

func (pos Position) NextColumn() Position {
	return Position{pos.Line, pos.Column + 1}
}
