// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

type Range struct {
	start, end Position
}

func (range_ Range) Start() Position { return range_.start }
func (range_ Range) End() Position   { return range_.end }

func (range_ Range) Add(r rune) Range {
	next := range_
	next.end = range_.end.Next(r)
	return next
}
