module ahm

open util/graph[Line]
open util/graph[Indent]

run example {}

run depth { depth[children] } for 5

run branching { branching[children] } for 5

run forest { not tree[children] } for 5

run branchingIndents { branching[prefixOf] }

run siblingsWithDifferentIndents {
	some siblings
	some s1, s2: siblings | no s1.indent & s2.indent
} for 5

run nasty { not wellFormedInput }

run _all {
	depth[children]
	branching[children]
	branching[prefixOf]
	not tree[children]
} for 5

run aChildCanBeEmpty {
	some ran[children] & Empty
} for 5

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

check childrenIsForest { forest[children] } for 6

pred depth[t: univ -> univ] { some (dom[t] & ran[t]) }
pred branching[t: univ -> univ] { some p: dom[t] | #p.t > 1 }

pred wellFormedInput {
	first.indent = NoIndent
	all l: Line | no l.nextLine or l.indent = l.nextLine.indent or l.indent.prefixOf = l.nextLine.indent or l.indent = l.nextLine.indent.prefixOf
}

fun siblings : Line -> Line {
    { s1, s2: Line | some p: Line | s1 + s2 in p.children } + (roots[children] -> roots[children]) - iden
}

fun children : ProcHeader -> Line {
	let step0 = nonEmptyChildren |
	let step1 = 
		// Add parents for empty lines.
		step0 + {
			p: ProcHeader, c: Empty {
				some cs: Line {
					cs in p.step0
					c in between[nextLine, p, cs]

					no altP: ProcHeader | altP in between[step0, p, cs]
				}
			}
		}
	| step1
}

fun nonEmptyChildren : ProcHeader -> (Line - Empty) {
	let step0 = {
		// We are intentionally not worrying about empty lines for now.
		p: ProcHeader, c : Line - Empty {
			// A child candidate must come after a parent candidate.
			c in p.^nextNonEmptyLine
			// The indent of a non-empty child is further than that of a parent candidate.
			c.indent in p.indent.^prefixOf
		}
	} | let step1 = {
		p, c : Line {
			p -> c in step0
			// The parent of a non-empty line is the closest parent candidate.
			all pc: step0.c - p | p in pc.^nextNonEmptyLine
		}
	} | let step2 = {
		p, c: Line {
			p -> c in step1
			// All non-empty-lines between parent and child must be descendants of the parent.
			between[nextNonEmptyLine, p, c] in p.^step1
		}
	} | step2
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

sig Empty extends Line {}
sig ProcHeader extends Line {}
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

fun between[nextL : Line -> Line, p : Line, n : Line]: Line {
	p.^nextL - n.*nextL
}

sig FurtherIndent extends Indent {} 
one sig NoIndent extends Indent {}

abstract sig Indent {
	, prefixOf : set Indent
}

fact {
	// The indents form a tree under a strict (< vs <=) order of prefixing.
	tree[prefixOf]

	// NoIndent should be the root indent level.
	rootedAt[prefixOf, NoIndent]

	// Only allow models where all indents are either indents of lines or their prefixes.
	// Aka, prune useless branches.
	rootedAt[~prefixOf, Line.indent]
}
