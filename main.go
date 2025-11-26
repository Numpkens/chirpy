package main

import (
	"log"
	"net/http"
	"time"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handlerReadiness)

	fsHandler := http.FileServer(http.Dir("."))

	strippedHandler := http.StripPrefix("/app", fsHandler)

	mux.Handle("/app/", strippedHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,

		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on %s", server.Addr)

	err := server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
}
