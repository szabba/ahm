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
	pos := Position{}.Next('a').Next('\n').Next('c')

	// when
	span := pos.StartSpan()

	// then
	assert.That(span.Start == pos, t.Errorf, "a new span should start at the position it was created from")
	assert.That(span.End == pos, t.Errorf, "a new span should end at the position it was created from")
}

func TestSpanAdd(t *testing.T) {
	// given
	pos := Position{}.Next('a').Next('\n').Next('c')
	initSpan := pos.StartSpan()
	nextPos := pos.Next('d').Next('\n').Next('\n').Next('g')

	// when
	span := initSpan.Add('d').Add('\n').Add('\n').Add('g')

	// then
	assert.That(
		span.Start == pos,
		t.Errorf,
		"after adds, a span should start where it initially did")

	assert.That(
		span.End == nextPos,
		t.Errorf,
		"after adds, a span should end where one would get to by traversing the rune sequence")
}

func TestSpanAddStartingAtZeroPos(t *testing.T) {
	// given
	zero := Position{}
	firstValid := zero.Next('a')
	nextPos := firstValid.Next('b').Next('\n').Next('c')

	// when
	span := zero.StartSpan().Add('a').Add('b').Add('\n').Add('c')

	// then
	assert.That(
		span.Start == firstValid,
		t.Errorf,
		"after adds, a span created from a zero position should start at the first valid one")

	assert.That(
		span.End == nextPos,
		t.Errorf,
		"after adds, a span created from a zero positon should end where one would get to by traversing the rune sequence")
}
