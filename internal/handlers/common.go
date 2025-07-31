package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jibe0123/mysteryfactory/internal/config"
	"github.com/jibe0123/mysteryfactory/pkg/db"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
)

// BaseHandler contains common dependencies for all handlers
type BaseHandler struct {
	config *config.Config
	logger *logger.Logger
	db     *db.DB
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(cfg *config.Config, logger *logger.Logger, db *db.DB) *BaseHandler {
	return &BaseHandler{
		config: cfg,
		logger: logger,
		db:     db,
	}
}

// HealthCheck handler for health check endpoint
// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func HealthCheck(db *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		if err := db.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"message": "Database connection failed",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Service is running",
			"version": "1.0.0",
		})
	}
}

// ReadinessCheck handler for readiness check endpoint
// @Summary Readiness check
// @Description Check if the service is ready to serve requests
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ready [get]
func ReadinessCheck(db *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database readiness
		if err := db.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "not ready",
				"message": "Database not ready",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"message": "Service is ready to serve requests",
		})
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

// respondWithError sends an error response
func (h *BaseHandler) respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	})
}

// respondWithSuccess sends a success response
func (h *BaseHandler) respondWithSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

// respondWithPagination sends a paginated response
func (h *BaseHandler) respondWithPagination(c *gin.Context, data interface{}, total int64, page, limit int) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// getUserFromContext extracts user from gin context
func (h *BaseHandler) getUserFromContext(c *gin.Context) (string, string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", "", gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic}
	}

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", "", gin.Error{Err: http.ErrAbortHandler, Type: gin.ErrorTypePublic}
	}

	return userID.(string), tenantID.(string), nil
}

// getPaginationParams extracts pagination parameters from context
func (h *BaseHandler) getPaginationParams(c *gin.Context) (int, int) {
	limit, exists := c.Get("limit")
	if !exists {
		limit = 20
	}

	offset, exists := c.Get("offset")
	if !exists {
		offset = 0
	}

	return limit.(int), offset.(int)
}
