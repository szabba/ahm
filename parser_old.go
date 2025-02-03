// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build !monoidal

package ahm

import (
	"errors"
	"io"
	"slices"
	"strings"
)

// Parse the text of r producing top-level Ahm nodes.
func Parse(r io.Reader) ([]Node, error) {
	in := &input{}
	in.br.Reset(r)

	toplevel, errs := parseLines(in)

	toplevel, errs2 := forestify(toplevel)

	out := slices.Grow([]Node{}, len(toplevel))
	for _, nwi := range toplevel {
		out = append(out, nwi.Node)
	}

	err := errors.Join(append(errs2, errs...)...)

	return out, err
}

func forestify(nodes []indentedNode) ([]indentedNode, []error) {
	out := make([]indentedNode, 0, len(nodes))
	errs := []error{}

	left := nodes
	// The lenght of left will be changing as we go - tricky stuff!
	for len(left) > 0 {
		n := left[0]

		if n.Indent != "" {
			err := ErrOnLine(n.LineNo(), ErrMismatchedIndents())
			errs = append(errs, err)
		}
		if n.Text() {
			out = append(out, n)
			left = left[1:]
			continue
		}

		folded, rest, foldErrs := fold(n, left[1:])
		out = append(out, folded)
		errs = append(errs, foldErrs...)
		left = rest
	}

	return out, errs
}

func fold(parent indentedNode, left []indentedNode) (indentedNode, []indentedNode, []error) {

	errs := []error{}
	childIndent := parent.Indent

	for {
		if len(left) == 0 {
			return parent, left, errs
		}

		n := left[0]

		if strings.HasPrefix(parent.Indent, n.Indent) {
			// dedent - some ancestor might deal with a mismatch
			break
		}

		if n.Empty() {
			parent.children = append(parent.children, n.Node)
			left = left[1:]
			continue
		}

		if !strings.HasPrefix(n.Indent, parent.Indent) {
			// (?) mismatch
			break
		}

		// if !strings.HasPrefix(n.Indent, childIndent) {
		// 	// (?) mismatch
		// 	break
		// }

		if strings.HasPrefix(n.Indent, parent.Indent) && childIndent == parent.Indent {
			// indent - n is the first non-empty child
			childIndent = n.Indent
		}

		if childIndent == n.Indent {
			// n is a non-empty child indented the same as the first
			if n.Text() {
				// text
				parent.children = append(parent.children, n.Node)
				left = left[1:]
				continue
			}

			// proc
			folded, rest, foldErrs := fold(n, left[1:])
			parent.children = append(parent.children, folded.Node)
			left = rest
			errs = append(errs, foldErrs...)
			continue
		}

		if childIndent != n.Indent && strings.HasPrefix(n.Indent, parent.Indent) {
			// mismatched child
			errs = append(errs, ErrOnLine(n.LineNo(), ErrMismatchedIndents()))

			if n.Text() {
				// text
				parent.children = append(parent.children, n.Node)
				left = left[1:]
				continue
			}

			// proc
			folded, rest, foldErrs := fold(n, left[1:])
			parent.children = append(parent.children, folded.Node)
			left = rest
			errs = append(errs, foldErrs...)
			continue
		}

		panic("unreachable")
	}
	return parent, left, errs
}
