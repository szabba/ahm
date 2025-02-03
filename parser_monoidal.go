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

type parse struct{}

func oneLineParse(n indentedNode) parse {
	return parse{}
}

func (p parse) nodes() []Node { return nil }

func (p parse) errors() []error { return nil }

func (lhs parse) merge(rhs parse) parse {
	return parse{}
}
