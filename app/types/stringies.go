package types

import (
	"sort"
	"strings"
)

type Stringies []string

func (r Stringies) Delete(payload string) Stringies {
	if len(r) == 0 {
		return nil
	}

	out := make(Stringies, 0, len(r))

	for _, str := range r {
		if str != payload {
			out = append(out, str)
		}
	}

	return out
}

func (r Stringies) Join(s string) string {
	return strings.Join(r, s)
}

func (r Stringies) Sort() Stringies {
	if len(r) == 0 {
		return nil
	}

	cp := make([]string, len(r))
	copy(cp, r)

	sort.Slice(cp, func(i, j int) bool { return cp[i] < cp[j] })

	return cp
}
