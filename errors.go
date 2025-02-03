// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"errors"
	"fmt"
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
