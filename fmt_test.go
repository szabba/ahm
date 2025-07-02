// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm_test

import (
	"bytes"
	"embed"
	"fmt"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	"github.com/szabba/assert/v3"
	"github.com/szabba/assert/v3/assertions/theerr"

	"github.com/szabba/ahm"
)

//go:embed testdata/fmt/*
var fmtFs embed.FS

func TestFmt(t *testing.T) {
	// given
	cases := map[string][]ahm.Node{

		"nil": nil,

		"empty": {},

		"some-text": {
			ahm.Text("Some text lines."),
			ahm.Text("Nothing fancy."),
		},

		"some-procs": {
			ahm.Proc("PROC-1", ""),
			ahm.Proc("PROC-2", "Title"),
			ahm.Proc("PROC-3", ""),
		},

		"proc-tree": {
			ahm.Proc("GRANDPARENT", "",
				ahm.Proc("PARENT", "",
					ahm.Proc("CHILD", ""),
					ahm.Proc("SIBLING", "")),
				ahm.Proc("PARENT-SIBLING", "")),
		},

		"tree-with-text": {
			ahm.Proc("H1", "Document title"),
			ahm.Text(""),
			ahm.Proc("STEP", "Action"),
			ahm.Proc("STEP", "Action"),
			ahm.Proc("STEP", "",
				ahm.Text("More complex action."),
				ahm.Text("Requires a longer explanation."),
				ahm.Text(""),
				ahm.Proc("STEP", "Sub-action."),
				ahm.Proc("STEP", "Sub-action."),
				ahm.Proc("STEP", "Sub-action.")),
			ahm.Text(""),
			ahm.Proc("STEP", "Action"),
		},

		"embed-code": {
			ahm.Proc("H1", "Document title"),
			ahm.Text(""),
			ahm.Proc("CODE", "go",
				ahm.EscapedText("package main"),
				ahm.EscapedText(""),
				ahm.EscapedText("func main() {"),
				ahm.EscapedText("}")),
			ahm.Text(""),
			ahm.Text("Some explanatory text."),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {

			fname := fmt.Sprintf("testdata/fmt/%s.ahm", name)
			wantBytes, err := fmtFs.ReadFile(fname)
			assert.UsingFmt(t.Fatalf).That(theerr.IsNil(err))

			buf := new(bytes.Buffer)

			// when
			err = ahm.Fmt(buf, tt)

			// then
			assert.UsingFmt(t.Errorf).That(theerr.IsNil(err))

			diff := udiff.Unified("want", "got", string(wantBytes), buf.String())

			if diff == "" {
				return
			}
			t.Fatal("\n" + diff)
		})
	}
}
