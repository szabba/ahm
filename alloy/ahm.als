// This Alloy 6 model is concerned with the tree/forest structure of an Ahm document.
module ahm

open util/graph[Line]
open util/graph[Indent]

// Make sure the model has examples at all.
run example {}

// Make sure the model does not define out some types of examples we care about.
run depth { depth[children] } for 5

run branching { branching[children] } for 5

run forest { not tree[children] } for 5

run branchingIndents { branching[prefixOf] }

run siblingsWithDifferentIndents {
	some siblings
	some s1, s2: siblings | no s1.indent & s2.indent
} for 5

run nasty { not niceIndents }

run _all {
	depth[children]
	branching[children]
	branching[prefixOf]
	not tree[children]
} for 5

run aChildCanBeEmpty {
	some ran[children] & Empty
} for 5


pred depth[t: univ -> univ] { some (dom[t] & ran[t]) }
pred branching[t: univ -> univ] { some p: dom[t] | #p.t > 1 }

// Check that the parsed structure has properties that we care about.

check childrenIsForest { forest[children] } for 6

check allNonRootsAreChildren {
	Line - roots[children] = ran[children]
}

check allLinesBetweenTheParentAndChildAreDescendantsOfTheParent {
	all child: ran[children] |
	let parent = children.child |
	between[nextLine, parent, child] in parent.^children
}

check noParentHasOnlyEmptyChildren {
	no p: dom[children] | p.children in Empty
}

pred niceIndents {
	first.indent = NoIndent
	all l: Line | no l.nextLine or l.indent = l.nextLine.indent or l.indent.prefixOf = l.nextLine.indent or l.indent = l.nextLine.indent.prefixOf
}

fun siblings : Line -> Line {
    { s1, s2: Line | some p: Line | s1 + s2 in p.children } + (roots[children] -> roots[children]) - iden
}

// The parse forest implied by the sequence of lines and their indents.
fun children : ProcHeader -> Line {
	nonEmptyChildren + {
		p: ProcHeader, c: Empty {
			some cs: Line {
				cs in p.nonEmptyChildren
				c in between[nextLine, p, cs]

				no altP: ProcHeader | altP in between[nonEmptyChildren, p, cs]
			}
		}
	}
}

fun nonEmptyChildren : ProcHeader -> (Line - Empty) {
	{
		p, c : Line {
			p -> c in underNearestParentCandidate
			between[nextNonEmptyLine, p, c] in p.^underNearestParentCandidate
		}
	}
}

fun underNearestParentCandidate : ProcHeader -> (Line - Empty) {
	{
		p, c : Line {
			p -> c in afterAndIndentedFurther
			all pc: afterAndIndentedFurther.c - p | p in pc.^nextNonEmptyLine
		}
	}
}

fun afterAndIndentedFurther : ProcHeader -> (Line - Empty) {
	{ p, c : Line | p -> c in nonEmptyAfter and c.indent in p.indent.^prefixOf }
}

fun nonEmptyAfter : ProcHeader -> (Line - Empty) {
	{ p : ProcHeader, c : (Line - Empty) | c in p.^nextNonEmptyLine }
}

fun nextNonEmptyLine : (Line - Empty) -> (Line - Empty) {
	{
		p, n: (Line - Empty) {
			// The next non-empty line is reachable from the previous one.
			n in p.^nextLine

			// No line between the previous and next non-empty one is non-empty.
			no between[nextLine, p, n] & (Line - Empty)
		}
	}
}

// A line is either empty, a proc header, or non-empty text.
// - empty (this includes all-whitespace lines).
// - a proc header (this is all lines with the first non-whitespace character being @

// An empty line - this includes all-whitespace lines.
sig Empty extends Line {}

// A proc header, ie a non-empty line where @ (the at-sign) is the first non-whitespace character.
// Not all of them are error-free, but we can still determine a tree/forest structure when some have errors.
sig ProcHeader extends Line {}

// A text line is one that's neither empty nor a proc header.
//
// The full spec has ways to escape whitespace / at-signs at the beginning of a text line.
sig Text extends Line {}

abstract sig Line {
	, indent : Indent
	, nextLine : disj lone Line
}

fact {
	// We treat documents as sequences of newline-separated lines.
	// So an empty file has one empty line.
	//
	// This is different than having newline-terminated lines.
	some Line
	
	// One line is first.
	one l: Line | l.*nextLine = Line
}

fun first : one Line {
	{ l: Line | l.*nextLine = Line }
}

// The set of lines reachable from p but not n, following edges of nextL, excluding p and n.
fun between[nextL : Line -> Line, p : Line, n : Line]: Line {
	p.^nextL - n.*nextL
}

// An Indent is either the empty indent (NoIndent) or futher than some other indent.
//
// We don't model the characters that the indents differ by.
// They are irrelevant to the properties we want to check the definitions for.

sig FurtherIndent extends Indent {} 
one sig NoIndent extends Indent {}

abstract sig Indent {
	, prefixOf : set Indent
}

fact {
	// The indents form a tree under a strict (< vs <=) order of prefixing.
	tree[prefixOf]

	// NoIndent is be the root indent level.
	rootedAt[prefixOf, NoIndent]

	// We only care about indents that are line prefixes.
    // This constraint ensures models do not contain other, junk indents.
	rootedAt[~prefixOf, Line.indent]
}
