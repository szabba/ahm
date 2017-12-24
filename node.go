// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package ahm

type Document struct {
	nodes []Node
}

//go:generate irgen Node NodeConsumer

type Node interface {
	FeedTo(NodeConsumer)
}

type NodeConsumer interface {
	Proc(Name, Title string, Children []Node)
	Text(Text string)
}
