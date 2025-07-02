// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"fmt"
	"strings"
)

type Node struct {
	lineNo int64

	proc        bool
	name, title string
	children    []Node

	text    string
	escaped bool
}

func Proc(name, title string, children ...Node) Node {
	return Node{
		proc:     true,
		name:     name,
		title:    title,
		children: children,
	}
}

func Text(text string) Node {
	if strings.ContainsRune(text, '\n') {
		err := fmt.Errorf("string %q contains %q", text, '\n')
		panic(err)
	}
	return Node{text: text, escaped: startsWithSpace(text)}
}

func startsWithSpace(s string) bool {
	return !strings.HasPrefix(s, strings.TrimSpace(s))
}

func EscapedText(text string) Node {
	return Node{text: text, escaped: true}
}

func (n Node) Proc() bool    { return n.proc }
func (n Node) Text() bool    { return !n.proc }
func (n Node) Escaped() bool { return n.escaped }

func (n Node) Empty() bool { return n.Text() && n.NodeText() == "" }

func (n Node) LineNo() int64 { return n.lineNo }

func (n Node) PlacedOnLine(lineNo int64) Node {
	if lineNo < 1 {
		msg := fmt.Sprintf("invalid line number %d", lineNo)
		panic(msg)
	}
	n.lineNo = lineNo
	return n
}

func (n Node) Name() string {
	n.mustProc()
	return n.name
}

func (n Node) Title() string {
	n.mustProc()
	return n.title
}

func (n Node) Children() []Node {
	n.mustProc()
	return n.children
}

func (n Node) NodeText() string {
	n.mustText()
	return n.text
}

func (n Node) mustProc() {
	if !n.proc {
		msg := fmt.Sprintf("%s is not a proc", n)
		panic(msg)
	}
}

func (n Node) mustText() {
	if n.proc {
		msg := fmt.Sprintf("%s is not text", n)
		panic(msg)
	}
}

func (n Node) String() string {
	var buf strings.Builder

	if !n.proc {
		if !n.escaped {

			buf.WriteString(`Text(`)
		} else {
			buf.WriteString(`EscapedText(`)
		}
		fmt.Fprintf(&buf, "%q", n.text)
		buf.WriteString(`)`)
		return buf.String()
	}

	buf.WriteString(`Proc(`)
	fmt.Fprintf(&buf, "%q", n.name)
	buf.WriteString(`, `)
	fmt.Fprintf(&buf, "%q", n.title)

	for _, c := range n.children {
		buf.WriteString(`, `)
		buf.WriteString(c.String())
	}

	buf.WriteString(`)`)
	return buf.String()
}
