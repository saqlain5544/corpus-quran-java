// Package bktree implements a Burkhard-Keller tree over a metric
// space (we use Levenshtein distance) for sub-linear fuzzy lookup
// over a vocabulary.
//
// Reference: Wikipedia — BK-tree. Burkhard & Keller 1973.
//
// Average lookup time is Θ(c^k · log n) for distance threshold k,
// vs. Θ(n) for linear scan. For our 5K English vocabulary with k=1
// or k=2, this means ~50-200 distance computations per query
// instead of 5,000.
package bktree

import (
	"sort"

	"quranreader/types"
)

// Node is a BK-tree node holding one vocabulary word and child
// subtrees keyed by their distance to the parent word.
//
// The tree is built by inserting words one at a time (see Insert).
// The order of insertion affects balance — for a balanced tree,
// shuffle the input before inserting.
type Node struct {
	Word     string
	Children map[int]*Node // edge distance → subtree
}

// Tree is the root of a BK-tree. Implements types.BKTreeIface.
type Tree struct {
	Root *Node
	Size int
}

// New returns an empty BK-tree.
func New() *Tree { return &Tree{} }

// Distance is the metric function. We use Levenshtein; replace
// with another metric for other use cases.
type Distance func(a, b string) int

// Insert adds a word to the tree. O(distances) where distances is
// the length of the path from root to leaf.
//
// If `dist` is nil, the package-internal Levenshtein is used.
func (t *Tree) Insert(word string, dist Distance) {
	if dist == nil {
		dist = levenshtein
	}
	if t.Root == nil {
		t.Root = &Node{
			Word:     word,
			Children: make(map[int]*Node, 4),
		}
		t.Size = 1
		return
	}
	node := t.Root
	for {
		d := dist(word, node.Word)
		if child, ok := node.Children[d]; ok {
			node = child
		} else {
			node.Children[d] = &Node{
				Word:     word,
				Children: make(map[int]*Node, 4),
			}
			t.Size++
			return
		}
	}
}

// Query implements types.BKTreeIface — returns all vocabulary
// words within edit distance k of target, sorted by distance then
// alphabetically. Uses the package's default Levenshtein metric.
//
// Average case O(c^k · log n) by the triangle-inequality prune
// rule. Worst case O(n) when the tree is degenerate.
//
// Implementation note: iterative DFS to avoid Go's recursion-depth
// issues on degenerate trees (we observed depths > 1000 in early
// versions when the root word had small edit distance to many
// others).
func (t *Tree) Query(target string, k int) []types.BKMatch {
	if t.Root == nil {
		return nil
	}
	var out []types.BKMatch
	stack := []*Node{t.Root}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		// NB: do NOT cap the distance at k here — the prune rule
		// needs the real distance to decide which subtrees to
		// recurse into. (A capped distance > k gives wrong
		// prune decisions.)
		d := levenshtein(target, node.Word)
		if d <= k {
			out = append(out, types.BKMatch{Word: node.Word, Dist: d})
		}
		for childDist, child := range node.Children {
			// Triangle-inequality prune: child at edge distance c
			// from this node has lower-bound distance
			// c - d to target (we only care about the lower bound
			// for the prune). Recurse only if it could be ≤ k.
			if childDist-d <= k {
				stack = append(stack, child)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Dist != out[j].Dist {
			return out[i].Dist < out[j].Dist
		}
		return out[i].Word < out[j].Word
	})
	return out
}

// levenshtein is the package-internal metric. Kept here so the
// BK-tree is self-contained (callers don't need to thread a
// Distance function through).
//
// Implementation: standard dynamic-programming with O(|a|·|b|)
// time and O(min(|a|, |b|)) space (rolling rows). No early-exit
// cap — the BK-tree prune needs the real distance.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	// Ensure b is the shorter string (rolling-row optimization).
	if len(a) < len(b) {
		a, b = b, a
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(
				curr[j-1]+1,      // insertion
				prev[j]+1,        // deletion
				prev[j-1]+cost,   // substitution
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
