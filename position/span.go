// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

import "fmt"

type SpanIn struct {
	Source string
	Span
}

func (spanIn SpanIn) String() string {
	return fmt.Sprintf("%s:%s:%s", spanIn.Source, spanIn.Start, spanIn.End)
}

type Span struct {
	Start, End Position
}

func (span Span) String() string {
	return fmt.Sprintf("%s:%s", span.Start, span.End)
}

func (span Span) In(source string) SpanIn {
	return SpanIn{source, span}
}
