// This Alloy 6 model is concerned with the tree/forest structure of an Ahm document.
module ahm

open document as doc
open util/graph[doc/Line]
open util/graph[MergeTree]

// Make sure the model has examples at all.
run example {} for 4 but 8 MergeTree

// Make sure the model does not exclude some types of examples we care about.

run depth { depth[children] } for 5 but 10 MergeTree

run branching { branching[children] } for 5 but 10 MergeTree

run forest { not tree[children] } for 5 but 10 MergeTree

run branchingIndents { branching[prefixOf] } for 5 but 10 MergeTree

run siblingsWithDifferentIndents {
	some siblings
	some s1, s2: siblings | no s1.indent & s2.indent
} for 5 but 10 MergeTree

run nasty { not doc/niceIndents } for 6

run _all {
	depth[children]
	branching[children]
	branching[prefixOf]
	not tree[children]
} for 10

run aChildCanBeEmpty {
	some ran[children] & Empty
} for 5 but 10 MergeTree

run longDocument {} for 6 but 15 MergeTree, 6 Line

pred niceIndents {
	first.indent = NoIndent
	all l: Line | no l.next or l.indent = l.next.indent or l.indent.prefixOf = l.next.indent or l.indent = l.next.indent.prefixOf
}

// Check that the parsed structure has properties that we care about.

check noChildHasTwoParents {
	all c: ran[children] | one children.c
} for 4 but 8 MergeTree, 4 Line

check noChildHasTwoParentsWithDifferentIndents {
	all c: ran[children] | one c.~children.indent
} for 4 but 8 MergeTree, 4 Line

check childrenIsForest { forest[children] } for 4 but 8 MergeTree, 4 Line

check allNonRootsAreChildren {
	Line - roots[children] = ran[children]
} for 6 but 15 MergeTree, 6 Line

check allLinesBetweenTheParentAndChildAreDescendantsOfTheParent {
	all child: ran[children] |
	let parent = children.child |
	between[doc/ordering/next, parent, child] in parent.^children
} for 6 but 15 MergeTree, 6 Line

check noParentHasOnlyEmptyChildren {
	no p: dom[children] | p.children in Empty
} for 6 but 15 MergeTree, 6 Line

check noParentHasAnEmptyLastChild {
	no parent: ProcHeader | lastLine[parent.children] in Empty
} for 6 but 15 MergeTree, 6 Line

fun siblings : Line -> Line {
    { s1, s2: Line | some p: Line | s1 + s2 in p.children } + (roots[children] -> roots[children]) - iden
}

// The parse forest implied by the sequence of lines and their indents.

fun children : ProcHeader -> Line {
	ProcHeader <: roots[merges].parse
}

// Monoidal parser.

fun merge[lhs : Parse, rhs : Parse] : Parse {
	// No lines on the lhs change their parent.
	// The merging done so far should've added those where possible.
	//
	// Where possible rhs orphan lines are adopted.
	//
    // An implementation would not need to keep track of already parented rhs lines.
	// They would be the non-roots of the rhs forest.
	//
	// (The roots of the rhs forest are the same as it's orphans.)
	lhs + rhs :> parented[rhs] + adopt[lhs, rhs :> orphan[rhs]] + trackUnadopted[lhs, rhs :> orphan[rhs]]
}

fun adopt[lhs : Parse, rhs : Parse] : Parse {
	{
		parent : ran[lhs :> ProcHeader], child : ran[rhs]
	|
		{
			// The child is indented further than the parent.
			child.indent in parent.indent.^prefixOf

			// All rejected parents are indented less than the actual one.
			all alt: ran[lhs :> ProcHeader] - parent {
				child.indent in alt.indent.^prefixOf implies parent.indent in alt.indent.^prefixOf
			}

			// All rejected parents are earlier in the document than the actual one.
			all alt: ran[lhs :> ProcHeader] - parent {
				child.indent in alt.indent.^prefixOf implies alt in parent.^prev
			}
		}
	}
}

fun trackUnadopted[lhs : Parse, rhs : Parse] : Parse {
	let adopted = ran[adopt[lhs, rhs]] | rhs :> (ran[rhs] - adopted)
}

fun parented[p : Parse] : Line {
	ProcHeader.p
}

fun orphan[p : Parse] : Line {
	NoParent.p
}

// A tree describing how to merge partial parses using the monoidal parser.
abstract sig MergeTree {
	, leftChild : disj lone MergeTree
	, rightChild : disj lone MergeTree
	, mergedLine : disj lone doc/Line
	, parse : Parse
}

sig Branch extends MergeTree {}
sig Leaf extends MergeTree {}

// The set of lines merged by a given subtree.
fun mergedLines[t : MergeTree]: doc/Line {
	t.*merges.mergedLine
}

fun merges : MergeTree -> MergeTree {
	leftChild + rightChild
}

fact {
	tree[merges]

	// All branches have two children.
	all branch : Branch | #merges[branch] = 2

	// No branch has a (singular) merged line.
	no branch : Branch | some branch.mergedLine
	
	// For every line there exists a leaf.
	//all l : doc/Line | one leaf : Leaf | leaf -> l in mergedLine
//	all leaf : Leaf | one l : Line | leaf -> l in mergedLine
	bijection[mergedLine, Leaf, Line]

	// No leaf has children
	no Leaf.merges

	all t : MergeTree | consecutive[mergedLines[t]]

	// Merge tree order respects document order.
	all branch : Branch {
		lastLine[mergedLines[branch.leftChild]].next = firstLine[mergedLines[branch.rightChild]]
	}

	// The parses of all branches come from merging parses of their children.
	all branch : Branch | branch.parse = merge[branch.leftChild.parse, branch.rightChild.parse]

	// The parses of all leafs are their lines, marked as orphans.
	all leaf : Leaf | leaf.parse = NoParent -> leaf.mergedLine
}

check rootMergesAllLines {
	mergedLines[roots[merges]] = Line
} for 5 but 10 MergeTree

check firstLineIsInLeftmostLeafInTree {
	one Line or doc/ordering/first = (roots[merges].*leftChild & Leaf).mergedLine
} for 5 but 10 MergeTree

check lastLineIsRightmostLeafInTree {
	one Line or doc/ordering/last = (roots[merges].*rightChild & Leaf).mergedLine
} for 5 but 10 MergeTree

fun firstLine[lines : Line] : lone Line {
	{ l: lines | no l.^prev & lines }
}

fun lastLine[lines : Line] : lone Line {
	{ l: lines | no l.^next & lines }
}

// A parse of a subset of lines must account for all the lines in it's scope.
// That includes those that do not have a parent within that parse.

let Parse = MaybeParent -> doc/Line

fun parsedLines[p : Parse] : doc/Line {
	ran[p]
}

sig MaybeParent in NoParent + doc/ProcHeader {}

one sig NoParent {}

// Helpers

fun between[r : univ -> univ, f : univ, l : univ]: univ {
	f.^r - l.*r
}
