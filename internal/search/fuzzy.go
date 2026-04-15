package search

import (
	"strings"
	"unicode"
)

type Match struct {
	Score     int
	Positions []int
}

// FuzzyMatch performs a fast, ranked subsequence match with bonuses for
// contiguous runs and word boundaries. If exact subsequence match fails, it
// falls back to single-typo tolerant matching (deletion or adjacent swap) with
// a penalty.
func FuzzyMatch(query, candidate string) (Match, bool) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Match{Score: 0}, true
	}
	q := []rune(query)
	c := []rune(candidate)

	if m, ok := subseq(q, c); ok {
		return m, true
	}

	// Typo tolerance: allow one deletion in query OR one adjacent swap.
	if len(q) <= 24 && len(c) <= 256 {
		best := Match{Score: -1}
		for i := range q {
			q2 := append([]rune(nil), q[:i]...)
			q2 = append(q2, q[i+1:]...)
			if len(q2) == 0 {
				continue
			}
			if m, ok := subseq(q2, c); ok {
				m.Score -= 40 // penalty
				if m.Score > best.Score {
					best = m
				}
			}
		}
		for i := 0; i+1 < len(q); i++ {
			q2 := append([]rune(nil), q...)
			q2[i], q2[i+1] = q2[i+1], q2[i]
			if m, ok := subseq(q2, c); ok {
				m.Score -= 60 // slightly larger penalty
				if m.Score > best.Score {
					best = m
				}
			}
		}
		if best.Score >= 0 {
			return best, true
		}
	}
	return Match{}, false
}

func subseq(q, c []rune) (Match, bool) {
	ql := make([]rune, len(q))
	for i, r := range q {
		ql[i] = unicode.ToLower(r)
	}
	cl := make([]rune, len(c))
	for i, r := range c {
		cl[i] = unicode.ToLower(r)
	}

	pos := make([]int, 0, len(ql))
	ci := 0
	for qi := 0; qi < len(ql); qi++ {
		found := -1
		for ci < len(cl) {
			if cl[ci] == ql[qi] {
				found = ci
				ci++
				break
			}
			ci++
		}
		if found < 0 {
			return Match{}, false
		}
		pos = append(pos, found)
	}

	// Score: base for each char + bonuses for contiguous and word boundaries.
	score := 0
	last := -2
	for i, p := range pos {
		score += 10
		if p == last+1 {
			score += 18 // contiguous run bonus
		}
		if p == 0 {
			score += 12
		} else {
			prev := c[p-1]
			cur := c[p]
			if isBoundary(prev) {
				score += 14
			}
			// CamelCase bump.
			if unicode.IsLower(prev) && unicode.IsUpper(cur) {
				score += 8
			}
		}
		// Prefer earlier matches.
		score -= p / 8
		// Slight preference for shorter queries matching early.
		if i == 0 {
			score -= p / 4
		}
		last = p
	}

	// Penalize gaps.
	for i := 1; i < len(pos); i++ {
		gap := pos[i] - pos[i-1] - 1
		if gap > 0 {
			score -= gap
		}
	}

	return Match{Score: score, Positions: pos}, true
}

func isBoundary(r rune) bool {
	return r == ' ' || r == '_' || r == '-' || r == '/' || r == '.' || r == ':' || r == '#' || r == '\t'
}
