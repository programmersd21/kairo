package search

import (
	"sort"
	"strings"
)

type Kind string

const (
	KindTask    Kind = "task"
	KindCommand Kind = "command"
	KindTag     Kind = "tag"
)

type Item struct {
	ID    string
	Kind  Kind
	Title string
	Desc  string
	Hint  string
}

type Result struct {
	Item  Item
	Match Match
}

type Index struct {
	items []Item
	keys  []string // normalized search key per item
}

func NewIndex(items []Item) *Index {
	idx := &Index{}
	idx.Replace(items)
	return idx
}

func (i *Index) Replace(items []Item) {
	i.items = append([]Item(nil), items...)
	i.keys = make([]string, len(items))
	for n, it := range i.items {
		key := strings.TrimSpace(it.Title + " " + it.Hint + " " + it.Desc)
		i.keys[n] = strings.ToLower(key)
	}
}

func (i *Index) Search(query string, limit int) []Result {
	if limit <= 0 {
		limit = 20
	}
	q := strings.TrimSpace(query)
	if q == "" {
		out := make([]Result, 0, min(limit, len(i.items)))
		for n := 0; n < len(i.items) && len(out) < limit; n++ {
			out = append(out, Result{Item: i.items[n], Match: Match{Score: 0}})
		}
		return out
	}

	res := make([]Result, 0, limit)
	for n := range i.items {
		m, ok := FuzzyMatch(q, i.keys[n])
		if !ok {
			continue
		}
		res = append(res, Result{Item: i.items[n], Match: m})
	}
	sort.SliceStable(res, func(a, b int) bool {
		if res[a].Match.Score != res[b].Match.Score {
			return res[a].Match.Score > res[b].Match.Score
		}
		// Secondary: prefer commands over tasks over tags in ties (palette feel).
		ka, kb := res[a].Item.Kind, res[b].Item.Kind
		if ka != kb {
			return kindRank(ka) < kindRank(kb)
		}
		return res[a].Item.Title < res[b].Item.Title
	})
	if len(res) > limit {
		res = res[:limit]
	}
	return res
}

func kindRank(k Kind) int {
	switch k {
	case KindCommand:
		return 0
	case KindTask:
		return 1
	case KindTag:
		return 2
	default:
		return 9
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
