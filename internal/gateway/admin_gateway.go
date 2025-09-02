package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type AdminGateway struct {
	servicesAPIURL string
	contentAPIURL  string
	rateLimiter    *UserRateLimiter
	auditLogger    *AuditLogger
}

type AdminProxyConfig struct {
	ServicesAPIURL string
	ContentAPIURL  string
	RateLimit      int
}

type UserClaims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

func NewAdminGateway(config *AdminProxyConfig) *AdminGateway {
	return &AdminGateway{
		servicesAPIURL: config.ServicesAPIURL,
		contentAPIURL:  config.ContentAPIURL,
		rateLimiter:    NewUserRateLimiter(config.RateLimit, time.Minute),
		auditLogger:    NewAuditLogger(),
	}
}

func (g *AdminGateway) SetupRoutes(router *mux.Router) {
	// Apply middleware in order
	router.Use(g.corsMiddleware)
	router.Use(g.authenticationMiddleware)
	router.Use(g.rateLimitMiddleware)
	router.Use(g.auditLoggingMiddleware)
	router.Use(g.securityHeadersMiddleware)
	router.Use(g.loggingMiddleware)
	
	// Health check (no auth required)
	router.HandleFunc("/health", g.healthCheckHandler).Methods("GET")
	router.HandleFunc("/health/ready", g.readinessCheckHandler).Methods("GET")
	
	// Admin API routes - require authentication and admin role
	adminAPI := router.PathPrefix("/admin/api/v1").Subrouter()
	adminAPI.Use(g.authorizationMiddleware([]string{"admin", "editor"}))
	
	// Services API admin routes
	servicesHandler := g.createReverseProxy(g.servicesAPIURL)
	adminAPI.PathPrefix("/services").Handler(http.StripPrefix("/admin/api/v1", servicesHandler))
	
	// Content API admin routes
	contentHandler := g.createReverseProxy(g.contentAPIURL)
	adminAPI.PathPrefix("/content").Handler(http.StripPrefix("/admin/api/v1", contentHandler))
}

func (g *AdminGateway) createReverseProxy(targetURL string) http.Handler {
	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", targetURL, err)
	}
	
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the proxy to handle errors
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v", r.URL.Path, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `{"error":"Service unavailable","message":"Backend service is not responding"}`)
	}
	
	// Modify requests to add user context
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		
		// Add user context to backend requests
		if userClaims, ok := r.Context().Value("user_claims").(*UserClaims); ok {
			r.Header.Set("X-User-ID", userClaims.UserID)
			r.Header.Set("X-User-Email", userClaims.Email)
			r.Header.Set("X-User-Roles", strings.Join(userClaims.Roles, ","))
		}
	}
	
	return proxy
}

func (g *AdminGateway) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Allow requests from admin origins only
		allowedOrigins := []string{
			"https://admin.international-center.com",
			"https://dashboard.international-center.com",
			"http://localhost:3001", // Development admin
		}
		
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (g *AdminGateway) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health checks
		if strings.HasPrefix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}
		
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			g.sendAuthError(w, "Missing or invalid authorization header")
			return
		}
		
		token := strings.TrimPrefix(authHeader, "Bearer ")
		userClaims, err := g.validateToken(token)
		if err != nil {
			g.sendAuthError(w, "Invalid token: "+err.Error())
			return
		}
		
		if !userClaims.IsActive {
			g.sendAuthError(w, "User account is inactive")
			return
		}
		
		// Add user claims to request context
		ctx := r.Context()
		ctx = contextWithUserClaims(ctx, userClaims)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

