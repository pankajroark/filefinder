package main

import (
	"flag"
	"fmt"
	"github.com/pankajroark/pathsearch/lib"
	"log"
	"net/http"
	"time"
)

import _ "net/http/pprof"

// todo instead of polling, listen to filesystem events using watchman or sth
const indexEverySeconds = 600

func scheduleIndex(s *lib.Server) {
	ticker := time.NewTicker(indexEverySeconds * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Perform scheduled indexing...")
				s.Index()
			}
		}
	}()
}

func main() {
	// todo - referesh index periodically
	serv := lib.Server{}
	serv.Init()

	port := flag.String("port", "10121", "port on which to run the wiki")
	flag.Parse()
	app := "pathsearch"
	fmt.Printf("starting up %s on port %s ...\n", app, *port)
	http.HandleFunc("/query", lib.CreateQueryHandler(&serv))
	http.HandleFunc("/index", lib.CreateIndexHander(&serv))
	http.HandleFunc("/addroot", lib.CreateAddRootHandler(&serv))
	scheduleIndex(&serv)
	/*
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	*/
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
