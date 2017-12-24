// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

type ParserOption struct {
	f func(*Parser)
}

func (opt ParserOption) applyTo(p *Parser) {
	if p == nil || opt.f == nil {
		return
	}
	opt.f(p)
}

func UseRangeMap(ranges *RangeMap) ParserOption {
	return ParserOption{func(p *Parser) {
		p.rangeMap = ranges
	}}
}

func SourceName(name string) ParserOption {
	return ParserOption{func(p *Parser) {
		p.pos.srcName = name
	}}
}
