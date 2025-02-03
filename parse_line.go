// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"io"
	"regexp"
	"strings"
)

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
	return indent, after
}
