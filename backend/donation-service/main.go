// Command donation-service records pledges (promises to give) for Priyosha
// Charity Organization.
//
// Endpoints:
//
//	GET  /health  - liveness check
//	GET  /pledges - list recorded pledges (newest first) — ADMIN ONLY, see below
//	POST /pledges - record a pledge: {"amount": 50, "frequency": "one-time"|"monthly", "name": "...", "email": "..."}
//
// Admin access: GET /pledges requires the X-Admin-Token header to match the
// ADMIN_TOKEN env var. If ADMIN_TOKEN is not set, GET /pledges always
// returns 403 (fail-closed — donor names and emails must never be public).
// POST /pledges is public, as visitors must be able to pledge.
//
// Storage: in-memory, persisted to a JSON file (DATA_FILE env var, default
// ./data/pledges.json) so pledges survive a restart. This is deliberately
// simple — replace with a real database before handling real money.
//
// IMPORTANT — payments are NOT wired up: a pledge here is a record of intent,
// not a charge. To take real payments, plug a payment provider (e.g. Stripe)
// in here — see the TODO in handleCreatePledge — and never store card numbers
// in this service.
//
// Only the Go standard library is used.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// maxRequestBody is the largest POST body accepted (64 KB).
const maxRequestBody = 64 * 1024

type Pledge struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Frequency string    `json:"frequency"` // "one-time" or "monthly"
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type store struct {
	mu      sync.Mutex
	pledges []Pledge
	file    string
}

func (s *store) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return // no file yet — start empty
	}
	var pledges []Pledge
	if err := json.Unmarshal(data, &pledges); err == nil {
		s.pledges = pledges
	}
}

func (s *store) save() {
	data, err := json.MarshalIndent(s.pledges, "", "  ")
	if err != nil {
		log.Printf("donation-service: could not encode pledges: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.file), 0o755); err != nil {
		log.Printf("donation-service: could not create data dir: %v", err)
		return
	}
	if err := os.WriteFile(s.file, data, 0o644); err != nil {
		log.Printf("donation-service: could not save pledges: %v", err)
	}
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

// validEmail checks a plain email address: at most 254 characters, a
// parseable address, and nothing else (no display name, no angle brackets).
func validEmail(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 254 || strings.ContainsAny(s, "<> \t\r\n") {
		return false
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return false
	}
	return addr.Address == s
}

// validText checks a free-text field: trimmed length within 1..max
// characters, with no angle brackets (so no HTML/script can be stored).
func validText(s string, max int) bool {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > max {
		return false
	}
	return !strings.ContainsAny(s, "<>")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// requireAdmin checks the X-Admin-Token header against ADMIN_TOKEN.
// Fail-closed: if ADMIN_TOKEN is not set, every admin request gets 403.
// Returns true only when the caller is authorised.
func requireAdmin(w http.ResponseWriter, r *http.Request, adminToken string) bool {
	if adminToken == "" {
		http.Error(w, "admin access not configured", http.StatusForbidden)
		return false
	}
	given := r.Header.Get("X-Admin-Token")
	if len(given) != len(adminToken) || subtle.ConstantTimeCompare([]byte(given), []byte(adminToken)) != 1 {
		http.Error(w, "forbidden: valid X-Admin-Token header required", http.StatusForbidden)
		return false
	}
	return true
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "./data/pledges.json"
	}
	// ADMIN_TOKEN is required to read pledges (GET /pledges). It is
	// intentionally allowed to be empty at startup so the service still
	// boots for local dev — but while it is empty, GET /pledges returns
	// 403 for every request (see requireAdmin). Set a long random value
	// in production and share it only with the site owner.
	adminToken := os.Getenv("ADMIN_TOKEN")

	st := &store{file: dataFile}
	st.load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "donation-service"})
	})

	mux.HandleFunc("/pledges", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if !requireAdmin(w, r, adminToken) {
				return
			}
			st.mu.Lock()
			defer st.mu.Unlock()
			out := make([]Pledge, len(st.pledges))
			copy(out, st.pledges)
			sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
			writeJSON(w, http.StatusOK, map[string]any{"pledges": out})

		case http.MethodPost:
			if r.ContentLength > maxRequestBody {
				http.Error(w, "request body too large (max 64 KB)", http.StatusRequestEntityTooLarge)
				return
			}
			var in struct {
				Amount    float64 `json:"amount"`
				Frequency string  `json:"frequency"`
				Name      string  `json:"name"`
				Email     string  `json:"email"`
			}
			dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody))
			if err := dec.Decode(&in); err != nil {
				if strings.Contains(err.Error(), "too large") {
					http.Error(w, "request body too large (max 64 KB)", http.StatusRequestEntityTooLarge)
				} else {
					http.Error(w, "invalid JSON body", http.StatusBadRequest)
				}
				return
			}
			if math.IsNaN(in.Amount) || math.IsInf(in.Amount, 0) || in.Amount < 1 || in.Amount > 100000 {
				http.Error(w, "amount must be between 1 and 100000", http.StatusBadRequest)
				return
			}
			if in.Frequency != "one-time" && in.Frequency != "monthly" {
				http.Error(w, "frequency must be 'one-time' or 'monthly'", http.StatusBadRequest)
				return
			}
			if !validText(in.Name, 100) {
				http.Error(w, "name must be 1–100 characters and must not contain < or >", http.StatusBadRequest)
				return
			}
			if !validEmail(in.Email) {
				http.Error(w, "email must be a valid email address (max 254 characters)", http.StatusBadRequest)
				return
			}

			// TODO (real payments): this only records the pledge. To charge a
			// card, create a Stripe Checkout Session / PaymentIntent here
			// (e.g. github.com/stripe/stripe-go) using the amount and the
			// donor's email, store the resulting payment ID on the pledge,
			// and return the checkout URL to the frontend instead of
			// confirming a payment. No card details should ever reach this
			// service — card entry must happen on the provider's hosted page.
			p := Pledge{
				ID:        time.Now().Format("20060102150405.000000000"),
				Amount:    in.Amount,
				Frequency: in.Frequency,
				Name:      strings.TrimSpace(in.Name),
				Email:     strings.TrimSpace(in.Email),
				CreatedAt: time.Now().UTC(),
			}
			st.mu.Lock()
			st.pledges = append(st.pledges, p)
			st.save()
			st.mu.Unlock()

			writeJSON(w, http.StatusCreated, map[string]any{
				"pledge":  p,
				"message": "Thank you — your pledge has been recorded. This is a pledge of intent; no payment has been taken yet.",
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("donation-service listening on :%s (data: %s, admin GET protection: %v)", port, dataFile, adminToken != "")
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
