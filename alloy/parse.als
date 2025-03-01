// This Alloy 6 model is concerned with the tree/forest structure of an Ahm document.
module parse

open document as doc
open util/ternary
open util/graph[doc/Line]
open util/graph[MergeTree]

// Make sure the model has examples at all.

run example {} for 4 but 8 MergeTree

// Make sure the model does not exclude some types of examples we care about.

run depth { depth[children] } for 4 but 8 MergeTree

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

run longDocument {} for 6 but 12 MergeTree, exactly 6 Line

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
} for 6 but 12 MergeTree

check allLinesBetweenTheParentAndChildAreDescendantsOfTheParent {
	all parent: dom[children] |
	all child: parent.children |
	between[parent, child] in parent.^children
} for 4 but 8 MergeTree

check allNonEmptyLinesBetweenTheParentAndChildAreDescendantsOfTheParent {
	all parent: dom[children] |
	all child: parent.children |
	between[parent, child] - Empty in parent.^children
} for 4 but 8 MergeTree

check allNonEmptyChildrenAreIndentedFurtherThanTheirParents {
	all parent: dom[children] |
	all child: parent.children - Empty |
	child.indent in parent.indent.^prefixOf
} for 5 but 10 MergeTree

check noParentHasOnlyEmptyChildren {
	no p: dom[children] | p.children in Empty
} for 5 but 10 MergeTree

check noParentHasAnEmptyLastChild {
	no parent: dom[children] | lastLine[parent.children] in Empty
} for 5 but 10 MergeTree

check noChildrenWhenAllLinesHaveTheSameIndent {
	one Line.indent implies no children
} for 3 but 6 MergeTree

// The parse forest implied by the sequence of lines and their indents.

fun children : ProcHeader -> Line {
	let rootParse = roots[merges].parse | { p : ProcHeader, c : Line | inRange[c, rootParse.p] }
}

fun parent : Line -> ProcHeader { ~children }

fun siblings : Line -> Line {
    { s1, s2: Line | s1.parent = s2.parent or no s1.parent + s2.parent }
}

// A parse of a subset of lines must account for all the lines in it's scope.
// That includes those that do not have a parent within that parse.

let Parse = LineRange -> MaybeParent

sig MaybeParent in NoParent + doc/ProcHeader {}

one sig NoParent {}

fun lines[parse : Parse] : Line {
	lines[select12[parse]]
}

fun parented[p : Parse] : LineRange {
	p.ProcHeader
}

fun orphan[p : Parse] : LineRange {
	p.NoParent
}

// Monoidal parser.

fun merge[lhs, rhs : Parse] : Parse {
	let mrs = mergeRanges[select12[lhs], select12[rhs]] |
	let kept =
		{ last, first : Line, p : ProcHeader
		| 
			{
				last -> first in mrs
				p = last.(lhs + rhs)[first]
			}
		}
	|
	let newlyAdopted =
		{ last, first : Line, p : ProcHeader
		|
			{
				last -> first in mrs
				p in dom[mrs]

				last -> first not in kept.MaybeParent

				last not in Empty

				first in p.^next
				last.deeperThan[p]

				all mid : between[p, last] & dom[mrs] {
					mid.deeperThan[p]
					mid in ProcHeader implies not last.deeperThan[mid]
					// mid in ProcHeader implies (mid.indent = last.indent or mid.deeperThan[last])
				}
			}
		}
	|
	let stillOrphaned =
		{ last, first : Line, p : NoParent
		|
			{
				last -> first in mrs

				last -> first not in (kept + newlyAdopted).MaybeParent
			}
		}
	|
		kept + newlyAdopted + stillOrphaned
}

check theEmptyTernaryRelationIsTheLeftIdentityOfMerge {
	all t : MergeTree | merge[none -> none -> none, t.parse] = t.parse
} for 5 but 10 MergeTree

