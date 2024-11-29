// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"bufio"
	"errors"
	"slices"
	"strings"
)

// ErrTooManyLines indicates that the input has more lines than the parser supports.
func ErrTooManyLines() error { return errTooManyLines }

var errTooManyLines = errors.New("input has too many lines")

type input struct {
	now, future []state
	br          bufio.Reader
}

type state struct {
	lineNo int64
	line   string
	err    error
}

func (in *input) Advance() error {

	// If we have a future state queued up... use it!
	if len(in.future) > 0 {
		in.now = []state{in.future[0]}
		in.future = in.future[1:]
		return in.now[0].err
	}

	// If we've failed already, stay put.
	if len(in.now) == 1 && in.now[0].err != nil {
		return in.now[0].err
	}

	// Otherwise, read more!
	l, err := in.br.ReadString('\n')
	l = strings.TrimSuffix(l, "\n")

	// Update current state
	lineNo := int64(1)
	if len(in.now) > 0 {
		lineNo = in.now[0].lineNo + 1
	}

	// TODO: ErrTooManyLines

	in.now = []state{{line: l}}
	in.now[0].lineNo = lineNo

	// Queue any I/O error to be returned in the future
	if err != nil {
		in.future = []state{{
			lineNo: in.now[0].lineNo,
			err:    err,
		}}
	}

	return in.now[0].err
}

func (in *input) Unadvance() {
	if len(in.now) == 0 {
		panic("cannot unadvance")
	}

	in.now, in.future = []state{}, slices.Insert(in.future, 0, in.now[0])
}

func (in *input) LineNo() int64 {
	in.somethingRead()
	return in.now[0].lineNo
}

func (in *input) Line() string {
	in.somethingRead()
	return in.now[0].line
}

func (in *input) somethingRead() {
	if len(in.now) != 1 {
		panic("no line read yet")
	}
	if in.now[0].lineNo < 0 {
		panic(ErrTooManyLines())
	}
	if in.now[0].err != nil {
		panic(in.now[0].err)
	}
}
