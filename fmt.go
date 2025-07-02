// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"fmt"
	"io"
	"iter"
)

// Fmt serializes the given docs.
func Fmt(w io.Writer, doc []Node) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("ahm.Fmt: %w", err)
		}
	}()

	firstLine := true
	for lvl, node := range walkDoc(doc) {
		if !firstLine {
			_, err = fmt.Fprintln(w)
			if err != nil {
				return err
			}
		}
		firstLine = false

		if node.proc || len(node.text) > 0 || node.escaped {
			for range lvl {
				_, err = io.WriteString(w, "\t")
				if err != nil {
					return err
				}
			}
		}

		if node.proc {
			err = writeProc(w, node)
			if err != nil {
				return err
			}
		} else {
			err = writeText(w, node)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func walkDoc(doc []Node) iter.Seq2[int, Node] {
	return func(yield func(level int, _ Node) bool) {
		stack := [][]Node{doc}
		for {
			if len(stack) == 0 {
				return
			}

			frame := stack[len(stack)-1]
			if len(frame) == 0 {
				stack = stack[:len(stack)-1]
				continue
			}

			node := frame[0]
			if !yield(len(stack)-1, node) {
				return
			}

			stack[len(stack)-1] = frame[1:]
			if len(node.children) > 0 {
				stack = append(stack, node.children)
			}
		}
	}
}

func writeProc(w io.Writer, proc Node) error {
	_, err := io.WriteString(w, "@")
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, proc.name)
	if err != nil {
		return err
	}
	if len(proc.title) > 0 {
		_, err = io.WriteString(w, " ")
		if err != nil {
			return err
		}
		_, err = io.WriteString(w, proc.title)
		return err
	}
	return nil
}

func writeText(w io.Writer, node Node) error {
	if node.escaped {
		_, err := io.WriteString(w, "|")
		if err != nil {
			return err
		}
		if len(node.text) > 0 {
			_, err = io.WriteString(w, "\t")
			if err != nil {
				return err
			}
		}
	}
	_, err := io.WriteString(w, node.text)
	return err
}
