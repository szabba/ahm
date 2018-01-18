// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahmerr

import (
	"fmt"

	"github.com/pkg/errors"
)

type UnexpectedRuneError interface {
	error
	RuneGotVsWanted() (got, want rune)
}

func IsUnexpectedRune(err error) bool {
	_, ok := errors.Cause(err).(UnexpectedRuneError)
	return ok
}

func UnexpectedRune(err error) (got, want rune, ok bool) {
	var unexpected UnexpectedRuneError
	unexpected, ok = errors.Cause(err).(UnexpectedRuneError)
	if ok {
		got, want = unexpected.RuneGotVsWanted()
	}
	return got, want, ok
}

func MustUnexpectedRune(err error) (got, want rune) {
	return errors.Cause(err).(UnexpectedRuneError).RuneGotVsWanted()
}

type unexpectedRune struct {
	got, want rune
}

func NewUnexpectedRuneError(got, want rune) error {
	err := new(unexpectedRune)
	err.got = got
	err.want = want
	return err
}

func (err *unexpectedRune) Error() string {
	return fmt.Sprintf("got rune %q, wanted %q", err.got, err.want)
}

func (err *unexpectedRune) RuneGotVsWanted() (got, want rune) {
	return err.got, err.want
}
