// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position

import (
	"testing"

	"github.com/szabba/ahm/assert"
)

func TestRangeFromPosition(t *testing.T) {
	// given
	pos := Position{}.Next('a').Next('\n').Next('c')

	// when
	range_ := pos.StartRange()

	// then
	assert.That(range_.Start() == pos, t.Errorf, "a new range should start at the position it was created from")
	assert.That(range_.End() == pos, t.Errorf, "a new range should end at the position it was created from")
}

func TestRangeAdd(t *testing.T) {
	// given
	pos := Position{}.Next('a').Next('\n').Next('c')
	initRange := pos.StartRange()
	nextPos := pos.Next('d').Next('\n').Next('\n').Next('g')

	// when
	range_ := initRange.Add('d').Add('\n').Add('\n').Add('g')

	// then
	assert.That(
		range_.Start() == pos,
		t.Errorf,
		"a range after adds should start at the same position as initially")

	assert.That(
		range_.End() == nextPos,
		t.Errorf,
		"a range after adds should have the same position one would get to by traversing the rune sequence")
}

// TODO: valid position creates valid range
// TODO: invalid position creates invalid range
