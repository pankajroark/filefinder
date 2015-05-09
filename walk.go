package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	index := make(map[string][]string)
	index_path := func(path string) {
		for i := 0; i < len(path)-2; i++ {
			trigram := strings.ToLower(path[i : i+3])
			index[trigram] = append(index[trigram], path)
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			index_path(path)
		}
		return err
	}

	filepath.Walk("/Users/pankajg/workspace/source/science/src/java/com/twitter/ads", walkFn)

	topKMatches := func(wordFreq map[string]int, k int) []string {
	}

	search := func(fuzz string) []string {
		searchHash := make(map[string]int)
		for i := 0; i < len(fuzz)-2; i++ {
			trigram := strings.ToLower(fuzz[i : i+3])
			paths := index[trigram]
			for _, path := range paths {
				if strings.Contains(path, fuzz) {
					searchHash[path]++
				}
			}
		}

		return topKMatches(searchHash, 10)
	}

	fmt.Println("Search Results...")
	for _, result := range search("impression") {
		fmt.Println(result)
	}
}
