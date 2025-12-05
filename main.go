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

	assetsDir := filepath.Join(baseDir, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve the marketing landing at root.
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(baseDir, "landing.html"))
	})

	http.HandleFunc("/landing", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(baseDir, "landing.html"))
	})

	http.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		// Serve the clinical assistant UI at /app.
		http.ServeFile(w, r, filepath.Join(baseDir, "index (3).html"))
	})

	http.HandleFunc("/api/audit", func(w http.ResponseWriter, r *http.Request) {
		addCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis.LatestAudits(10))
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
		if len(resp.ValidationErrors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "validation_failed",
				"details": resp.ValidationErrors,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		// Minimal audit logging (redacted name).
		ref := req.PatientName
		if len(ref) > 2 {
			ref = ref[:1] + "***"
		}
		log.Printf("analysis audit_id=%s patient=%s complaint=%s risk=%s score=%d", resp.AuditID, ref, req.Complaint, resp.RiskLevel, resp.RiskScore)
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
