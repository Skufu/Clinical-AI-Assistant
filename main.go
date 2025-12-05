package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Skufu/Clinical-AI-Assistant/internal/analysis"
)

func main() {
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to resolve working directory: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the single-page HTML from the project root.
		http.ServeFile(w, r, filepath.Join(baseDir, "index (3).html"))
	})

	http.HandleFunc("/api/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			addCORS(w)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		addCORS(w)

		var req analysis.Intake
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		resp := analysis.Analyze(req)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	addr := ":8080"
	log.Printf("Clinical AI Assistant backend running on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func addCORS(w http.ResponseWriter) {
	// Allow same-origin plus simple dev usage.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
		"Content-Type",
	}, ", "))
}
