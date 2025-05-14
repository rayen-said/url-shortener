package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os" // Import os to get the PORT environment variable
	"strings"
	"sync"
	"time"

	"github.com/rs/cors"
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
	// CORS middleware handles OPTIONS requests and sets headers,
	// so we only need to handle POST here.
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
	// Note: This is a simple check. A more robust validator is recommended.
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

	// Construct the short URL using the expected frontend domain
	// It's better to get this base URL from an environment variable
	// in a real deployment (e.g., VERCEL_URL for the frontend).
	// For now, using the hardcoded Vercel domain from your example.
	shortenedURL := fmt.Sprintf("https://url-shortener-seven-theta.vercel.app/%s", shortCode)

	resp := ShortenResponse{ShortURL: shortenedURL}
	w.Header().Set("Content-Type", "application/json")
	// CORS headers are handled by the middleware now, remove manual setting
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	// w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Shortened URL: %s -> %s", req.URL, shortenedURL)
}

// handleRedirect handles requests to redirect from a short code to the original URL
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	// CORS middleware handles OPTIONS requests, so we only need to handle GET here.
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract the short code from the URL path
	// r.URL.Path will be like "/ABCDEF"
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		// If path is just "/", it's not a short code, maybe serve a landing page?
		// For this example, we'll treat it as not found for a short code.
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

	// Perform the redirect
	http.Redirect(w, r, longURL, http.StatusFound) // 302 Found redirect
	log.Printf("Redirected %s to %s", shortCode, longURL)
}

func main() {
	// Define your allowed origins (the domains your frontend will be hosted on)
	// This should include your Vercel production domain, preview domains, and localhost for dev.
	// Reading this from an environment variable in Railway is a good practice for production.
	allowedOrigins := []string{
		"http://localhost:3000",                                  // For local Next.js development
		"https://url-shortener-rayen-saids-projects.vercel.app/", // Your Vercel production domain
		// Add other Vercel preview domains if needed, or use a wildcard cautiously in dev/staging
		// e.g., "https://*-url-shortener-seven-theta.vercel.app" // Wildcard for preview deployments (use with caution)
	}

	// Configure the CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"}, // Only need methods used by frontend for API calls and redirects
		AllowedHeaders:   []string{"Content-Type"},           // Only need headers your frontend sends for API calls
		AllowCredentials: true,                               // Set to true if your frontend sends cookies or auth headers
		// Debug: true, // Uncomment in development to see CORS logs
	})

	// Create your main router
	router := http.NewServeMux()

	// Register your handlers with the router
	router.HandleFunc("/shorten", handleShorten) // POST to create a short URL
	// The root path "/" will be handled by handleRedirect for short codes
	router.HandleFunc("/", handleRedirect) // GET /<shortCode> to redirect

	// Wrap your router with the CORS middleware
	// This is the key change: http.ListenAndServe will now use the handler
	// provided by the CORS middleware, which wraps your router.
	handler := c.Handler(router)

	// Get the port from the environment variable provided by Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if PORT is not set (e.g., for local testing outside Railway)
	}
	listenAddr := fmt.Sprintf(":%s", port)

	log.Printf("Starting URL shortener service on %s", listenAddr)
	// Use the wrapped handler here
	if err := http.ListenAndServe(listenAddr, handler); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
