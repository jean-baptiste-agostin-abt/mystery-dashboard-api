package services

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/mysteryfactory/pkg/aws"
	"github.com/yourorg/mysteryfactory/pkg/logger"
	"github.com/yourorg/mysteryfactory/pkg/metrics"
)

// aiService implements the AIService interface
type aiService struct {
	promptService PromptService
	bedrockClient aws.BedrockClient
	logger        *logger.Logger
	metrics       *metrics.Metrics
}

// NewAIService creates a new AI service instance
func NewAIService(promptService PromptService, bedrockClient aws.BedrockClient, logger *logger.Logger, metrics *metrics.Metrics) AIService {
	return &aiService{
		promptService: promptService,
		bedrockClient: bedrockClient,
		logger:        logger,
		metrics:       metrics,
	}
}

// GenerateMagicBrush generates content using AI magic brush
func (s *aiService) GenerateMagicBrush(ctx context.Context, tenantID string, req *MagicBrushRequest) (*MagicBrushResponse, error) {
	start := time.Now()
	s.logger.Info("Generating magic brush content", "tenant_id", tenantID, "video_id", req.VideoID, "brush_type", req.BrushType)

	// Track in-flight AI requests
	s.metrics.IncrementAIInFlight()
	defer s.metrics.DecrementAIInFlight()

	// Determine prompt key based on brush type
	var promptKey string
	switch req.BrushType {
	case "title":
		promptKey = "magic_brush/title_gen"
	case "description":
		promptKey = "magic_brush/description_gen"
	case "tags":
		promptKey = "magic_brush/tags_gen"
	default:
		s.metrics.RecordMagicBrush(req.BrushType, "error", tenantID)
		s.metrics.RecordError("unsupported_brush_type", "ai_service", tenantID)
		return nil, fmt.Errorf("unsupported brush type: %s", req.BrushType)
	}

	// Prepare prompt data from request context
	promptData := make(map[string]interface{})
	if req.Context != nil {
		for key, value := range req.Context {
			promptData[key] = value
		}
	}

	// Add request-specific data
	promptData["video_id"] = req.VideoID
	if req.Language != "" {
		promptData["language"] = req.Language
	}
	if req.Tone != "" {
		promptData["tone"] = req.Tone
	}
	if req.MaxLength > 0 {
		promptData["max_length"] = req.MaxLength
	}

	// Set defaults for common fields
	if _, exists := promptData["platform"]; !exists {
		promptData["platform"] = "youtube"
	}
	if _, exists := promptData["audience"]; !exists {
		promptData["audience"] = "general audience"
	}

	// Process with Bedrock using the prompt
	result, err := s.ProcessWithBedrock(ctx, promptKey, promptData)
	if err != nil {
		s.logger.Error("Failed to process magic brush with Bedrock", "error", err, "prompt_key", promptKey)
		s.metrics.RecordMagicBrush(req.BrushType, "error", tenantID)
		s.metrics.RecordError("bedrock_processing_failed", "ai_service", tenantID)
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract result from Bedrock response
	var generatedContent string
	if resultStr, ok := result["result"].(string); ok {
		generatedContent = resultStr
	} else {
		generatedContent = fmt.Sprintf("Generated %s content", req.BrushType)
	}

	response := &MagicBrushResponse{
		VideoID:     req.VideoID,
		BrushType:   req.BrushType,
		Result:      generatedContent,
		Confidence:  0.85,
		Metadata:    result,
		ProcessedAt: time.Now(),
	}

	// Record successful magic brush request
	duration := time.Since(start)
	s.metrics.RecordMagicBrush(req.BrushType, "success", tenantID)

	// Extract tokens used from result metadata if available
	tokensUsed := 0
	if tokens, ok := result["tokens_used"].(int); ok {
		tokensUsed = tokens
	}

	// Record AI request metrics
	model := "claude-3-sonnet"
	if modelStr, ok := result["model"].(string); ok {
		model = modelStr
	}
	s.metrics.RecordAIRequest(model, promptKey, req.BrushType, "success", tenantID, duration, tokensUsed)

	s.logger.Info("Magic brush content generated", "tenant_id", tenantID, "video_id", req.VideoID, "brush_type", req.BrushType)
	return response, nil
}

// ProcessWithBedrock processes input using AWS Bedrock
func (s *aiService) ProcessWithBedrock(ctx context.Context, promptKey string, input map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Processing with Bedrock", "prompt_key", promptKey)

	// Render the prompt using the prompt service
	renderedPrompt, err := s.promptService.RenderPrompt(ctx, promptKey, input)
	if err != nil {
		s.logger.Error("Failed to render prompt", "error", err, "prompt_key", promptKey)
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	s.logger.Debug("Prompt rendered", "prompt_key", promptKey, "length", len(renderedPrompt))

	// Create conversation with the rendered prompt
	conversation := aws.NewConversation(
		aws.NewUserMessage(renderedPrompt),
	)

	// Create Bedrock conversation request using Claude 4
	bedrockReq := aws.NewConversationRequest(aws.ModelClaude4Sonnet, conversation).
		WithMaxTokens(8192).
		WithTemperature(0.7).
		WithTopP(0.9)

	// Add metadata
	bedrockReq.Metadata = map[string]interface{}{
		"prompt_key":      promptKey,
		"input_variables": len(input),
	}

	// Invoke Bedrock model with conversation
	startTime := time.Now()
	bedrockResp, err := s.bedrockClient.InvokeConversation(ctx, bedrockReq)
	if err != nil {
		s.logger.Error("Failed to invoke Bedrock model", "error", err, "prompt_key", promptKey)
		return nil, fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}
	processingTime := time.Since(startTime)

	// Build response
	result := map[string]interface{}{
		"prompt_key":      promptKey,
		"rendered_prompt": renderedPrompt,
		"result":          bedrockResp.Content,
		"confidence":      0.85, // Default confidence, could be enhanced based on model response
		"tokens_used":     bedrockResp.TokensUsed,
		"processing_time": processingTime.String(),
		"model":           bedrockReq.Model,
		"timestamp":       bedrockResp.ProcessedAt,
		"finish_reason":   bedrockResp.FinishReason,
		"metadata": map[string]interface{}{
			"prompt_length":    len(renderedPrompt),
			"response_length":  len(bedrockResp.Content),
			"input_variables":  len(input),
			"bedrock_metadata": bedrockResp.Metadata,
		},
	}

	s.logger.Info("Bedrock processing completed",
		"prompt_key", promptKey,
		"tokens_used", bedrockResp.TokensUsed,
		"processing_time", processingTime,
		"content_length", len(bedrockResp.Content))

	return result, nil
}
