// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

import (
	"bytes"
	"strings"
)

type textBuilder struct {
	buf bytes.Buffer
}

func (builder *textBuilder) Write(p []byte) (int, error) {
	return builder.buf.Write(p)
}

func (builder *textBuilder) Build() *Text {
	content := builder.buf.String()
	content = strings.TrimSuffix(content, "\n")
	return &Text{content}
}
