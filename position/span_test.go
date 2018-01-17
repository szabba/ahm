// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position_test

import (
	"testing"

	"github.com/szabba/ahm/assert"
)

func TestSpanFromPosition(t *testing.T) {
	// given
	pos := Position{}.NextAfter('a').NextAfter('\n').NextAfter('c')

	// when
	span := pos.StartSpan()

	// then
	assert.That(span.StartsAt() == pos, t.Errorf, "a new span should start at the position it was created from")
	assert.That(span.EndsBefore() == pos, t.Errorf, "a new span should end at the position it was created from")
}

func TestSpanAdd(t *testing.T) {
	// given
	pos := Position{}.NextAfter('a').NextAfter('\n').NextAfter('c')
	initSpan := pos.StartSpan()
	nextPos := pos.NextAfter('d').NextAfter('\n').NextAfter('\n').NextAfter('g')

	// when
	span := initSpan.Add('d').Add('\n').Add('\n').Add('g')

	// then
	assert.That(
		span.StartsAt() == pos,
		t.Errorf,
		"after adds, a span should start where it initially did")

	t.Logf("nextPos = %s", nextPos)
	t.Logf("span.EndsBefore() = %s", span.EndsBefore())

	assert.That(
		span.EndsBefore() == nextPos,
		t.Errorf,
		"after adds, a span should end where one would get to by traversing the rune sequence")
}
