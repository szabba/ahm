// This Alloy 6 model is concerned with the Ahm document as input.
module document

open util/ordering[Line]
open util/graph[Line]
open util/graph[Indent]

// Make sure the model has examples at all.
run example {} for 6

// Make sure the model does not exclude some types of examples we care about.

run branchingIndents { branching[prefixOf] } for 6

run nasty { not niceIndents } for 6

run longDocument {} for 6 but exactly 6 Line

pred niceIndents {
	first.indent = NoIndent
	all l: Line | no l.next or l.indent = l.next.indent or l.indent.prefixOf = l.next.indent or l.indent = l.next.indent.prefixOf
}

// A line is either empty, a proc header, or non-empty text.
// - empty (this includes all-whitespace lines).
// - a proc header (this is all lines with the first non-whitespace character being @)
// - a text line (anything else).

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
}

fact {
	// We treat documents as sequences of newline-separated lines.
	// So an empty file has one empty line.
	//
	// This is different than having newline-terminated lines.
	some Line
}

pred consecutive[lines : Line] {
	no lines or one lines or one first : lines | lines = first.*(lines <: ordering/next :> lines)
}

check allAreConsecutive {
	consecutive[Line]
} for 6

check anyGapMakesNonConsecutive {
	all gap : Line - first - last | not consecutive[Line - gap]
} for 6

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

	// NoIndent is the root indent level.
	rootedAt[prefixOf, NoIndent]

	// We only care about indents that are line prefixes.
    // This constraint ensures models do not contain other, junk indents.
	rootedAt[~prefixOf, Line.indent]
}

// A few custom graph predicates.

pred depth[t: univ -> univ] { some (dom[t] & ran[t]) }

pred branching[t: univ -> univ] { some p: dom[t] | not lone p.t }
