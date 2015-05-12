package main

import (
	"flag"
	"fmt"
	"github.com/pankajroark/pathsearch/lib"
	"log"
	"net/http"
)

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
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
