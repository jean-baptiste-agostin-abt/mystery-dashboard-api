package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jibe0123/mysteryfactory/internal/models"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
	"github.com/rs/cors"
	"golang.org/x/time/rate"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS(allowed string) gin.HandlerFunc {
	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(allowed, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Requested-With", "X-Request-ID", "X-Tenant-ID"},
		AllowCredentials: true,
	})
	return func(ctx *gin.Context) {
		c.HandlerFunc(ctx.Writer, ctx.Request)
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.Request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	})
}

// Logger middleware for structured logging
func Logger(log *logger.Logger) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request ID
		requestID, _ := c.Get("request_id")

		// Get user info if available
		userID := ""
		tenantID := ""
		if user, exists := c.Get("user"); exists {
			if u, ok := user.(*models.User); ok {
				userID = u.ID
				tenantID = u.TenantID
			}
		}

		// Build full path
		if raw != "" {
			path = path + "?" + raw
		}

		// Log the request
		fields := []interface{}{
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		if requestID != nil {
			fields = append(fields, "request_id", requestID)
		}
		if userID != "" {
			fields = append(fields, "user_id", userID)
		}
		if tenantID != "" {
			fields = append(fields, "tenant_id", tenantID)
		}

		// Log based on status code
		status := c.Writer.Status()
		if status >= 500 {
			log.Error("HTTP request completed with server error", fields...)
		} else if status >= 400 {
			log.Warn("HTTP request completed with client error", fields...)
		} else {
			log.Info("HTTP request completed", fields...)
		}
	})
}

// RateLimiter middleware for rate limiting requests
func RateLimiter() gin.HandlerFunc {
	// Create a rate limiter that allows 100 requests per minute
	limiter := rate.NewLimiter(rate.Every(time.Minute/100), 100)

	return gin.HandlerFunc(func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate Limit Exceeded",
				"message": "Too many requests, please try again later",
			})
			c.Abort()
			return
		}
		c.Next()
	})
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuth middleware for JWT token authentication
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*JWTClaims); ok {
			// Create user object from claims
			user := &models.User{
				ID:       claims.UserID,
				TenantID: claims.TenantID,
				Email:    claims.Email,
				Role:     claims.Role,
			}

			// Set user in context
			c.Set("user", user)
			c.Set("user_id", claims.UserID)
			c.Set("tenant_id", claims.TenantID)
			c.Set("user_role", claims.Role)
		}

		c.Next()
	})
}

// TenantResolver middleware resolves tenant information
func TenantResolver() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get tenant ID from JWT claims (already set by JWTAuth middleware)
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Tenant information not found",
			})
			c.Abort()
			return
		}

		// You can add additional tenant validation here
		// For now, we just ensure the tenant ID is present
		c.Set("current_tenant_id", tenantID)
		c.Next()
	})
}

// RequireRole middleware checks if user has required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User information not found",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid user data",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		if u.Role != requiredRole && u.Role != "admin" { // Admin can access everything
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequirePermission middleware checks if user has specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User information not found",
			})
			c.Abort()
			return
		}

		u, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid user data",
			})
			c.Abort()
			return
		}

		// Check if user has required permission
		if !u.HasPermission(permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// WebhookAuth middleware for webhook authentication
func WebhookAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get platform from URL parameter
		platform := c.Param("platform")
		if platform == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Platform parameter is required",
			})
			c.Abort()
			return
		}

		// Validate webhook signature based on platform
		switch platform {
		case "youtube":
			if !validateYouTubeWebhook(c) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized",
					"message": "Invalid webhook signature",
				})
				c.Abort()
				return
			}
		case "tiktok":
			if !validateTikTokWebhook(c) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error":   "Unauthorized",
					"message": "Invalid webhook signature",
				})
				c.Abort()
				return
			}
		default:
			// For other platforms, implement similar validation
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Bad Request",
				"message": "Unsupported platform",
			})
			c.Abort()
			return
		}

		c.Set("webhook_platform", platform)
		c.Next()
	})
}

// validateYouTubeWebhook validates YouTube webhook signature
func validateYouTubeWebhook(c *gin.Context) bool {
	// Implement YouTube webhook signature validation
	// This is a placeholder - implement actual validation logic
	signature := c.Request.Header.Get("X-Hub-Signature")
	return signature != ""
}

// validateTikTokWebhook validates TikTok webhook signature
func validateTikTokWebhook(c *gin.Context) bool {
	// Implement TikTok webhook signature validation
	// This is a placeholder - implement actual validation logic
	signature := c.Request.Header.Get("X-TikTok-Signature")
	return signature != ""
}

// Timeout middleware adds request timeout
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":   "Request Timeout",
				"message": "Request took too long to process",
			})
			c.Abort()
		}
	})
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})
}

// PaginationMiddleware middleware for handling pagination parameters
func PaginationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Default values
		limit := 20
		offset := 0

		// Parse limit
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
				if limit > 100 { // Max limit
					limit = 100
				}
			}
		}

		// Parse offset
		if offsetStr := c.Query("offset"); offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Parse page (alternative to offset)
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
				offset = (page - 1) * limit
			}
		}

		c.Set("limit", limit)
		c.Set("offset", offset)
		c.Next()
	})
}
