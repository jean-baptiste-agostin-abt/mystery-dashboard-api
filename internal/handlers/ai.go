package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/mysteryfactory/internal/services"
	"github.com/yourorg/mysteryfactory/pkg/logger"
)

// AIHandler handles AI-related HTTP requests
type AIHandler struct {
	aiService services.AIService
	logger    *logger.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService services.AIService, logger *logger.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// GenerateMagicBrush generates content using AI magic brush
// @Summary Generate content using magic brush
// @Description Generate titles, descriptions, or tags for videos using AI
// @Tags AI
// @Accept json
// @Produce json
// @Param tenant_id header string true "Tenant ID"
// @Param request body MagicBrushRequest true "Magic brush request"
// @Success 200 {object} MagicBrushResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/magic-brush [post]
func (h *AIHandler) GenerateMagicBrush(c *gin.Context) {
	h.logger.Info("Magic brush generation request received")

	// Get tenant ID from header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		h.logger.Error("Missing tenant ID in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Tenant ID is required",
		})
		return
	}

	// Parse request body
	var req services.MagicBrushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse magic brush request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if req.VideoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Video ID is required",
		})
		return
	}

	if req.BrushType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Brush type is required",
		})
		return
	}

	// Validate brush type
	validBrushTypes := []string{"title", "description", "tags"}
	isValidType := false
	for _, validType := range validBrushTypes {
		if req.BrushType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid brush type. Must be one of: title, description, tags",
		})
		return
	}

	// Generate content using AI service
	response, err := h.aiService.GenerateMagicBrush(c.Request.Context(), tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to generate magic brush content", "error", err, "tenant_id", tenantID, "video_id", req.VideoID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to generate content",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Magic brush content generated successfully", "tenant_id", tenantID, "video_id", req.VideoID, "brush_type", req.BrushType)
	c.JSON(http.StatusOK, gin.H{
		"message": "Content generated successfully",
		"data":    response,
	})
}

// TestPrompt tests a prompt from the catalog
// @Summary Test a prompt from the catalog
// @Description Test a prompt with provided data to see the rendered output
// @Tags AI
// @Accept json
// @Produce json
// @Param tenant_id header string true "Tenant ID"
// @Param request body TestPromptRequest true "Test prompt request"
// @Success 200 {object} TestPromptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/test-prompt [post]
func (h *AIHandler) TestPrompt(c *gin.Context) {
	h.logger.Info("Test prompt request received")

	// Get tenant ID from header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		h.logger.Error("Missing tenant ID in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Tenant ID is required",
		})
		return
	}

	// Parse request body
	var req TestPromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse test prompt request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate request
	if req.PromptKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Prompt key is required",
		})
		return
	}

	// Process with Bedrock
	result, err := h.aiService.ProcessWithBedrock(c.Request.Context(), req.PromptKey, req.TestData)
	if err != nil {
		h.logger.Error("Failed to test prompt", "error", err, "tenant_id", tenantID, "prompt_key", req.PromptKey)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to test prompt",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Prompt tested successfully", "tenant_id", tenantID, "prompt_key", req.PromptKey)
	c.JSON(http.StatusOK, gin.H{
		"message": "Prompt tested successfully",
		"data":    result,
	})
}

// GetPrompts lists available prompts
// @Summary List available prompts
// @Description Get a list of all available prompts in the catalog
// @Tags AI
// @Produce json
// @Param tenant_id header string true "Tenant ID"
// @Param category query string false "Filter by category"
// @Success 200 {object} PromptsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/ai/prompts [get]
func (h *AIHandler) GetPrompts(c *gin.Context) {
	h.logger.Info("Get prompts request received")

	// Get tenant ID from header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		h.logger.Error("Missing tenant ID in request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "Tenant ID is required",
		})
		return
	}

	category := c.Query("category")

	// TODO: Get prompts from prompt service
	// For now, return a mock response
	prompts := []map[string]interface{}{
		{
			"key":         "magic_brush/title_gen",
			"name":        "Video Title Generator",
			"description": "Generates engaging video titles",
			"category":    "magic_brush",
		},
		{
			"key":         "magic_brush/description_gen",
			"name":        "Video Description Generator",
			"description": "Creates comprehensive video descriptions",
			"category":    "magic_brush",
		},
		{
			"key":         "magic_brush/tags_gen",
			"name":        "Video Tags Generator",
			"description": "Generates relevant tags and hashtags",
			"category":    "magic_brush",
		},
	}

	// Filter by category if provided
	if category != "" {
		var filteredPrompts []map[string]interface{}
		for _, prompt := range prompts {
			if prompt["category"] == category {
				filteredPrompts = append(filteredPrompts, prompt)
			}
		}
		prompts = filteredPrompts
	}

	h.logger.Info("Prompts retrieved successfully", "tenant_id", tenantID, "count", len(prompts), "category", category)
	c.JSON(http.StatusOK, gin.H{
		"message": "Prompts retrieved successfully",
		"data":    prompts,
		"total":   len(prompts),
	})
}

// Request/Response types

// TestPromptRequest represents a request to test a prompt
type TestPromptRequest struct {
	PromptKey string                 `json:"prompt_key" binding:"required"`
	TestData  map[string]interface{} `json:"test_data,omitempty"`
}

// TestPromptResponse represents the response from testing a prompt
type TestPromptResponse struct {
	PromptKey string                 `json:"prompt_key"`
	Result    map[string]interface{} `json:"result"`
}

// PromptsResponse represents the response for listing prompts
type PromptsResponse struct {
	Prompts []map[string]interface{} `json:"prompts"`
	Total   int                      `json:"total"`
}