func (g *AdminGateway) authorizationMiddleware(requiredRoles []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value("user_claims").(*UserClaims)
			if !ok {
				g.sendAuthError(w, "User context not found")
				return
			}
			
			hasRequiredRole := false
			for _, userRole := range userClaims.Roles {
				for _, requiredRole := range requiredRoles {
					if userRole == requiredRole {
						hasRequiredRole = true
						break
					}
				}
				if hasRequiredRole {
					break
				}
			}
			
			if !hasRequiredRole {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"error":"Forbidden","message":"Insufficient permissions"}`)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

func (g *AdminGateway) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := r.Context().Value("user_claims").(*UserClaims)
		if !ok {
			// Skip rate limiting for health checks
			if strings.HasPrefix(r.URL.Path, "/health") {
				next.ServeHTTP(w, r)
				return
			}
			g.sendAuthError(w, "User context not found for rate limiting")
			return
		}
		
		if !g.rateLimiter.Allow(userClaims.UserID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, `{"error":"Rate limit exceeded","message":"Too many requests from this user"}`)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (g *AdminGateway) auditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip audit logging for health checks
		if strings.HasPrefix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}
		
		start := time.Now()
		userClaims, _ := r.Context().Value("user_claims").(*UserClaims)
		
		// Wrap response writer to capture status code and response size
		auditWriter := &auditResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(auditWriter, r)
		
		duration := time.Since(start)
		
		auditEvent := &AuditEvent{
			Timestamp:      start,
			UserID:         getUserID(userClaims),
			UserEmail:      getUserEmail(userClaims),
			Action:         fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			Resource:       r.URL.Path,
			IPAddress:      g.getClientIP(r),
			UserAgent:     r.Header.Get("User-Agent"),
			StatusCode:    auditWriter.statusCode,
			ResponseTime:  duration,
			RequestSize:   r.ContentLength,
			ResponseSize:  auditWriter.responseSize,
		}
		
		g.auditLogger.Log(auditEvent)
	})
}

func (g *AdminGateway) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		next.ServeHTTP(w, r)
	})
}

func (g *AdminGateway) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		clientIP := g.getClientIP(r)
		userClaims, _ := r.Context().Value("user_claims").(*UserClaims)
		
		// Wrap response writer to capture status code
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrappedWriter, r)
		
		duration := time.Since(start)
		userID := getUserID(userClaims)
		
		log.Printf("ADMIN_GATEWAY - IP: %s, User: %s, Method: %s, Path: %s, Status: %d, Duration: %v",
			clientIP, userID, r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
	})
}

func (g *AdminGateway) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","service":"admin-gateway"}`)
}

func (g *AdminGateway) readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check if backend services are available
	servicesHealthy := g.checkBackendHealth(g.servicesAPIURL + "/health")
	contentHealthy := g.checkBackendHealth(g.contentAPIURL + "/health")
	
	if !servicesHealthy || !contentHealthy {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"not_ready","services_api":%t,"content_api":%t}`, servicesHealthy, contentHealthy)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ready","service":"admin-gateway"}`)
}

func (g *AdminGateway) checkBackendHealth(healthURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK
}

func (g *AdminGateway) getClientIP(r *http.Request) string {
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

func (g *AdminGateway) sendAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":"Unauthorized","message":"%s"}`, message)
}

func (g *AdminGateway) validateToken(token string) (*UserClaims, error) {
	// In production, this would validate JWT token against identity provider
	// For CICD testing, return simple validation
	if token == "admin-test-token" {
		return &UserClaims{
			UserID:   "test-admin-user",
			Email:    "admin@international-center.com",
			Roles:    []string{"admin"},
			IsActive: true,
		}, nil
	}
	
	if token == "editor-test-token" {
		return &UserClaims{
			UserID:   "test-editor-user",
			Email:    "editor@international-center.com",
			Roles:    []string{"editor"},
			IsActive: true,
		}, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

func getUserID(claims *UserClaims) string {
	if claims != nil {
		return claims.UserID
	}
	return "anonymous"
}

func getUserEmail(claims *UserClaims) string {
	if claims != nil {
		return claims.Email
	}
	return ""
}

// UserRateLimiter implements user-based rate limiting
type UserRateLimiter struct {
	limits   map[string]*rateLimiter
	mutex    sync.RWMutex
	rate     int
	duration time.Duration
}

func NewUserRateLimiter(rate int, duration time.Duration) *UserRateLimiter {
	limiter := &UserRateLimiter{
		limits:   make(map[string]*rateLimiter),
		rate:     rate,
		duration: duration,
	}
	
	// Clean up old entries periodically
	go limiter.cleanupWorker()
	
	return limiter
}

func (rl *UserRateLimiter) Allow(userID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	limiter, exists := rl.limits[userID]
	
	if !exists {
		rl.limits[userID] = &rateLimiter{
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

func (rl *UserRateLimiter) cleanupWorker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case now := <-ticker.C:
			rl.mutex.Lock()
			for userID, limiter := range rl.limits {
				// Remove entries that haven't been used in the last hour
				if now.Sub(limiter.lastRequest) > time.Hour {
					delete(rl.limits, userID)
				}
			}
			rl.mutex.Unlock()
		}
	}
}

// auditResponseWriter wraps http.ResponseWriter to capture response size
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *auditResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *auditResponseWriter) Write(data []byte) (int, error) {
	rw.responseSize += int64(len(data))
	return rw.ResponseWriter.Write(data)
}