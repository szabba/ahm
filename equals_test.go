// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"testing"

	"github.com/kr/pretty"
	"github.com/szabba/ahm/assert"
)

func TestEquals(t *testing.T) {
	kases := map[string]struct {
		left, right Node
		equal       bool
	}{
		"twoNils":    {equal: true},
		"nilAndText": {left: &Text{"abcd"}},
		"nilAndProc": {left: &Proc{"name", "title", []Node{}}},
		"twoTexts": {
			left:  &Text{"abcd"},
			right: &Text{"abcd"},
			equal: true,
		},
		"twoFlatProcs": {
			left:  &Proc{"name", "title", nil},
			right: &Proc{"name", "title", nil},
			equal: true,
		},
		"twoProcsWithChildren": {
			left:  &Proc{"name", "title", []Node{&Text{"a child"}}},
			right: &Proc{"name", "title", []Node{&Text{"a child"}}},
			equal: true,
		},
		"procsWithDifferentNames": {
			left:  &Proc{"name-1", "title", nil},
			right: &Proc{"name-2", "title", nil},
		},
		"procsWithDifferentTitles": {
			left:  &Proc{"name", "title-1", nil},
			right: &Proc{"name", "title-2", nil},
		},
		"twoProcsWithDifferentChildren": {
			left:  &Proc{"name", "title", []Node{&Text{"first child"}}},
			right: &Proc{"name", "title", []Node{&Text{"second child"}}},
		},
	}

	for name, kase := range kases {
		t.Run(name, func(t *testing.T) {

			leftToRight := Equals(kase.left, kase.right)
			rightToLeft := Equals(kase.right, kase.left)

			assert.That(
				leftToRight == rightToLeft,
				t.Fatalf,
				"does not commute: %#v",
				pretty.Diff(leftToRight, rightToLeft))

			assert.That(
				Equals(kase.left, kase.right) == kase.equal,
				t.Fatalf,
				"result %v, wanted %v",
				leftToRight, kase.equal)
		})
	}
}
