package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

const sliceLimit = 10

func Top10(s string) []string {
	str := getByFrequency(strings.Fields(s))

	return getSubSlice(str, sliceLimit)
}

func getByFrequency(s []string) []string {
	frequency := make(map[string]int)
	for _, c := range s {
		frequency[c]++
	}

	sl := make([]string, 0, len(frequency))
	for fr := range frequency {
		sl = append(sl, fr)
	}

	sort.Slice(sl, func(i, j int) bool {
		if frequency[sl[i]] == frequency[sl[j]] {
			return lexicographicallySort(sl[i], sl[j])
		}
		return frequency[sl[i]] > frequency[sl[j]]
	})

	return sl
}

func lexicographicallySort(s1 string, s2 string) bool {
	return sort.IsSorted(sort.StringSlice([]string{s1, s2}))
}

func getSubSlice(s []string, limit int) []string {
	if len(s) >= limit {
		return s[0:limit]
	}

	return s
}
