package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jibe0123/mysteryfactory/internal/config"
	"github.com/jibe0123/mysteryfactory/pkg/db"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *AuthHandler) {
	gin.SetMode(gin.TestMode)

	// Mock config
	cfg := &config.Config{
		JWTSecret:   "test-secret-key",
		Environment: "test",
		LogLevel:    "info",
	}

	// Mock logger
	logger := logger.New("info", "test")

	// Mock database (nil for now, would use test DB in real implementation)
	var mockDB *db.DB

	// Create handler
	authHandler := NewAuthHandler(cfg, logger, mockDB)

	// Setup router
	r := gin.New()
	return r, authHandler
}

func TestAuthHandler_Login(t *testing.T) {
	r, authHandler := setupTestRouter()
	r.POST("/login", authHandler.Login)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "Valid login request",
			requestBody: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"token", "expires_at", "user"},
		},
		{
			name: "Invalid email format",
			requestBody: LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "message"},
		},
		{
			name: "Missing password",
			requestBody: LoginRequest{
				Email: "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "message"},
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error", "message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request body
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Perform request
			r.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Assert expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Additional assertions for successful login
			if tt.expectedStatus == http.StatusOK {
				assert.NotEmpty(t, response["token"], "Token should not be empty")
				assert.NotEmpty(t, response["expires_at"], "Expires at should not be empty")

				user, ok := response["user"].(map[string]interface{})
				assert.True(t, ok, "User should be an object")
				assert.NotEmpty(t, user["email"], "User email should not be empty")
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	r, authHandler := setupTestRouter()
	r.POST("/register", authHandler.Register)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "Valid registration request",
			requestBody: map[string]interface{}{
				"email":      "newuser@example.com",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
				"role":       "editor",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid email format",
			requestBody: map[string]interface{}{
				"email":      "invalid-email",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
				"role":       "editor",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing required fields",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	r, authHandler := setupTestRouter()

	// Add middleware to simulate authenticated user
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Set("tenant_id", "test-tenant-123")
		c.Next()
	})

	r.GET("/me", authHandler.GetProfile)

	req, err := http.NewRequest("GET", "/me", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Contains(t, response, "data")

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test-user-123", data["id"])
	assert.Equal(t, "test-tenant-123", data["tenant_id"])
}

func TestAuthHandler_GetProfile_Unauthorized(t *testing.T) {
	r, authHandler := setupTestRouter()
	r.GET("/me", authHandler.GetProfile)

	req, err := http.NewRequest("GET", "/me", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Contains(t, response, "message")
}

func TestAuthHandler_Logout(t *testing.T) {
	r, authHandler := setupTestRouter()

	// Add middleware to simulate authenticated user
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Set("tenant_id", "test-tenant-123")
		c.Next()
	})

	r.POST("/logout", authHandler.Logout)

	req, err := http.NewRequest("POST", "/logout", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Equal(t, "Logged out successfully", response["message"])
}
