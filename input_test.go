// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"io"
	"strings"
	"testing"

	"github.com/szabba/assert/v3"
	"github.com/szabba/assert/v3/assertions/theerr"
	"github.com/szabba/assert/v3/assertions/theslice"
)

func TestInputEmpty(t *testing.T) {
	// given
	r := strings.NewReader("")

	in := &input{}
	in.br.Reset(r)

	// when
	lines, err := readAll(in)

	// then
	assert.UsingFmt(t.Errorf).
		That(theslice.Equal(lines, []string{""})).
		That(theerr.Is(err, io.EOF))
}

func TestInputOneLine(t *testing.T) {
	// given
	r := strings.NewReader("Abba")

	in := &input{}
	in.br.Reset(r)

	// when
	lines, err := readAll(in)

	// then
	assert.UsingFmt(t.Errorf).
		That(theslice.Equal(lines, []string{"Abba"})).
		That(theerr.Is(err, io.EOF))
}

func TestInputManyLines(t *testing.T) {
	// given
	r := strings.NewReader("Abba\nMamma Mia!")

	in := &input{}
	in.br.Reset(r)

	// when
	lines, err := readAll(in)

	// then
	assert.UsingFmt(t.Errorf).
		That(theslice.Equal(lines, []string{"Abba", "Mamma Mia!"})).
		That(theerr.Is(err, io.EOF))
}

func TestInputEmptyUnadvancingEach(t *testing.T) {
	// given
	r := strings.NewReader("")

	in := &input{}
	in.br.Reset(r)

	// when
	lines, err := readAllUnadvancingOnce(in)

	// then
	assert.UsingFmt(t.Errorf).
		That(theslice.Equal(lines, []string{"", ""})).
		That(theerr.Is(err, io.EOF))
}

func readAll(in *input) ([]string, error) {
	lines := []string{}
	for {
		err := in.Advance()

		if err != nil {
			return lines, err
		}

		lines = append(lines, in.Line())
	}
}

func readAllUnadvancingOnce(in *input) ([]string, error) {
	lines := []string{}
	for {
		err := in.Advance()

		if err != nil {
			return lines, err
		}

		lines = append(lines, in.Line())

		in.Unadvance()
		in.Advance()

		lines = append(lines, in.Line())
	}
}
