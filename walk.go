package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WeighedDistance(s, t string) int {
	substitutionCost := 21
	deletionCost := 20
	insertionCost := 1
	// degenerate cases
	if s == t {
		return 0
	}

	if len(s) == 0 {
		return len(t) * insertionCost
	}
	if len(t) == 0 {
		return len(s) * deletionCost
	}

	// create two work vectors of integer distances
	v0 := make([]int, len(s)+1)
	v1 := make([]int, len(s)+1)

	// initialize v0 (the previous row of distances)
	// this row is A[0][i]: edit distance for an empty s
	// the distance is just the number of characters to delete from t
	for i := 0; i < len(v0); i++ {
		v0[i] = i * deletionCost
	}

	for i := 0; i < len(t); i++ {
		// calculate v1 (current row distances) from the previous row v0

		// first element of v1 is A[i+1][0]
		//   edit distance is delete (i+1) chars from s to match empty t
		v1[0] = (i + 1) * deletionCost

		// use formula to fill in the rest of the row
		for j := 0; j < len(s); j++ {
			cost := 0
			if t[i] != s[j] {
				cost = substitutionCost
			}
			v1[j+1] = min(v1[j]+deletionCost, v0[j+1]+insertionCost, v0[j]+cost)
		}

		// copy v1 (current row) to v0 (previous row) for next iteration
		for j := 0; j < len(v0); j++ {
			v0[j] = v1[j]
		}
	}

	return v1[len(s)]
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

func index(path string) map[string][]string {
	idx := make(map[string][]string)
	index_path := func(path string) {
		base := filepath.Base(path)
		for i := 0; i < len(base)-2; i++ {
			trigram := strings.ToLower(base[i : i+3])
			idx[trigram] = append(idx[trigram], path)
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			index_path(path)
		}
		return err
	}

	filepath.Walk(path, walkFn)
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
		if count > 2 {
			cands = append(cands, cand)
		}
	}
	return cands
}

func score(cand, fuzz string) int {
	lscore := WeighedDistance(strings.ToLower(fuzz), strings.ToLower(cand))
	//fmt.Printf("candidate: %s, word: %s, score: %d\n", cand, fuzz, lscore)
	return lscore
}

func rank(cands []string, fuzz string) []string {
	// assign a score to each candidate
	// sort by them
	candscores := make([]candscor, 0)
	for _, cand := range cands {
		basecand := filepath.Base(cand)
		score := score(basecand, fuzz)
		cs := candscor{cand: cand, score: score}
		//fmt.Println(cs)
		candscores = append(candscores, cs)
	}
	sort.Sort(ByScore(candscores))
	ret := make([]string, 0)
	for _, cs := range candscores {
		fmt.Println(cs)
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

type server struct {
	idx map[string][]string
}

func (s *server) initIndex() {
	s.idx = index("/Users/pankajg/workspace/source/science/src")
}

func (s *server) findMatches(word string) []string {
	candidates := findCandidates(word, s.idx)
	return uptoN(rank(candidates, word), 10)
}

func createQueryHandler(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		word := r.URL.Path[1:]

		//fmt.Println(candidates)
		for _, result := range s.findMatches(word) {
			//_ = result
			fmt.Fprintln(w, string(result))
		}
	}
}

func main() {
	serv := server{}
	serv.initIndex()

	port := flag.String("port", "10120", "port on which to run the wiki")
	flag.Parse()
	app := "pathfinder"
	fmt.Printf("starting up %s on port %s ...", app, *port)
	http.HandleFunc("/", createQueryHandler(&serv))
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
