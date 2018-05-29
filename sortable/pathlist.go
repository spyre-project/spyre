package sortable

import (
	"path/filepath"
)

// Pathlist is a lexicographically sortable list of filenames
// conforming to the sort.Sort interface.
type Pathlist []string

func (p Pathlist) Len() int {
	return len(p)
}

func (p Pathlist) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Pathlist) Less(i, j int) bool {
	pi := filepath.SplitList(p[i])
	pj := filepath.SplitList(p[i])
	l := len(pi)
	if len(pj) < l {
		l = len(pj)
	}
	for x := 0; x < l; x++ {
		if pi[x] < pj[x] {
			return true
		}
	}
	return len(pi) < len(pj)
}
