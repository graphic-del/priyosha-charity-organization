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
// Security (see SECURITY.md): every response carries security headers
// (HSTS, nosniff, DENY framing, strict referrer policy, Permissions-Policy
// and a conservative API Content-Security-Policy). Requests are rate
// limited per IP — POST /api/pledges and POST /api/messages: 10/minute;
// GET /api/*: 120/minute (429 with Retry-After when exceeded). Request
// bodies forwarded to the services are capped at 64 KB.
//
// Service addresses come from env vars (defaults are the docker-compose
// service names, which also work locally if you add them to /etc/hosts —
// for local runs without Docker, set them to http://localhost:8081 etc).
//
// Only the Go standard library is used.
package main

import (
	"crypto/subtle"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// maxRequestBody is the largest request body the gateway will forward (64 KB).
const maxRequestBody = 64 * 1024

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// securityHeaders sets the security headers on every response, including
// errors. HSTS only takes effect over HTTPS, which Caddy provides in
// production; setting it here as well is harmless on plain HTTP.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		// JSON API: nothing may be loaded, framed or submitted from here.
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
		next.ServeHTTP(w, r)
	})
}

// clientIP returns the caller IP. Behind Caddy it is the first entry of
// X-Forwarded-For; otherwise it is the TCP peer. X-Forwarded-For can be
// spoofed by a direct caller, so it is used only for rate limiting, never
// for authentication — admin access always needs X-Admin-Token.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// rateLimiter is a small in-memory, per-IP, fixed-window limiter.
// It is deliberately simple: it protects a small charity site from spam
// and brute-force floods. It is not a distributed limiter — each gateway
// replica has its own counters, which is fine at this scale.
type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
}

type rateBucket struct {
	count  int
	window time.Time
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{buckets: make(map[string]*rateBucket)}
}

// allow reports whether key may proceed under limit requests per minute,
// and how many seconds to wait if not.
func (l *rateLimiter) allow(key string, limit int) (bool, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok || now.Sub(b.window) >= time.Minute {
		l.buckets[key] = &rateBucket{count: 1, window: now}
		return true, 0
	}
	if b.count >= limit {
		retry := int((time.Minute - now.Sub(b.window)).Seconds()) + 1
		if retry < 1 {
			retry = 1
		}
		return false, retry
	}
	b.count++
	// Prune stale buckets so the map cannot grow without bound.
	if len(l.buckets) > 5000 {
		for k, v := range l.buckets {
			if now.Sub(v.window) >= time.Minute {
				delete(l.buckets, k)
			}
		}
	}
	return true, 0
}

// limit enforces the documented limits: POST /api/pledges and
// POST /api/messages 10/minute per IP; GET /api/* 120/minute per IP.
// OPTIONS (CORS preflight) and /health are not rate limited.
func (l *rateLimiter) limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions || !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		var key string
		var limit int
		switch {
		case r.Method == http.MethodPost && (r.URL.Path == "/api/pledges" || r.URL.Path == "/api/messages"):
			key = "post:" + clientIP(r)
			limit = 10
		case r.Method == http.MethodGet:
			key = "get:" + clientIP(r)
			limit = 120
		default:
			next.ServeHTTP(w, r)
			return
		}
		if ok, retry := l.allow(key, limit); !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			http.Error(w, fmt.Sprintf("too many requests — please wait %d seconds and try again", retry), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// proxy forwards the request to target (e.g. http://donation-service:8081/pledges)
// and copies the response back, preserving status code and JSON body.
func proxy(target string) http.HandlerFunc {
	client := &http.Client{Timeout: 10 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > maxRequestBody {
			http.Error(w, "request body too large (max 64 KB)", http.StatusRequestEntityTooLarge)
			return
		}
		var body io.Reader
		if r.Body != nil {
			body = http.MaxBytesReader(w, r.Body, maxRequestBody)
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

	limiter := newRateLimiter()
	handler := securityHeaders(limiter.limit(withCORS(allowedOrigin, mux)))

	log.Printf("api-gateway listening on :%s (env=%s, allowed_origin=%s, admin GET protection: %v)", port, envMode, allowedOrigin, adminToken != "")
	log.Printf("  content=%s donation=%s contact=%s", content, donation, contact)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
