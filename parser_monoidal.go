// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build monoidal

package ahm

import (
	"errors"
	"io"
)

// Parse the text of r producing top-level Ahm nodes.
func Parse(r io.Reader) ([]Node, error) {
	in := &input{}
	in.br.Reset(r)

	nodes, errs := parseLines(in)

	out := parse{}

	for _, n := range nodes {
		out = out.merge(oneLineParse(n))
	}

	err := errors.Join(append(out.errors(), errs...)...)

	return out.nodes(), err
}

type parse struct {
	groups []group
}

func oneLineParse(n indentedNode) parse {
	return parse{
		groups: []group{
			{lines: []Node{n.Node}},
		},
	}
}

func (p parse) nodes() []Node {
	nodes := []Node{}
	for _, g := range p.groups {
		nodes = append(nodes, g.lines...)
	}
	return nodes
}

func (p parse) errors() []error { return nil }

func (lhs parse) merge(rhs parse) parse {
	merged := parse{}
	merged.groups = lhs.mergeGroups(rhs)

	return merged
}

func (lhs parse) mergeGroups(rhs parse) []group {

	if len(lhs.groups) == 0 {
		return rhs.groups
	}

	if len(rhs.groups) == 0 {
		return lhs.groups
	}

	gs := make([]group, 0, len(lhs.groups)+len(rhs.groups))

	if lhs.groups[len(lhs.groups)-1].last().Empty() {

		gs = append(gs, lhs.groups[0:len(lhs.groups)-1]...)

		lastLeft, firstRight := lhs.groups[len(lhs.groups)-1], rhs.groups[0]

		midLines := make([]Node, 0, len(lastLeft.lines)+len(firstRight.lines))
		midLines = append(midLines, lastLeft.lines...)
		midLines = append(midLines, firstRight.lines...)

		mid := group{lines: midLines}
		gs = append(gs, mid)

		gs = append(gs, rhs.groups[1:]...)

		return gs
	}

	gs = append(gs, lhs.groups...)
	gs = append(gs, rhs.groups...)

	return gs
}

type group struct {
	lines []Node // Not indentedNode? Curious...
}

func (g group) last() Node {
	if len(g.lines) == 0 {
		panic("uninitialized group")
	}
	return g.lines[len(g.lines)-1]
}
