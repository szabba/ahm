// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package position_test

import (
	"testing"

	"github.com/szabba/ahm/assert"
)

func TestZeroPosition(t *testing.T) {
	// given

	// when
	var first Position

	// then
	assert.That(first.Line() == 1, t.Errorf, "the first position must be on line 1")
	assert.That(first.Column() == 1, t.Errorf, "the first position must be in column 1")
}

func TestPositionAfterOneRune(t *testing.T) {
	// given
	var first Position

	// when
	afterOne := first.NextAfter('a')

	// then
	assert.That(afterOne.Line() == 1, t.Errorf, "the position must be on line 1")
	assert.That(afterOne.Column() == 2, t.Errorf, "the position must be in column 2")
}

func TestPositionAfterNewline(t *testing.T) {
	// given
	var first Position
	prev := first.NextAfter('a').NextAfter('b').NextAfter('c')

	// when
	startOfNewLine := prev.NextAfter('\n')

	// then
	assert.That(
		startOfNewLine.Line() == 1+prev.Line(),
		t.Errorf,
		"the position after a newline must be on the next line, compared to the previous position")

	assert.That(
		startOfNewLine.Column() == 1,
		t.Errorf,
		"the position after a newline must be in the first column")
}
