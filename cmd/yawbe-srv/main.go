package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/virtualtam/yawbe/pkg/http/www"
)

const (
	defaultListenAddr = ":8080"
)

func main() {
	listenAddr := flag.String("listenAddr", defaultListenAddr, "Listen on this address")
	flag.Parse()

	router := mux.NewRouter()

	www.AddRoutes(router)

	srv := &http.Server{
		Addr:         *listenAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