check theEmptyTernaryRelationIsTheRightIdentityOfMerge {
	all t : MergeTree | t.parse = merge[t.parse, none -> none -> none]
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

// A single line range is lastLine -> firstLine pair.

let LineRange = Line one -> one Line

fun mergeRanges[lhs : LineRange, rhs : LineRange] : LineRange {
	{
		one lastLine[dom[lhs]]
		lastLine[dom[lhs]] in Empty
		some rhs

	} implies {
		let mergedRange = firstLine[dom[rhs]] -> lastLine[dom[lhs]].lhs | 
		(Line - lastLine[dom[lhs]]) <: (lhs + rhs ++ mergedRange)
		// Equivalently(?):
		// (Line - lastLine[dom[lhs]]) <: lhs + (Line - firstLine[dom[lhs]]) <: rhs + mergedRange
	
	} else {
		lhs + rhs
	}
}

fun lines[ranges : LineRange] : Line {
	{ line : Line
	|
		some l, f : Line {
			l -> f in ranges
			line in f + between[f, l] + l
		} 
	}
}

pred inRange[l : one Line, r : /* one */ LineRange] {
	l in ran[r] + between[ran[r], dom[r]] + dom[r]
}

check theEmptyBinaryRelationIsTheLeftIdentityOfMergRanges {
	all r : MergeTree.ranges | mergeRanges[none -> none, r] = r
} for 5 but 10 MergeTree

check theEmptyBinaryRelationIsTheRightIdentityOfMergeRanges {
	all r : MergeTree.ranges | r = mergeRanges[r, none -> none]
} for 5 but 10 MergeTree

check mergeRangesIsAssociative {
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
			mergeRanges[a.ranges, mergeRanges[b.ranges, c.ranges]] = mergeRanges[mergeRanges[a.ranges, b.ranges], c.ranges]
		}
	}
} for 5 but 10 MergeTree

check allButTheLastLineInARangeAreAlwaysEmpty {
	all t : MergeTree |
	all r : t.ranges |
		#lastLine[lines[r]] > 1 implies lines[r] - lastLine[lines[r]] in Empty
} for 5 but 10 MergeTree

check rangesOnATreeCoverAllTheMergedLines {
	all t : MergeTree | t.mergedLines = lines[t.ranges]
} for 5 but 10 MergeTree

check noRangesOnATreeOverlap {
	no t : MergeTree | some disj l1, l2: dom[t.ranges] {
		some lines[l1 -> t.ranges[l1]] & lines[l2 -> t.ranges[l2]]
	}
} for 5 but 10 MergeTree

// A tree describing how to merge partial parses using the monoidal parser.
// We use it for associativity checks.

abstract sig MergeTree {
	, leftChild : disj lone MergeTree
	, rightChild : disj lone MergeTree
	, mergedLine : disj lone doc/Line
	, parse : Parse
	// TODO: read these out from the parses instead. The current representation allows that.
	, ranges : LineRange
}

sig Branch extends MergeTree {}
sig Leaf extends MergeTree {}

// The set of lines merged by a given subtree.
fun mergedLines[t : MergeTree]: doc/Line {
	t.*merges.mergedLine
}

fun merges : MergeTree -> MergeTree { leftChild + rightChild }

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
	all leaf : Leaf | leaf.parse = leaf.mergedLine -> leaf.mergedLine -> NoParent

	// The line ranges of all branches come from merging the range sets of their children.
	all branch : Branch | branch.ranges = mergeRanges[branch.leftChild.ranges, branch.rightChild.ranges]

	// The line ranges of all leafs are the single-element ranges of their lines.
	all leaf : Leaf | leaf.ranges = leaf.mergedLine -> leaf.mergedLine
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

// Helpers

fun firstLine[lines : Line] : lone Line {
	{ l: lines | no l.^prev & lines }
}

fun lastLine[lines : Line] : lone Line {
	{ l: lines | no l.^next & lines }
}

fun between[f : one Line, l : one Line] : Line {
	f.^next - l.*next
}
