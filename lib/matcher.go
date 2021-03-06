package lib

import (
	"path/filepath"
	"sort"
	"strings"
)

type candscor struct {
	cand  string
	score int
}

type ByScore []candscor

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }

// todo return a clone
func reverse(ss []string) {
	l := len(ss)
	for i, j := 0, l-1; i < j; i, j = i+1, j-1 {
		ss[i], ss[j] = ss[j], ss[i]
	}
}

// Query will have the reverse form of the bath
// i.e. filequery/dir
func match(cands []string, query string) []string {
	qparts := strings.Split(query, "/")
	//qfilepart := qparts[len(qparts)-1]
	qfilepart := qparts[0]
	baseExtractor := func(path string) string {
		return filepath.Base(path)
	}
	matches := rank(cands, qfilepart, baseExtractor)
	// Find somewhat big number of matches based on filepart match
	top := uptoN(matches, 100)
	if len(qparts) > 1 {
		identityExtractor := func(path string) string {
			return path
		}
		// note that qparts is modified here
		reverse(qparts)
		revQuery := strings.Join(qparts, "/")
		top = rank(top, revQuery, identityExtractor)
	}
	// Pick smaller number from the large set based on full match
	return uptoN(top, 10)
}

func uptoN(slice []string, n int) []string {
	if len(slice) > n {
		return slice[:n]
	} else {
		return slice
	}
}

func rank(cands []string, fuzz string, candExtractor func(string) string) []string {
	// assign a score to each candidate
	// sort by them
	candscores := make([]candscor, 0)
	for _, cand := range cands {
		basecand := candExtractor(cand)
		score := score(basecand, fuzz)
		cs := candscor{cand: cand, score: score}
		//fmt.Println(cs)
		candscores = append(candscores, cs)
	}
	sort.Sort(ByScore(candscores))
	ret := make([]string, 0)
	for _, cs := range candscores {
		//fmt.Println(cs)
		ret = append(ret, cs.cand)
	}
	return ret
}

func rankByPath(cands []string, pathpart string) {
}
