// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"testing"

	"github.com/szabba/ahm/assert"
)

func TestZeroPosition(t *testing.T) {
	// given
	var pos Position

	// when

	// then
	assert.That(pos.Line() == 0, t.Errorf, "the zero position must be on line 0")
	assert.That(pos.Column() == 0, t.Errorf, "the zero position must be in column 0")
	assert.That(!pos.IsValid(), t.Errorf, "the zero position must not be valid")
}

func TestFirstPositionAfterZero(t *testing.T) {
	// given
	var zero Position

	// when
	first := zero.Next('a')

	// then
	assert.That(first.Line() == 1, t.Errorf, "the first position must be on line 1")
	assert.That(first.Column() == 1, t.Errorf, "the first position must be in column 1")
	assert.That(first.IsValid(), t.Errorf, "the first position must be valid")
}

func TestNewlinePosition(t *testing.T) {
	// given
	var zero Position
	prev := zero.Next('a').Next('b').Next('c')

	// when
	newline := prev.Next('\n')

	// then
	assert.That(
		newline.Line() == prev.Line(),
		t.Errorf,
		"the newline position must be on the same line as the previous position")

	assert.That(
		newline.Column() == 1+prev.Column(),
		t.Errorf,
		"the newline position must be in the next column, compared to the previous position")

	assert.That(
		newline.IsValid(),
		t.Errorf,
		"the newline position must be valid")
}

func TestPositionAfterNewline(t *testing.T) {
	// given
	var zero Position
	eol := zero.Next('a').Next('b').Next('c').Next('\n')

	// when
	nextLine := eol.Next('d')

	// then
	assert.That(
		nextLine.Line() == 1+eol.Line(),
		t.Errorf,
		"the position after a newline must be on the next line, compared to the previous position")

	assert.That(
		nextLine.Column() == 1,
		t.Errorf,
		"the position after a newline must be in the first column")

	assert.That(
		nextLine.IsValid(),
		t.Errorf,
		"the position after a newline position must be valid")
}
