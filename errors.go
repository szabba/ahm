// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var (
	errNoNewIndent    = errors.New("no extra indent where it is permitted")
	errStartOfSibling = errors.New("encountered start of a sibling node")
)

type LocatedError interface {
	error
	Cause() error
	At() Position
}

type locatedError struct {
	err error
	at  Position
}

func (err *locatedError) Cause() error { return err.err }
func (err *locatedError) At() Position { return err.at }

func (err *locatedError) Error() string {
	at := err.At()
	return fmt.Sprintf(
		"at %q:%d:%d: %s",
		at.SourceName(), at.Line(), at.Column(), err.Cause())
}

type WrongRuneError interface {
	error
	RuneGot() rune
	RuneWanted() rune
}

type wrongRuneError struct {
	got, wanted rune
}

func (err *wrongRuneError) Error() string {
	return fmt.Sprintf("wrong rune: got %q, wanted %q", err.got, err.wanted)
}

func (err *wrongRuneError) RuneGot() rune    { return err.got }
func (err *wrongRuneError) RuneWanted() rune { return err.wanted }

func writeRune(w io.Writer, r rune) error {
	buf := [4]byte{}
	n := utf8.EncodeRune(buf[:], r)
	_, err := w.Write(buf[:n])
	return err
}

type dedentError struct {
	dedent int
}

func (err *dedentError) Error() string {
	return fmt.Sprintf("dedent %d levels", err.dedent)
}

func (err *dedentError) Dedent() int { return err.dedent }
