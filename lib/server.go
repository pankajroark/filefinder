package lib

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const IndexPath = "/Users/pankajg/.pathsearchindex"
const StringidsPath = "/Users/pankajg/.pathstringids"

type candscor struct {
	cand  string
	score int
}

type ByScore []candscor

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].score < a[j].score }

type Server struct {
	idx       map[string][]uint32
	roots     []string
	stringids *Stringids
}

func mergeIndices(idx1, idx2 map[string][]uint32) map[string][]uint32 {
	indices := make([]map[string][]uint32, 0)
	indices = append(indices, idx1)
	indices = append(indices, idx2)

	newIdx := make(map[string][]uint32)
	for _, idx := range indices {
		for trigram, paths := range idx {
			for _, path := range paths {
				newIdx[trigram] = append(newIdx[trigram], path)
			}
		}
	}
	return newIdx
}

func (s *Server) StoreIndex() {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	err := e.Encode(s.idx)
	if err != nil {
		log.Fatal("failed to encode index")
	}
	ioutil.WriteFile(IndexPath, b.Bytes(), 0644)
}

func (s *Server) ReadIndex() error {
	fmt.Println("Reading Index...")
	_, cerr := os.Stat(IndexPath)
	if cerr != nil {
		fmt.Println("Index does not exist.")
		return cerr
	}
	var decodedIdx map[string][]uint32
	bs, err := ioutil.ReadFile(IndexPath)
	if err != nil {
		log.Fatal("failed to decode index")
		return err
	}
	d := gob.NewDecoder(bytes.NewBuffer(bs))
	d.Decode(&decodedIdx)
	s.idx = decodedIdx
	return nil
}

func (s *Server) Init() {
	s.roots = make([]string, 0)
	s.roots = append(s.roots, "/Users/pankajg/workspace/source/science")
	s.roots = append(s.roots, "/Users/pankajg/workspace/source/birdcage")
	err := s.ReadIndex()
	s.stringids = NewStringids(StringidsPath)
	if err != nil {
		s.Index()
	}
}

func (s *Server) Index() {
	fmt.Printf("indexing %s\n", s.roots[0])
	s.idx = s.index(s.roots[0])
	for i := 1; i < len(s.roots); i++ {
		fmt.Printf("indexing %s\n", s.roots[i])
		newIdx := s.index(s.roots[i])
		s.idx = mergeIndices(s.idx, newIdx)
	}
	s.StoreIndex()
}

func (s *Server) index(path string) map[string][]uint32 {
	fmt.Printf("scanning %s\n", path)
	idx := make(map[string][]uint32)
	index_path := func(path string) {
		pathId := s.stringids.Add(path)
		base := filepath.Base(path)
		for i := 0; i < len(base)-2; i++ {
			trigram := strings.ToLower(base[i : i+3])
			idx[trigram] = append(idx[trigram], pathId)
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) != ".class" {
			index_path(path)
		}
		return err
	}

	filepath.Walk(path, walkFn)
	return idx
}

func (s *Server) FindMatches(word string) []string {
	candidates := s.findCandidates(word, s.idx)
	return uptoN(rank(candidates, word), 10)
}

func (s *Server) findCandidates(fuzz string, idx map[string][]uint32) []string {
	candsSeen := make(map[uint32]int)
	for i := 0; i < len(fuzz)-2; i++ {
		trigram := strings.ToLower(fuzz[i : i+3])
		pathIds := idx[trigram]
		for _, pathId := range pathIds {
			candsSeen[pathId]++
		}
	}

	// at least two trigrams should match
	cands := make([]string, 0)
	for cand, count := range candsSeen {
		if count > 2 {
			cands = append(cands, s.stringids.StrAtOffset(cand))
		}
	}
	return cands
}

func score(cand, fuzz string) int {
	lscore := WeightedDistance(strings.ToLower(fuzz), strings.ToLower(cand))
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
		//fmt.Println(cs)
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

func CreateQueryHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		word := r.URL.Query().Get("word")

		//fmt.Println(candidates)
		for _, result := range s.FindMatches(word) {
			//_ = result
			fmt.Fprintln(w, string(result))
		}
	}
}

func CreateIndexHander(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Index()
		fmt.Fprintln(w, "Indexing done!")
	}
}

func CreateAddRootHandler(s *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := r.URL.Query().Get("root")
		// todo validate root, path must exist on disk
		s.roots = append(s.roots, root)
		s.Index()
		fmt.Fprintf(w, "Added root %s\n", root)
	}
}
