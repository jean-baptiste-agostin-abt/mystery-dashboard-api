package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/mysteryfactory/internal/config"
	"github.com/yourorg/mysteryfactory/internal/models"
	"github.com/yourorg/mysteryfactory/pkg/db"
	"github.com/yourorg/mysteryfactory/pkg/logger"
)

func setupVideoTestRouter() (*gin.Engine, *VideoHandler) {
	gin.SetMode(gin.TestMode)

	// Mock config
	cfg := &config.Config{
		Environment: "test",
		LogLevel:    "info",
	}

	// Mock logger
	logger := logger.New("info", "test")

	// Mock database
	var mockDB *db.DB

	// Create handler
	videoHandler := NewVideoHandler(cfg, logger, mockDB)

	// Setup router
	r := gin.New()
	return r, videoHandler
}

func addAuthMiddleware(r *gin.Engine) {
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Set("tenant_id", "test-tenant-123")
		c.Set("limit", 20)
		c.Set("offset", 0)
		c.Next()
	})
}

func TestVideoHandler_ListVideos(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.GET("/videos", videoHandler.ListVideos)

	req, err := http.NewRequest("GET", "/videos", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "total")
	assert.Contains(t, response, "page")
	assert.Contains(t, response, "limit")
}

func TestVideoHandler_ListVideos_Unauthorized(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	r.GET("/videos", videoHandler.ListVideos)

	req, err := http.NewRequest("GET", "/videos", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVideoHandler_CreateVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.POST("/videos", videoHandler.CreateVideo)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "Valid video creation",
			requestBody: models.CreateVideoRequest{
				Title:       "Test Video",
				Description: "A test video description",
				FileName:    "test-video.mp4",
				FileSize:    1024000,
				Format:      "mp4",
				Tags:        []string{"test", "video"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing title",
			requestBody: models.CreateVideoRequest{
				Description: "A test video description",
				FileName:    "test-video.mp4",
				FileSize:    1024000,
				Format:      "mp4",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid file size",
			requestBody: models.CreateVideoRequest{
				Title:    "Test Video",
				FileName: "test-video.mp4",
				FileSize: 0,
				Format:   "mp4",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/videos", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "message")
				assert.Contains(t, response, "data")

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotEmpty(t, data["id"])
				assert.Equal(t, "test-user-123", data["user_id"])
				assert.Equal(t, "test-tenant-123", data["tenant_id"])
			}
		})
	}
}

func TestVideoHandler_GetVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.GET("/videos/:id", videoHandler.GetVideo)

	tests := []struct {
		name           string
		videoID        string
		expectedStatus int
	}{
		{
			name:           "Valid video ID",
			videoID:        "video-123",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty video ID",
			videoID:        "",
			expectedStatus: http.StatusNotFound, // Gin returns 404 for missing path params
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/videos/" + tt.videoID
			req, err := http.NewRequest("GET", path, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "message")
				assert.Contains(t, response, "data")

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.videoID, data["id"])
			}
		})
	}
}

func TestVideoHandler_UpdateVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.PUT("/videos/:id", videoHandler.UpdateVideo)

	tests := []struct {
		name           string
		videoID        string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:    "Valid update",
			videoID: "video-123",
			requestBody: models.UpdateVideoRequest{
				Title:       stringPtr("Updated Title"),
				Description: stringPtr("Updated description"),
				Tags:        []string{"updated", "tags"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Partial update",
			videoID: "video-123",
			requestBody: models.UpdateVideoRequest{
				Title: stringPtr("Only Title Updated"),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			videoID:        "video-123",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				jsonBody = []byte(str)
			} else {
				jsonBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			path := "/videos/" + tt.videoID
			req, err := http.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestVideoHandler_DeleteVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.DELETE("/videos/:id", videoHandler.DeleteVideo)

	req, err := http.NewRequest("DELETE", "/videos/video-123", nil)
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
	assert.Equal(t, "video-123", data["id"])
}

func TestVideoHandler_UploadVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.POST("/videos/:id/upload", videoHandler.UploadVideo)

	// Create a multipart form with a file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a file field
	fileWriter, err := writer.CreateFormFile("file", "test-video.mp4")
	require.NoError(t, err)

	// Write some dummy file content
	_, err = fileWriter.Write([]byte("dummy video content"))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/videos/video-123/upload", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	assert.Equal(t, "video-123", data["id"])
	assert.Equal(t, "test-video.mp4", data["filename"])
}

func TestVideoHandler_PublishVideo(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.POST("/videos/:id/publish", videoHandler.PublishVideo)

	tests := []struct {
		name           string
		videoID        string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:    "Valid publication request",
			videoID: "video-123",
			requestBody: models.CreatePublicationJobRequest{
				VideoID:  "video-123",
				Platform: "youtube",
				Config: map[string]interface{}{
					"title":       "Published Video",
					"description": "Video published to YouTube",
					"privacy":     "public",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Invalid platform",
			videoID: "video-123",
			requestBody: models.CreatePublicationJobRequest{
				VideoID:  "video-123",
				Platform: "invalid-platform",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing request body",
			videoID:        "video-123",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			path := "/videos/" + tt.videoID + "/publish"
			req, err := http.NewRequest("POST", path, bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.Contains(t, response, "message")
				assert.Contains(t, response, "data")

				data, ok := response["data"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.videoID, data["video_id"])
				assert.NotEmpty(t, data["publication_id"])
			}
		})
	}
}

func TestVideoHandler_GetVideoPublications(t *testing.T) {
	r, videoHandler := setupVideoTestRouter()
	addAuthMiddleware(r)
	r.GET("/videos/:id/publications", videoHandler.GetVideoPublications)

	req, err := http.NewRequest("GET", "/videos/video-123/publications", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
	assert.Contains(t, response, "data")

	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 0) // Should return array (even if empty)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
