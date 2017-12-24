// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

type RangeMap struct {
	toRange map[Node]Range
}

func newRangeMap() *RangeMap {
	rs := new(RangeMap)
	rs.toRange = make(map[Node]Range)
	return rs
}

func (rs *RangeMap) Get(node Node) Range {
	return rs.toRange[node]
}

func (rs *RangeMap) HasRangeFor(node Node) bool {
	_, ok := rs.toRange[node]
	return ok
}

func (rs *RangeMap) set(node Node, rang Range) {
	rs.toRange[node] = rang
}
