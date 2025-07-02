// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm_test

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	"github.com/szabba/assert/v3"
	"github.com/szabba/assert/v3/assertions/theerr"
	"github.com/szabba/assert/v3/assertions/theval"

	"github.com/szabba/ahm"
)

func TestErrorPlacedAt(t *testing.T) {

	t.Run("PanicsWithNonPositiveLineNumber", func(t *testing.T) {
		// given
		lineNo := int64(0)
		err := io.EOF

		// when
		caught := catchPanic(func() { ahm.ErrOnLine(lineNo, err) })

		// then
		assert.UsingFmt(t.Errorf).That(theval.Equal(caught, "line number 0 out of range"))
	})

	t.Run("PanicsWhenGivenANilError", func(t *testing.T) {
		// given
		lineNo := int64(9)
		var err error

		// when

		// when
		caught := catchPanic(func() { ahm.ErrOnLine(lineNo, err) })

		// then
		assert.UsingFmt(t.Errorf).That(theval.Equal(caught, "wrapping nil error at line 9"))
	})

	t.Run("CreatesAnErrorThatWrapsTheOriginal", func(t *testing.T) {
		// given
		lineNo := int64(11)
		err := io.EOF

		// when
		placedErr := ahm.ErrOnLine(lineNo, err)

		// then
		assert.UsingFmt(t.Errorf).That(theerr.Is(placedErr, err))
	})

	t.Run("DoubleWrappedErrorIsTheRoot", func(t *testing.T) {
		// given
		lineNo := int64(11)
		err := fmt.Errorf("context %d: %w", 10, io.EOF)

		// when
		placedErr := ahm.ErrOnLine(lineNo, err)

		// then
		assert.UsingFmt(t.Errorf).That(theerr.Is(placedErr, io.EOF))
	})

}

func FuzzParse(f *testing.F) {
	f.Add("")
	f.Add("@PROC")
	f.Add("Some text.")
	f.Add(strings.Join(
		[]string{
			"@PROC",
			"    Child text",
			"    @CHILD-PROC",
		},
		"\n"))

	f.Fuzz(func(t *testing.T, data string) {

		// We care about Parse not panicking.
		// We don't need a given/when/then structure to test that.
		t.Logf("%q", data)
		r := strings.NewReader(data)
		ahm.Parse(r)
	})
}

