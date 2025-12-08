package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic" // NEW: Required for atomic.Int32
	"time"
)

// Step 1: Create a struct that holds the stateful data.
// atomic.Int32 is used for safe, concurrent modification of the counter.
type apiConfig struct {
	fileserverHits atomic.Int32
}

// Handler for the /healthz endpoint (from previous assignment)
func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Step 3: Write a new middleware method on *apiConfig.
// This function returns a new http.Handler that wraps the 'next' handler.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// The closure (the function returned by the middleware) is the actual handler.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the counter safely using .Add(1)
		cfg.fileserverHits.Add(1)
		// Call the next handler in the chain (e.g., the file server)
		next.ServeHTTP(w, r)
	})
}

// Step 4: Create a new handler method on *apiConfig for the /metrics path.
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	// Read the current count safely using .Load()
	hits := cfg.fileserverHits.Load()

	// Format the output string
	responseBody := fmt.Sprintf("Hits: %d", hits)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBody))
}

// Step 5: Create a new handler method on *apiConfig for the /reset path.
func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	// Reset the counter to 0 safely using .Store(0)
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits count reset to 0"))
}

func main() {
	// Step 2: Initialize the apiConfig struct instance
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()

	// Register existing handlers
	mux.HandleFunc("/healthz", handlerReadiness)

	// Update the fileserver handler with middleware
	// The strippedHandler is now wrapped by middlewareMetricsInc.
	fsHandler := http.FileServer(http.Dir("."))
	strippedHandler := http.StripPrefix("/app", fsHandler)

	// Wrap the strippedHandler with the metric incrementer middleware
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(strippedHandler))

	// Step 6: Register the new stateful handlers
	mux.HandleFunc("/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/reset", apiCfg.handlerReset)

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
