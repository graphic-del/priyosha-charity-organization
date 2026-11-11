// Command content-service serves the public facts and programme data for
// Priyosha Charity Organization as JSON.
//
// Endpoints:
//
//	GET /health    - liveness check
//	GET /stats     - the need in Bangladesh (figures as provided by the founder)
//	GET /programs  - the charity's programmes
//
// Only the Go standard library is used. No database is required: the data
// is compiled in, so the service is stateless and safe to run as replicas.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// Stat is one row of "The Need". Figures are reproduced exactly as provided
// by the founder; see the README for the sources.
type Stat struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Figure string `json:"figure"`
	Detail string `json:"detail"`
	Source string `json:"source"`
}

type Program struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Focus       string `json:"focus"`
}

var stats = []Stat{
	{
		ID:     "food-insecurity",
		Title:  "People lacking enough food",
		Figure: "18.1 million — 1 crore 81 lakh",
		Detail: "Projected to face crisis-level food insecurity or worse during September–December 2026, within the areas assessed.",
		Source: "Integrated Food Security Phase Classification (IPC)",
	},
	{
		ID:     "sleeping-in-public",
		Title:  "People sleeping in public places",
		Figure: "22,185",
		Detail: "Counted in the 2022 census. This narrow category covers people sleeping on streets, at stations and similar places. It does not capture everyone without secure housing, and it is an older count.",
		Source: "Bangladesh Bureau of Statistics, 2022 census (nsds.bbs.gov.bd)",
	},
	{
		ID:     "dogs",
		Title:  "Dogs",
		Figure: "About 16.7 lakh dogs",
		Detail: "Estimated in a 2024 study. This estimates the dog population; it does not establish how many are homeless or hungry.",
		Source: "2024 study (PMC)",
	},
	{
		ID:     "animals-overall",
		Title:  "Hungry or homeless animals overall, including cats",
		Figure: "No reliable nationwide total found",
		Detail: "Research identifies gaps in Bangladesh's stray-animal population data. Giving an exact suffering count would be guesswork.",
		Source: "pmc.ncbi.nlm.nih.gov",
	},
}

// warning is shown with the stats wherever they are displayed.
const warning = "These figures cannot be added together: hunger and homelessness can affect the same people, and animal population estimates do not measure suffering."

var programs = []Program{
	{
		ID:          "food-for-people",
		Title:       "Food for people",
		Description: "Food support for people facing crisis-level food insecurity in Bangladesh — meals and food packs for families who do not have enough to eat.",
		Focus:       "People",
	},
	{
		ID:          "support-sleeping-rough",
		Title:       "Support for people sleeping rough",
		Description: "Help for people sleeping in public places — on streets, at stations and similar places — including shelter support, essentials and a path to more secure housing.",
		Focus:       "People",
	},
	{
		ID:          "care-for-animals",
		Title:       "Food and care for cats and dogs",
		Description: "Food, shelter and care for stray cats and dogs who suffer the most — feeding, basic veterinary care and rescue where we can reach them.",
		Focus:       "Animals — cats and dogs",
	},
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// securityHeaders sets the security headers on every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		next.ServeHTTP(w, r)
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "content-service"})
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"warning": warning, "stats": stats})
	})

	mux.HandleFunc("/programs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"programs": programs})
	})

	log.Printf("content-service listening on :%s", port)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
