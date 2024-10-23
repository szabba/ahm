// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm_test

import (
	"reflect"
	"testing"

	"github.com/szabba/assert/v3"
	"github.com/szabba/assert/v3/assertions/theslice"
	"github.com/szabba/assert/v3/assertions/theval"

	"github.com/szabba/ahm"
)

func TestTextIsText(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	is := n.Text()

	// then
	assert.UsingFmt(t.Errorf).
		True(is, "%s claims it is not text", n)
}

func TestTextIsNotProc(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	is := n.Proc()

	// then
	assert.UsingFmt(t.Errorf).
		True(!is, "%s claims it is a proc", n)
}

func TestProcIsNotText(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	is := n.Text()

	// then
	assert.UsingFmt(t.Errorf).
		True(!is, "%s claims it is text", n)
}

func TestProcIsProc(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	is := n.Proc()

	// then
	assert.UsingFmt(t.Errorf).
		True(is, "%s claims it is not a proc", n)
}

func TestEmptyTextIsZeroValue(t *testing.T) {
	// given
	text := ""

	// when
	n := ahm.Text(text)

	// then
	assert.UsingFmt(t.Errorf).That(theval.Zero(n))
}

func TestNewNodeHasZeroLineNo(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	no := n.LineNo()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Zero(no))
}

func TestPlacedNodeHasLineNoSet(t *testing.T) {
	// given
	n := ahm.Text("").PlacedAt(12)

	// when
	no := n.LineNo()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(no, 12))
}

func TestNodeCannotBePlacedAtNegativePosition(t *testing.T) {
	// given
	unplaced := ahm.Text("")

	// when
	caught := catchPanic(func() { unplaced.PlacedAt(-7) })

	// then
	assert.UsingFmt(t.Errorf).That(theval.NotZero(caught))
}

func TestRawTextIsNodeText(t *testing.T) {
	// given
	text := "Sample"
	n := ahm.Text(text)

	// when
	nt := n.NodeText()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(nt, text))
}

func TestProcPanicsOnNodeText(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	caught := catchPanic(func() { n.NodeText() })

	// then
	assert.UsingFmt(t.Errorf).That(theval.NotZero(caught))
}

func TestRawNameIsProcName(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	name := n.Name()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(name, "H1"))
}

func TestTextPanicsOnName(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	caught := catchPanic(func() { n.Name() })

	// then
	assert.UsingFmt(t.Errorf).That(theval.NotZero(caught))
}

func TestRawTitleIsProcTitle(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	title := n.Title()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(title, "A header"))
}

func TestTextPanicsOnTitle(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	caught := catchPanic(func() { n.Title() })

	// then
	assert.UsingFmt(t.Errorf).That(theval.NotZero(caught))
}

func TestRawChildrenAreProcChildren(t *testing.T) {
	// given
	wanted := []ahm.Node{
		ahm.Text("A sentence."),
		ahm.Text("And another."),
	}
	n := ahm.Proc("H1", "A header", wanted...)

	// when
	children := n.Children()

	// then
	assert.UsingFmt(t.Errorf).
		That(theslice.EqualFunc(
			children, wanted,
			func(l, r ahm.Node) bool { return reflect.DeepEqual(l, r) }))
}

func TestTextPanicsOnChildren(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	caught := catchPanic(func() { n.Children() })

	// then
	assert.UsingFmt(t.Errorf).That(theval.NotZero(caught))
}

func TestTextString(t *testing.T) {
	// given
	n := ahm.Text("Sample")

	// when
	s := n.String()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(s, `Text("Sample")`))
}

func TestMinProcString(t *testing.T) {
	// given
	n := ahm.Proc("", "")

	// when
	s := n.String()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(s, `Proc("", "")`))
}

func TestChildlessProcString(t *testing.T) {
	// given
	n := ahm.Proc("H1", "A header")

	// when
	s := n.String()

	// then
	assert.UsingFmt(t.Errorf).That(theval.Equal(s, `Proc("H1", "A header")`))
}

func TestOneChildProc(t *testing.T) {
	// given
	n := ahm.Proc("P", "", ahm.Text("A sentence, almost."))

	// when
	s := n.String()

	// then
	assert.UsingFmt(t.Errorf).
		That(theval.Equal(s, `Proc("P", "", Text("A sentence, almost."))`))
}

func TestParentProc(t *testing.T) {
	// given
	n := ahm.Proc("P", "", ahm.Text("A"), ahm.Text("B"))

	// when
	s := n.String()

	// then
	assert.UsingFmt(t.Errorf).
		That(theval.Equal(s, `Proc("P", "", Text("A"), Text("B"))`))
}

func catchPanic(f func()) (caught any) {
	defer func() { caught = recover() }()
	f()
	return nil
}
