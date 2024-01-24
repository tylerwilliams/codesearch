package result

import (
	"fmt"
)

type Result struct {
	Project string
	Count    int
	Filename string
	Snippets [][]byte
}

func (r Result) String() string {
	out := fmt.Sprintf("%s [%d matches]\n", r.Filename, r.Count)
	for _, snip := range r.Snippets {
		out += fmt.Sprintf("  %s", string(snip))
	}
	return out
}
