package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	responseBody := fmt.Sprintf("Hits: %d", hits)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBody))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits count reset to 0"))
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	// Update the routing to include the required HTTP methods

	// /healthz must only accept GET requests
	mux.HandleFunc("GET /healthz", handlerReadiness)

	// /metrics must only accept GET requests
	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)

	// /reset must only accept POST requests
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)

	// File server remains the same (handles GET requests implicitly)
	fsHandler := http.FileServer(http.Dir("."))
	strippedHandler := http.StripPrefix("/app", fsHandler)
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(strippedHandler))

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

