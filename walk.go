package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// levenshtein distance with no cost for insertion
func Distance(s1, s2 string) int {
	var cost, lastdiag, olddiag int
	len_s1 := len(s1)
	len_s2 := len(s2)

	column := make([]int, len_s1+1)

	for y := 1; y <= len_s1; y++ {
		column[y] = y
	}

	for x := 1; x <= len_s2; x++ {
		column[0] = x
		lastdiag = x - 1
		for y := 1; y <= len_s1; y++ {
			olddiag = column[y]
			cost = 0
			if s1[y-1] != s2[x-1] {
				cost = 1
			}
			column[y] = min(
				column[y]+1,
				column[y-1]+0, // no cost for insertion
				lastdiag+cost)
			lastdiag = olddiag
		}
	}
	return column[len_s1]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}

type candscor struct {
	cand  string
	score int
}

type ByScore []candscor

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }

func index() map[string][]string {
	idx := make(map[string][]string)
	index_path := func(path string) {
		base := filepath.Base(path)
		for i := 0; i < len(base)-2; i++ {
			trigram := strings.ToLower(base[i : i+3])
			idx[trigram] = append(idx[trigram], base)
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			index_path(path)
		}
		return err
	}

	filepath.Walk("/Users/pankajg/workspace/source/science/src/java/com/twitter", walkFn)
	return idx
}

func findCandidates(fuzz string, idx map[string][]string) []string {
	candsSeen := make(map[string]int)
	for i := 0; i < len(fuzz)-2; i++ {
		trigram := strings.ToLower(fuzz[i : i+3])
		paths := idx[trigram]
		for _, path := range paths {
			candsSeen[path]++
		}
	}

	// at least two trigrams should match
	cands := make([]string, 0)
	for cand, count := range candsSeen {
		if count > 3 {
			cands = append(cands, cand)
		}
	}
	return cands
}

func score(cand, fuzz string) int {
	lscore := Distance(strings.ToLower(cand), strings.ToLower(fuzz))
	return lscore
}

func rank(cands []string, fuzz string) []string {
	// assign a score to each candidate
	// sort by them
	candscores := make([]candscor, 0)
	for _, cand := range cands {
		score := score(cand, fuzz)
		cs := candscor{cand: cand, score: score}
		//fmt.Println(cs)
		candscores = append(candscores, cs)
	}
	sort.Sort(ByScore(candscores))
	ret := make([]string, 0)
	for _, cs := range candscores {
		ret = append(ret, cs.cand)
	}
	return ret
}

func uptoN(slice []string, n int) []string {
	if len(slice) > n {
		return slice[:n]
	} else {
		return slice
	}
}

func main() {
	idx := index()

	word := os.Args[1]
	candidates := findCandidates(word, idx)
	//fmt.Println(candidates)
	for _, result := range uptoN(rank(candidates, word), 10) {
		//_ = result
		fmt.Println(result)
	}

}
