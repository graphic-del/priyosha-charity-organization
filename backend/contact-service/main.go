// Command contact-service records messages from the website's contact /
// volunteer form for Priyosha Charity Organization.
//
// Endpoints:
//
//	GET  /health   - liveness check
//	GET  /messages - list recorded messages (newest first) — ADMIN ONLY, see below
//	POST /messages - record a message: {"name": "...", "email": "...", "interest": "donate"|"volunteer"|"other", "message": "..."}
//
// Admin access: GET /messages requires the X-Admin-Token header to match the
// ADMIN_TOKEN env var. If ADMIN_TOKEN is not set, GET /messages always
// returns 403 (fail-closed — sender names and emails must never be public).
// POST /messages is public, as visitors must be able to write in.
//
// Storage: in-memory, persisted to a JSON file (DATA_FILE env var, default
// ./data/messages.json) so messages survive a restart.
//
// IMPORTANT — no email is sent by this service: a message is only stored.
// To notify the team, plug in an email provider (e.g. SendGrid / SES) in
// handleCreateMessage — see the TODO there — or read messages with
// GET /messages.
//
// Only the Go standard library is used.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
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

type Message struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Interest  string    `json:"interest"` // "donate", "volunteer" or "other"
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type store struct {
	mu       sync.Mutex
	messages []Message
	file     string
}

func (s *store) load() {
	data, err := os.ReadFile(s.file)
	if err != nil {
		return // no file yet — start empty
	}
	var messages []Message
	if err := json.Unmarshal(data, &messages); err == nil {
		s.messages = messages
	}
}

func (s *store) save() {
	data, err := json.MarshalIndent(s.messages, "", "  ")
	if err != nil {
		log.Printf("contact-service: could not encode messages: %v", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(s.file), 0o755); err != nil {
		log.Printf("contact-service: could not create data dir: %v", err)
		return
	}
	if err := os.WriteFile(s.file, data, 0o644); err != nil {
		log.Printf("contact-service: could not save messages: %v", err)
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
		port = "8082"
	}
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "./data/messages.json"
	}
	// ADMIN_TOKEN is required to read messages (GET /messages). While it
	// is empty, GET /messages returns 403 for every request (see
	// requireAdmin). Set a long random value in production and share it
	// only with the site owner.
	adminToken := os.Getenv("ADMIN_TOKEN")

	st := &store{file: dataFile}
	st.load()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "contact-service"})
	})

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if !requireAdmin(w, r, adminToken) {
				return
			}
			st.mu.Lock()
			defer st.mu.Unlock()
			out := make([]Message, len(st.messages))
			copy(out, st.messages)
			sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
			writeJSON(w, http.StatusOK, map[string]any{"messages": out})

		case http.MethodPost:
			if r.ContentLength > maxRequestBody {
				http.Error(w, "request body too large (max 64 KB)", http.StatusRequestEntityTooLarge)
				return
			}
			var in struct {
				Name     string `json:"name"`
				Email    string `json:"email"`
				Interest string `json:"interest"`
				Message  string `json:"message"`
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
			if !validText(in.Name, 100) {
				http.Error(w, "name must be 1–100 characters and must not contain < or >", http.StatusBadRequest)
				return
			}
			if !validEmail(in.Email) {
				http.Error(w, "email must be a valid email address (max 254 characters)", http.StatusBadRequest)
				return
			}
			if in.Interest != "donate" && in.Interest != "volunteer" && in.Interest != "other" {
				http.Error(w, "interest must be 'donate', 'volunteer' or 'other'", http.StatusBadRequest)
				return
			}
			if !validText(in.Message, 2000) {
				http.Error(w, "message must be 1–2000 characters and must not contain < or >", http.StatusBadRequest)
				return
			}

			// TODO (real email): no email is sent yet. To notify the team,
			// send an email here via your provider (SendGrid, AWS SES, …)
			// to the address in CONTACT_NOTIFY_EMAIL, and set a reply-to of
			// the sender's email. Keep the stored copy as the record.
			m := Message{
				ID:        time.Now().Format("20060102150405.000000000"),
				Name:      strings.TrimSpace(in.Name),
				Email:     strings.TrimSpace(in.Email),
				Interest:  in.Interest,
				Message:   strings.TrimSpace(in.Message),
				CreatedAt: time.Now().UTC(),
			}
			st.mu.Lock()
			st.messages = append(st.messages, m)
			st.save()
			st.mu.Unlock()

			writeJSON(w, http.StatusCreated, map[string]any{
				"message": m,
				"note":    "Thank you — your message has been recorded. We will get back to you.",
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("contact-service listening on :%s (data: %s, admin GET protection: %v)", port, dataFile, adminToken != "")
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
