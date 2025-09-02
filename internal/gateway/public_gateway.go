package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/gorilla/mux"
)

type PublicGateway struct {
	daprClient  client.Client
	rateLimiter *IPRateLimiter
}

type ProxyConfig struct {
	RateLimit int
}

func NewPublicGateway(config *ProxyConfig) *PublicGateway {
	// Get Dapr gRPC endpoint from environment variable
	daprEndpoint := os.Getenv("DAPR_GRPC_ENDPOINT")
	if daprEndpoint == "" {
		daprEndpoint = "127.0.0.1:50001" // Default Dapr gRPC port
	}
	log.Printf("dapr client initializing for: %s", daprEndpoint)
	
	var daprClient client.Client
	var err error
	
	if daprEndpoint != "" {
		daprClient, err = client.NewClientWithAddress(daprEndpoint)
	} else {
		daprClient, err = client.NewClient()
	}
	
	if err != nil {
		log.Fatalf("Failed to create Dapr client: %v", err)
	}
	
	return &PublicGateway{
		daprClient:  daprClient,
		rateLimiter: NewIPRateLimiter(config.RateLimit, time.Minute),
	}
}

func (g *PublicGateway) Close() {
	if g.daprClient != nil {
		g.daprClient.Close()
	}
}

func (g *PublicGateway) SetupRoutes(router *mux.Router) {
	// Apply middleware
	router.Use(g.corsMiddleware)
	router.Use(g.rateLimitMiddleware)
	router.Use(g.securityHeadersMiddleware)
	router.Use(g.loggingMiddleware)
	
	// Health check
	router.HandleFunc("/health", g.healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", g.readinessCheckHandler).Methods("GET")
	
	// Services API routes
	router.PathPrefix("/api/v1/services").Handler(http.HandlerFunc(g.handleServicesProxy)).Methods("GET", "POST", "PUT", "DELETE")
	
	// Content API routes
	router.PathPrefix("/api/v1/content").Handler(http.HandlerFunc(g.handleContentProxy)).Methods("GET", "POST", "PUT", "DELETE")
}

func (g *PublicGateway) handleServicesProxy(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// Extract the path to forward to the service
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/services")
	if path == "" {
		path = "/"
	}
	
	// Add query parameters if present
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	
	// Use Dapr service invocation
	resp, err := g.daprClient.InvokeMethodWithContent(ctx, "services-api", path, r.Method, nil)
	if err != nil {
		log.Printf("Dapr service invocation error for services-api%s: %v", path, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `{"error":"Service unavailable","message":"Backend service is not responding"}`)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (g *PublicGateway) handleContentProxy(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	// Extract the path to forward to the service
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/content")
	if path == "" {
		path = "/"
	}
	
	// Add query parameters if present
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	
	// Use Dapr service invocation
	resp, err := g.daprClient.InvokeMethodWithContent(ctx, "content-api", path, r.Method, nil)
	if err != nil {
		log.Printf("Dapr service invocation error for content-api%s: %v", path, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `{"error":"Service unavailable","message":"Backend service is not responding"}`)
		return
	}
	
	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (g *PublicGateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Allow requests from public website origins
		allowedOrigins := []string{
			"https://international-center.com",
			"https://www.international-center.com",
			"http://localhost:3000", // Development
			"http://localhost:4321", // Astro dev server
		}
		
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (g *PublicGateway) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := g.getClientIP(r)
		
		if !g.rateLimiter.Allow(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":"Rate limit exceeded","message":"Too many requests from this IP address"}`)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (g *PublicGateway) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next.ServeHTTP(w, r)
	})
}

func (g *PublicGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := g.getClientIP(r)
		
		// Wrap response writer to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrappedWriter, r)
		
		duration := time.Since(start)
		log.Printf("PUBLIC_GATEWAY - IP: %s, Method: %s, Path: %s, Status: %d, Duration: %v",
			clientIP, r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
	})
}

func (g *PublicGateway) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","service":"public-gateway"}`)
}

func (g *PublicGateway) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if backend services are available via Dapr
	servicesHealthy := g.checkBackendHealthViaDapr("services-api")
	contentHealthy := g.checkBackendHealthViaDapr("content-api")
	
	if !servicesHealthy || !contentHealthy {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","services_api":%t,"content_api":%t}`, servicesHealthy, contentHealthy)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ready","service":"public-gateway"}`)
}

func (g *PublicGateway) checkBackendHealthViaDapr(appID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := g.daprClient.InvokeMethod(ctx, appID, "/health", "GET")
	if err != nil {
		log.Printf("Health check failed for %s: %v", appID, err)
		return false
	}
	
	return true
}

func (g *PublicGateway) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	
	// Fall back to remote address
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	
	return ip
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// IPRateLimiter implements IP-based rate limiting
type IPRateLimiter struct {
	limits   map[string]*rateLimiter
	mutex    sync.RWMutex
	rate     int
	duration time.Duration
}

type rateLimiter struct {
	requests    int
	resetTime   time.Time
	lastRequest time.Time
}

func NewIPRateLimiter(rate int, duration time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limits:   make(map[string]*rateLimiter),
		rate:     rate,
		duration: duration,
	}
	
	// Clean up old entries periodically
	go limiter.cleanupWorker()
	
	return limiter
}

func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	limiter, exists := rl.limits[ip]
	
	if !exists {
		rl.limits[ip] = &rateLimiter{
			requests:    1,
			resetTime:   now.Add(rl.duration),
			lastRequest: now,
		}
		return true
	}
	
	// Reset if duration has passed
	if now.After(limiter.resetTime) {
		limiter.requests = 1
		limiter.resetTime = now.Add(rl.duration)
		limiter.lastRequest = now
		return true
	}
	
	limiter.lastRequest = now
	limiter.requests++
	
	return limiter.requests <= rl.rate
}

func (rl *IPRateLimiter) cleanupWorker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case now := <-ticker.C:
			rl.mutex.Lock()
			for ip, limiter := range rl.limits {
				// Remove entries that haven't been used in the last hour
				if now.Sub(limiter.lastRequest) > time.Hour {
					delete(rl.limits, ip)
				}
			}
			rl.mutex.Unlock()
		}
	}
}