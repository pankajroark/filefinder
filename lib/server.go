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
	"strings"
)

const IndexPath = "/Users/pankajg/.pathsearchindex"
const StringidsPath = "/Users/pankajg/.pathstringids"

type Server struct {
	idx       map[string][]uint32
	roots     []string
	stringids *Stringids
}

func (s *Server) Roots() []string {
	return s.roots
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
	s.stringids = NewStringids(StringidsPath)
	err := s.ReadIndex()
	if err != nil {
		s.Index()
	}
}

func (s *Server) Index() {
	fmt.Printf("indexing %s\n", s.roots[0])
	s.idx = MergeIndices(s.idx, s.index(s.roots[0]))
	for i := 1; i < len(s.roots); i++ {
		fmt.Printf("indexing %s\n", s.roots[i])
		newIdx := s.index(s.roots[i])
		s.idx = MergeIndices(s.idx, newIdx)
	}
	fmt.Printf("Total number of trigrams: %d", len(s.idx))
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
	return match(candidates, word)
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
			pathstr, _ := s.stringids.StrAtOffset(cand)
			cands = append(cands, pathstr)
		}
	}
	return cands
}

func score(cand, fuzz string) int {
	lscore := WeightedDistance(strings.ToLower(fuzz), strings.ToLower(cand))
	//fmt.Printf("candidate: %s, word: %s, score: %d\n", cand, fuzz, lscore)
	return lscore
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
