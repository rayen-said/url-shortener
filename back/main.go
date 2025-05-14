package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Store our URL mappings in memory (for simplicity)
// In a production app, use a database (e.g., Redis, PostgreSQL)
var (
	urlStore        = make(map[string]string)
	mu              sync.RWMutex // To safely access urlStore concurrently
	letterRunes     = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	shortCodeLength = 6
)

// Request structure for shortening a URL
type ShortenRequest struct {
	URL string `json:"url"`
}

// Response structure for a shortened URL
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// Initialize random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// generateShortCode creates a random string of a fixed length
func generateShortCode() string {
	b := make([]rune, shortCodeLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// handleShorten handles requests to shorten a URL
func handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}
	defer r.Body.Close()

	if req.URL == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	// Basic URL validation (can be more sophisticated)
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		http.Error(w, "Invalid URL format (must start with http:// or https://)", http.StatusBadRequest)
		return
	}

	mu.Lock() // Lock for writing
	var shortCode string
	for {
		shortCode = generateShortCode()
		if _, exists := urlStore[shortCode]; !exists { // Ensure code is unique
			break
		}
	}
	urlStore[shortCode] = req.URL
	mu.Unlock()

	// Construct the short URL (assuming service runs on localhost:8080)
	// You might want to make the base URL configurable
	shortenedURL := fmt.Sprintf("https://url-shortener-seven-theta.vercel.app/%s", shortCode)

	resp := ShortenResponse{ShortURL: shortenedURL}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Allow requests from Next.js frontend (development)
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Shortened URL: %s -> %s", req.URL, shortenedURL)
}

// handleRedirect handles requests to redirect from a short code to the original URL
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// CORS headers for OPTIONS preflight requests (needed for some browsers/setups)
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		http.NotFound(w, r)
		return
	}

	mu.RLock() // Lock for reading
	longURL, exists := urlStore[shortCode]
	mu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		log.Printf("Short code not found: %s", shortCode)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound) // 302 Found redirect
	log.Printf("Redirected %s to %s", shortCode, longURL)
}

func main() {
	http.HandleFunc("/shorten", handleShorten) // POST to create a short URL
	http.HandleFunc("/", handleRedirect)       // GET /<shortCode> to redirect

	// Handle OPTIONS requests globally for CORS preflight if needed for /shorten
	// This is a simplified way; a middleware or router specific options might be better.
	http.HandleFunc("/shorten/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // Add any other headers your client sends
			w.WriteHeader(http.StatusOK)
			return
		}
		// If not OPTIONS, let the original handler process it (if it matches path exactly)
		// For this setup, /shorten is handled by handleShorten
		handleShorten(w, r)
	})

	port := ":8080"
	log.Printf("Starting URL shortener service on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