func TestParse(t *testing.T) {
	// given
	cases := map[string]ParseCase{
		"Empty": ParseCase{}.
			WithLines("").
			ExpectingNodes(ahm.Text("")),

		"OneLineOfText": ParseCase{}.
			WithLines("One line").
			ExpectingNodes(ahm.Text("One line")),

		"MultipleLinesOfText": ParseCase{}.
			WithLines(
				"One line.",
				"And another.").
			ExpectingNodes(
				ahm.Text("One line."),
				ahm.Text("And another.")),

		"EmptyNameProc": ParseCase{}.
			WithLines("@").
			ExpectingNodes().
			ExpectingErrorPlacedAt(1, ahm.ErrUnacceptableProcName("")),

		"NameOnlyProc": ParseCase{}.
			WithLines("@TOC").
			ExpectingNodes(ahm.Proc("TOC", "")),

		"OneLineProc": ParseCase{}.
			WithLines("@NAME Title").
			ExpectingNodes(ahm.Proc("NAME", "Title")),

		"IndentedFirstLine": ParseCase{}.
			WithLines(
				"    @DONE List tasks.",
				"@TODO Do the thing.").
			ExpectingNodes(
				ahm.Proc("DONE", "List tasks."),
				ahm.Proc("TODO", "Do the thing.")).
			ExpectingErrorPlacedAt(1, ahm.ErrMismatchedIndents()),

		"SuddenlyIndentedText": ParseCase{}.
			WithLines(
				"A line.",
				"    And another, unexpectedly indented.").
			ExpectingNodes(
				ahm.Text("A line."),
				ahm.Text("And another, unexpectedly indented.")).
			ExpectingErrorPlacedAt(2, ahm.ErrMismatchedIndents()),

		"ProcWithChild": ParseCase{}.
			WithLines(
				"@CODE bash",
				"    git status").
			ExpectingNodes(
				ahm.Proc("CODE", "bash",
					ahm.Text("git status"))),

		"ProcWithChildAndGrandChild": ParseCase{}.
			WithLines(
				"@PROC",
				"    @CHILD",
				"        @GRANDCHILD").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Proc("CHILD", "",
						ahm.Proc("GRANDCHILD", "")))),

		"ProcWithMisindentedChildText": ParseCase{}.
			WithLines(
				"@PROC",
				"    @CHILD-PROC",
				"  Misindented text.").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Proc("CHILD-PROC", ""),
					ahm.Text("Misindented text."))).
			ExpectingErrorPlacedAt(3, ahm.ErrMismatchedIndents()),

		"ProcWithMisindentedChildProc": ParseCase{}.
			WithLines(
				"@PROC",
				"    @CHILD-PROC",
				"  @MISINDENTED-CHILD-PROC").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Proc("CHILD-PROC", ""),
					ahm.Proc("MISINDENTED-CHILD-PROC", ""))).
			ExpectingErrorPlacedAt(3, ahm.ErrMismatchedIndents()),

		"TopLevelMisindentedProc": ParseCase{}.
			WithLines(
				"Some text.",
				"    @OVERINDENTED-PROC").
			ExpectingNodes(
				ahm.Text("Some text."),
				ahm.Proc("OVERINDENTED-PROC", "")).
			ExpectingErrorPlacedAt(2, ahm.ErrMismatchedIndents()),

		"ProcWithChildAfterEmmptyLine": ParseCase{}.
			WithLines(
				"@PROC",
				"        ",
				"    @CHILD").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Text(""),
					ahm.Proc("CHILD", ""))),

		"ProcWithChildrenAndSibling": ParseCase{}.
			WithLines(
				"@PROC",
				"    Child text",
				"Sibling text").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Text("Child text")),
				ahm.Text("Sibling text")),

		"ProcFollowedByMismactch": ParseCase{}.
			WithLines(
				"@PROC",
				"    Child text",
				"\tMismatch").
			ExpectingNodes(
				ahm.Proc("PROC", "",
					ahm.Text("Child text")),
				ahm.Text("Mismatch")),
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {

			in := strings.NewReader(tt.Input)

			// when
			nodes, err := ahm.Parse(in)

			// then
			assert := assert.UsingFmt(t.Errorf)

			assert.That(tt.ExpectedNodes(nodes))

			for _, errWanted := range tt.Errs {
				assert.That(theerr.Is(err, errWanted))
			}
		})
	}
}

type ParseCase struct {
	Input string
	Nodes []ahm.Node
	Errs  []error
}

func (c ParseCase) WithLines(lines ...string) ParseCase {
	c.Input = strings.Join(lines, "\n")
	return c
}

func (c ParseCase) ExpectingNodes(nodes ...ahm.Node) ParseCase {
	c.Nodes = nodes
	c.place(1, nodes)
	return c
}

func (c ParseCase) place(firstLine int64, nodes []ahm.Node) (nextLine int64) {

	nextLine = firstLine
	for i := range nodes {
		nodes[i] = nodes[i].PlacedOnLine(nextLine)
		nextLine++

		if nodes[i].Proc() {
			nextLine = c.place(nextLine, nodes[i].Children())
		}
	}
	return nextLine
}

func (c ParseCase) ExpectingErrorPlacedAt(lineNo int64, err error) ParseCase {
	c.Errs = append(slices.Clone(c.Errs), ahm.ErrOnLine(lineNo, err))
	return c
}

func (c ParseCase) ExpectedNodes(nodes []ahm.Node) error {
	want := ahm.FmtString(c.Nodes)
	got := ahm.FmtString(nodes)

	diff := udiff.Unified("want", "got", want, got)
	if diff != "" {
		return errors.New("\n" + diff)
	}
	return nil
}
