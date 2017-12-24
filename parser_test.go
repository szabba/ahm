// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"github.com/pkg/errors"
	"github.com/szabba/ahm/assert"
)

func TestSingleLineTextIsRead(t *testing.T) {
	// given
	wantNode := &Text{"A line."}

	rawInput := "A line."
	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, expected %q", err, io.EOF)
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	reportDiffs(t.Errorf, node, wantNode)
}

func TestMultipleLinesOfTextAreRead(t *testing.T) {
	// given
	wantNode := &Text{"Multiple\nlines."}
	rawInput := multiline(
		"Multiple",
		"lines.")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, expected %q", err, io.EOF)
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	reportDiffs(t.Errorf, node, wantNode)
}

func TestTextReadingStopsAtALinePrefixedWithAnAtSign(t *testing.T) {
	// given
	wantNode := &Text{"Some lines\nof text."}

	rawInput := multiline(
		"Some lines",
		"of text.",
		"@A-PROC")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(err == nil, t.Errorf, "got error %q, expected none", err)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestNameOnlyProcIsRead(t *testing.T) {
	// given
	wantNode := &Proc{Name: "A-PROC"}

	rawInput := "@A-PROC"

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestSpacesAreNotIncludedInReadProcName(t *testing.T) {
	// given
	wantNode := &Proc{Name: "A-PROC"}

	rawInput := "@A-PROC "

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestProcWithNameAndTitleIsRead(t *testing.T) {
	// given
	wantNode := &Proc{Name: "A-PROC", Title: "TITLE"}

	rawInput := "@A-PROC TITLE"

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestProcReadingStopsAtNextNonindentedLine(t *testing.T) {
	// given
	wantNode := &Proc{Name: "A-PROC", Title: "TITLE"}
	rawInput := multiline(
		"@A-PROC TITLE",
		"Some text.")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(err == nil, t.Errorf, "got error %q, wanted none", err)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestProcReadIncludesIndentedTextAsChild(t *testing.T) {
	// given
	wantNode := &Proc{
		Name:  "A-PROC",
		Title: "TITLE",
		Children: []Node{
			&Text{"Some text."},
		},
	}

	rawInput := multiline(
		"@A-PROC TITLE",
		"  Some text.")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestIndentedTextChildCanSpanMultipleLines(t *testing.T) {
	// given
	wantNode := &Proc{
		Name:  "A-PROC",
		Title: "TITLE",
		Children: []Node{
			&Text{"Some lines\nof text."},
		},
	}

	rawInput := multiline(
		"@A-PROC TITLE",
		"  Some lines",
		"  of text.")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestAProcCanHaveMultipleIndentedChildren(t *testing.T) {
	// given
	wantNode := &Proc{
		Name: "A-PARENT",
		Children: []Node{
			&Text{"Some text."},
			&Proc{Name: "A-CHILD"},
		},
	}

	rawInput := multiline(
		"@A-PARENT",
		"  Some text.",
		"  @A-CHILD")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestNodesCanBeNestedBeyondOneLevel(t *testing.T) {
	// given
	wantNode := &Proc{
		Name: "A-GRANDPARENT",
		Children: []Node{
			&Proc{
				Name: "A-PARENT",
				Children: []Node{
					&Proc{Name: "A-CHILD"},
				},
			},
		},
	}

	rawInput := multiline(
		"@A-GRANDPARENT",
		"  @A-PARENT",
		"    @A-CHILD")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func TestDedentsArePossibleWithinAValidNode(t *testing.T) {
	// given
	wantNode := &Proc{
		Name: "A-GRANDPARENT",
		Children: []Node{
			&Proc{
				Name: "A-PARENT",
				Children: []Node{
					&Proc{Name: "A-CHILD"},
				},
			},
			&Proc{Name: "A-PARENT-SIBLING"},
		},
	}

	rawInput := multiline(
		"@A-GRANDPARENT",
		"  @A-PARENT",
		"    @A-CHILD",
		"  @A-PARENT-SIBLING")

	input := strings.NewReader(rawInput)

	parser := NewParser(input)

	// when
	node, err := parser.Parse()

	// then
	assert.That(node != nil, t.Errorf, "the node returned must not be nil")
	assert.That(errors.Cause(err) == io.EOF, t.Errorf, "got error %q, wanted %q", err, io.EOF)
	reportDiffs(t.Errorf, node, wantNode)
}

func multiline(lines ...string) string {
	var buf bytes.Buffer
	last := len(lines) - 1
	for i, line := range lines {
		buf.WriteString(line)
		if i != last {
			buf.WriteRune('\n')
		}
	}
	return buf.String()
}

func reportDiffs(onErr func(string, ...interface{}), got, want interface{}) {
	diffs := pretty.Diff(got, want)
	for _, diff := range diffs {
		onErr(diff)
	}
	if len(diffs) > 0 {
		onErr("got: %# v", pretty.Formatter(got))
		onErr("wanted: is %# v", pretty.Formatter(want))
	}
}
