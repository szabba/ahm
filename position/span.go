// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

import (
	"fmt"
	"log"

	"github.com/szabba/ahm/assert"
)

type SpanIn struct {
	Source string
	Span
}

func (spanIn SpanIn) String() string {
	return fmt.Sprintf("%s:%s:%s", spanIn.Source, spanIn.start, spanIn.EndsBefore())
}

type Span struct {
	start, end Position
}

func SpanFromTo(from, to Position) Span {
	assert.That(
		from.GetLine() <= to.GetLine(),
		log.Panicf,
		"start position (%s) must not be on a line further than the end position (%s)",
		from, to)

	if from.GetLine() == to.GetLine() {
		assert.That(
			from.GetColumn() <= to.GetColumn(),
			log.Panicf,
			"start position (%s) must not be on a column further than the end position (%s) that is on the same line",
			from, to)
	}

	return Span{from, to}
}

func (span Span) StartsAt() Position { return span.start }

func (span Span) EndsBefore() Position { return span.end }

func (span Span) Add(r rune) Span {
	return Span{start: span.start, end: span.end.NextAfter(r)}
}

func (span Span) String() string {
	return fmt.Sprintf("%s:%s", span.start, span.EndsBefore())
}

func (span Span) In(source string) SpanIn {
	return SpanIn{source, span}
}
