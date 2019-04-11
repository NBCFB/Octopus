package main

import (
	"flag"
	"fmt"
	"github.com/NBCFB/Octopus"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func ping (w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong\n"))
}

func main() {
	var host string
	var port int

	// Handle command line flag
	flag.StringVar(&host, "h", "localhost", "specify host")
	flag.IntVar(&port, "p", 8080, "specify port")
	flag.Parse()

	// Make up addr with default settings
	addr := fmt.Sprintf("%s:%d", host, port)

	// Set up router
	r := chi.NewRouter()
	r.Get("/ping", ping)

	s := &http.Server{
		Addr: addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Start server
	Octopus.GracefulServe(s, false)
}