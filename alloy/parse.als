// This Alloy 6 model is concerned with the tree/forest structure of an Ahm document.
module parse

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

run nasty { not doc/niceIndents } for 5 but 10 MergeTree

run _all {
	depth[children]
	branching[children]
	branching[prefixOf]
	not tree[children]
} for 5 but 10 MergeTree

run aChildCanBeEmpty {
	some ran[children] & Empty
} for 5 but 10 MergeTree

run longDocument {} for 6 but 15 MergeTree, exactly 6 Line

pred niceIndents {
	first.indent = NoIndent
	all l: Line | no l.next or l.indent = l.next.indent or l.indent.prefixOf = l.next.indent or l.indent = l.next.indent.prefixOf
}

// Check that the parsed structure has properties that we care about.

check noChildHasTwoParents {
	all c: ran[children] | one children.c
} for 4 but 8 MergeTree

check noChildHasTwoParentsWithDifferentIndents {
	all c: ran[children] | one c.~children.indent
} for 4 but 8 MergeTree

check childrenIsForest { forest[children] } for 4 but 8 MergeTree

check allNonRootsAreChildren {
	Line - roots[children] = ran[children]
} for 6 but 15 MergeTree

check allLinesBetweenTheParentAndChildAreDescendantsOfTheParent {
	all parent: dom[children] |
	all child: parent.children |
	between[doc/ordering/next, parent, child] in parent.^children
} for 5 but 11 MergeTree

check allNonEmptyChildrenAreIndentedFurtherThanTheirParents {
	all parent: dom[children] |
	all child: parent.children - Empty |
	child.indent in parent.indent.^prefixOf
} for 5 but 11 MergeTree

check noParentHasOnlyEmptyChildren {
	no p: dom[children] | p.children in Empty
} for 5 but 11 MergeTree

check noParentHasAnEmptyLastChild {
	no parent: dom[children] | lastLine[parent.children] in Empty
} for 5 but 11 MergeTree

check noChildrenWhenAllLinesHaveTheSameIndent {
	one Line.indent implies no children
} for 3 but 7 MergeTree

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

check theEmptyBinaryRelationIsTheLeftIdentityOfMerge {
	all p : MergeTree.parse | merge[none -> none, p] = p
} for 5 but 10 MergeTree

check theEmptyBinaryRelationIsTheRightIdentityOfMerge {
	all p : MergeTree.parse | p = merge[p, none -> none]
} for 5 but 10 MergeTree

check mergeIsAssociative {
	all disj a, b, c : MergeTree {
		{
			one p : MergeTree {
				{
					a = p.leftChild
					b = p.rightChild.leftChild
					c = p.rightChild.rightChild
				} or {
					a = p.leftChild.leftChild
					b = p.leftChild.rightChild
					c = p.rightChild
				}
			}
		} implies {
				merge[a.parse, merge[b.parse, c.parse]] = merge[merge[a.parse, b.parse], c.parse]
		}
	}
} for 5 but 10 MergeTree

fun adopt[lhs : Parse, rhs : Parse] : Parse {
	{ p : ran[lhs :> ProcHeader]
	, c : ran[rhs]
	| bestParent[lhs, p, c]
	}
}

pred bestParent[lhs : Parse, parent : one ProcHeader, child : one Line] {

	// The best parent can be a parent.
	canBeParent[parent, child]

	all mid: between[doc/ordering/next, parent, child] {

		// No line between the parent and child can be a parent for the child.
		not canBeParent[mid, child]

		// All the lines between the parent and child can be children of the parent.
		canBeParent[parent, mid]
	}
}

pred canBeParent[candidate : one Line, child : one Line] {
	canBeParent2[candidate, child]

	child in Empty implies some later : Line - Empty {
		later in child.^next
		canBeParent2[candidate, later]
	}
}

pred canBeParent2[candidate : one Line, child : one Line] {

	candidate in ProcHeader

	// The non-empty child must be indented further than a valid parent.
	child in Line - Empty implies child.deeperThan[candidate]

	// A valid parent precedes the child.
	child in candidate.^next

	// All non-empty lines between a valid parent and child must be indented further than the parent.
	all mid : between[doc/ordering/next, candidate, child] - Empty {
		mid.deeperThan[candidate]
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
	
	// For every line there exists a leaf (and vice versa).
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
