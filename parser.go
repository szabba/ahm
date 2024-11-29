// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"slices"
	"strings"
)

// ErrOnLine wraps an error with line information.
//
// It returns an error that satisfies LineNoError.
func ErrOnLine(lineNo int64, err error) error {
	if lineNo < 1 {
		msg := fmt.Sprintf("line number %d out of range", lineNo)
		panic(msg)
	}
	if err == nil {
		msg := fmt.Sprintf("wrapping nil error at line %d", lineNo)
		panic(msg)
	}
	return _LineNoError{lineNo, err}
}

// LineNoError is an error that carries a line number, indicating where in the input it occurred.
type LineNoError interface {
	LineNo() int64
	error
}

type _LineNoError struct {
	lineNo int64
	error
}

var _ LineNoError = _LineNoError{}
var _ interface{ Unwrap() error } = _LineNoError{}

func (err _LineNoError) LineNo() int64 { return err.lineNo }

func (err _LineNoError) Error() string {
	return fmt.Sprintf("at line %d: %s", err.lineNo, err.error.Error())
}

func (err _LineNoError) Unwrap() error { return err.error }

// ErrUnacceptableProcName indicates that a proc name is unacceptable.
func ErrUnacceptableProcName(name string) error { return _ErrUnacceptableProcName(name) }

type UnacceptableProcNameError interface {
	UnacceptableProcName() string
	error
}

type _ErrUnacceptableProcName string

var _ UnacceptableProcNameError = _ErrUnacceptableProcName("")

func (err _ErrUnacceptableProcName) UnacceptableProcName() string {
	return string(err)
}

func (err _ErrUnacceptableProcName) Error() string {
	return fmt.Sprintf("name %q: unnaceptable proc name", string(err))
}

// ErrUnexpectedIndent indicated that a line is unexpectedly indented.
func ErrUnexpectedIndent() error { return errUnexpectedIndent }

// ErrMismatchedIndents indicates that two subsequent non-empty lines have:
//
//   - different indentation and
//   - neither is indented further than the other.
func ErrMismatchedIndents() error {
	// Fun observation:
	// Lines form a bounded meet-semilattice with the relation "indented further than".
	return errMismatchedIndents
}

var (
	errUnexpectedIndent  = errors.New("unexpected indent")
	errMismatchedIndents = errors.New("mismatched indents")
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

type indentedNode struct {
	Node
	Indent string
}

func parseLines(in *input) ([]indentedNode, []error) {
	out := []indentedNode{}
	errs := []error{}

	for {
		err := in.Advance()
		if err != nil {
			if err != io.EOF {
				errs = append(errs, ErrOnLine(in.LineNo(), err))
			}
			return out, errs
		}

		indent, after := splitIndent(in.Line())

		if after == "" {
			text := Text("").PlacedOnLine(in.LineNo())
			nwi := indentedNode{text, indent}
			out = append(out, nwi)
			continue
		}

		if !strings.HasPrefix(after, "@") {
			text := Text(after).PlacedOnLine(in.LineNo())
			nwi := indentedNode{text, indent}
			out = append(out, nwi)
			continue
		}

		proc, err := parseProcHeader(after)
		if len(proc) > 0 {
			nwi := indentedNode{
				proc[0].PlacedOnLine(in.LineNo()),
				indent,
			}
			out = append(out, nwi)
		}
		if err != nil {
			errs = append(errs, ErrOnLine(in.LineNo(), err))
		}
	}
}

func forestify(nodes []indentedNode) ([]indentedNode, []error) {
	out := make([]indentedNode, 0, len(nodes))
	errs := []error{}

	left := nodes
	// The lenght of left will be changing as we go - tricky stuff!
	for len(left) > 0 {
		log.Printf("left = %s", left)
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
		log.Printf("parent = %s", parent)
		log.Printf("left = %s", left)
		if len(left) == 0 {
			return parent, left, errs
		}

		n := left[0]
		log.Printf("indents = %q / %q / %q", parent.Indent, childIndent, n.Indent)

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
			log.Printf("first child")
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

func parseProcHeader(after string) ([]Node, error) {
	header := strings.TrimPrefix(after, "@")
	parts := procHeaderSep.Split(header, 2)

	name := parts[0]
	if !okNames.MatchString(name) {
		return nil, ErrUnacceptableProcName(name)
	}

	if len(parts) == 1 {
		proc := Proc(name, "")
		return []Node{proc}, nil
	}

	proc := Proc(name, title(parts))
	return []Node{proc}, nil
}

var (
	procHeaderSep = regexp.MustCompile(`[[:space:]]`)
	okNames       = regexp.MustCompile(`^([[:alpha:]]|_)([[:word:]]|-)*$`)
)

func title(parts []string) string {
	switch len(parts) {
	case 1:
		return ""
	case 2:
		return parts[1]
	default:
		panic("unreachable")
	}
}

func splitIndent(s string) (indent, after string) {
	after = strings.TrimLeft(s, "\t\v\f\r ")
	indent = s[:len(s)-len(after)]
	log.Printf("%q == %q + %q", s, indent, after)
	return indent, after
}
