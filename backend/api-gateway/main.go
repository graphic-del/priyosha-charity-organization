// Command api-gateway is the single entry point for the Priyosha Charity
// Organization backend (port :8080). It routes requests to the other Go
// microservices and adds CORS so the Next.js frontend can call it from the
// browser.
//
// Routes (all under /api, except /health):
//
//	GET  /health        - gateway liveness check
//	GET  /api/stats     -> content-service  /stats
//	GET  /api/programs  -> content-service  /programs
//	GET  /api/pledges   -> donation-service /pledges (ADMIN ONLY — X-Admin-Token, 403 without it)
//	POST /api/pledges   -> donation-service /pledges (public)
//	GET  /api/messages  -> contact-service  /messages (ADMIN ONLY — X-Admin-Token, 403 without it)
//	POST /api/messages  -> contact-service  /messages (public)
//
// CORS: ALLOWED_ORIGIN defaults to http://localhost:3000 for local dev.
// In production (ENV=production or NODE_ENV=production) the gateway
// refuses to start if ALLOWED_ORIGIN is empty or "*", so a misconfigured
// deploy can never accidentally allow every origin.
// Admin lists are enforced here AND in the downstream services, using the
// same ADMIN_TOKEN env var; the gateway forwards the X-Admin-Token header.
//
// Service addresses come from env vars (defaults are the docker-compose
// service names, which also work locally if you add them to /etc/hosts —
// for local runs without Docker, set them to http://localhost:8081 etc).
//
// Only the Go standard library is used.
package main

import (
	"crypto/subtle"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// proxy forwards the request to target (e.g. http://donation-service:8081/pledges)
// and copies the response back, preserving status code and JSON body.
func proxy(target string) http.HandlerFunc {
	client := &http.Client{Timeout: 10 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		var body io.Reader
		if r.Body != nil {
			body = r.Body
		}
		req, err := http.NewRequest(r.Method, target, body)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		// Forward the admin header so downstream services can enforce
		// their own check too. Never log its value.
		if tok := r.Header.Get("X-Admin-Token"); tok != "" {
			req.Header.Set("X-Admin-Token", tok)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("gateway: upstream %s unreachable: %v", target, err)
			http.Error(w, "upstream service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

// withCORS allows exactly one frontend origin to call the API from the
// browser. The origin is decided in main(): http://localhost:3000 for
// local dev, and the explicit ALLOWED_ORIGIN in production.
func withCORS(allowed string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Admin-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin enforces the admin token for list endpoints. Fail-closed:
// if ADMIN_TOKEN is not set, every admin request gets 403.
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

// adminOnlyGet wraps a proxied route so GET requests need the admin token
// (POSTs pass through unchanged — visitors must be able to pledge/write in).
func adminOnlyGet(adminToken string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && !requireAdmin(w, r, adminToken) {
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, v.(string))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Environment: ENV takes precedence, NODE_ENV is accepted as an alias.
	// Any value other than "production" is treated as development.
	envMode := env("ENV", env("NODE_ENV", "development"))
	isProduction := envMode == "production"

	// CORS origin. Local dev defaults to the Next.js dev server; production
	// MUST set ALLOWED_ORIGIN explicitly to the exact frontend URL
	// (e.g. https://priyoshacharity.com). Fail fast instead of
	// silently allowing every origin.
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if isProduction && (allowedOrigin == "" || allowedOrigin == "*") {
		log.Fatal("api-gateway: refusing to start in production with ALLOWED_ORIGIN unset or \"*\" — set ALLOWED_ORIGIN to your exact frontend URL, e.g. https://priyoshacharity.com")
	}
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}
	if allowedOrigin == "*" {
		log.Print("api-gateway: WARNING — ALLOWED_ORIGIN is \"*\" (any website may call this API). This is only acceptable for local development.")
	}
	// Same ADMIN_TOKEN must be set on the gateway and on the
	// donation/contact services (docker-compose.yml passes it to all three).
	adminToken := os.Getenv("ADMIN_TOKEN")

	content := strings.TrimSuffix(env("CONTENT_SERVICE_URL", "http://content-service:8083"), "/")
	donation := strings.TrimSuffix(env("DONATION_SERVICE_URL", "http://donation-service:8081"), "/")
	contact := strings.TrimSuffix(env("CONTACT_SERVICE_URL", "http://contact-service:8082"), "/")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, `{"status":"ok","service":"api-gateway"}`)
	})
	mux.HandleFunc("/api/stats", proxy(content+"/stats"))
	mux.HandleFunc("/api/programs", proxy(content+"/programs"))
	mux.HandleFunc("/api/pledges", adminOnlyGet(adminToken, proxy(donation+"/pledges")))
	mux.HandleFunc("/api/messages", adminOnlyGet(adminToken, proxy(contact+"/messages")))

	log.Printf("api-gateway listening on :%s (env=%s, allowed_origin=%s, admin GET protection: %v)", port, envMode, allowedOrigin, adminToken != "")
	log.Printf("  content=%s donation=%s contact=%s", content, donation, contact)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(allowedOrigin, mux)))
}
