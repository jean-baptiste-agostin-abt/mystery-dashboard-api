package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/jibe0123/mysteryfactory/pkg/logger"
)

// MessageRole represents the role of a message in a conversation
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// MessageContent represents the content of a message
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Message represents a single message in a conversation
type Message struct {
	Role    MessageRole      `json:"role"`
	Content []MessageContent `json:"content"`
}

// Conversation represents a conversation with multiple messages
type Conversation struct {
	Messages []Message `json:"messages"`
}

// FoundationModel represents different foundation models available
type FoundationModel string

const (
	ModelClaude4Sonnet FoundationModel = "anthropic.claude-3-5-sonnet-20241022-v2:0"
	ModelClaude4Haiku  FoundationModel = "anthropic.claude-3-5-haiku-20241022-v1:0"
	ModelClaude3Sonnet FoundationModel = "anthropic.claude-3-sonnet-20240229-v1:0"
)

// ModelConfig represents configuration for a specific foundation model
type ModelConfig struct {
	ModelID            string  `json:"model_id"`
	MaxTokens          int     `json:"max_tokens"`
	Temperature        float64 `json:"temperature"`
	TopP               float64 `json:"top_p"`
	AnthropicVersion   string  `json:"anthropic_version,omitempty"`
	SupportsStreaming  bool    `json:"supports_streaming"`
	SupportsSystemRole bool    `json:"supports_system_role"`
}

// Helper functions for creating messages and conversations

// NewTextMessage creates a new message with text content
func NewTextMessage(role MessageRole, text string) Message {
	return Message{
		Role: role,
		Content: []MessageContent{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(text string) Message {
	return NewTextMessage(RoleSystem, text)
}

// NewUserMessage creates a new user message
func NewUserMessage(text string) Message {
	return NewTextMessage(RoleUser, text)
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(text string) Message {
	return NewTextMessage(RoleAssistant, text)
}

// NewConversation creates a new conversation with the given messages
func NewConversation(messages ...Message) Conversation {
	return Conversation{
		Messages: messages,
	}
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(message Message) {
	c.Messages = append(c.Messages, message)
}

// AddUserMessage adds a user message to the conversation
func (c *Conversation) AddUserMessage(text string) {
	c.AddMessage(NewUserMessage(text))
}

// AddAssistantMessage adds an assistant message to the conversation
func (c *Conversation) AddAssistantMessage(text string) {
	c.AddMessage(NewAssistantMessage(text))
}

// AddSystemMessage adds a system message to the conversation
func (c *Conversation) AddSystemMessage(text string) {
	c.AddMessage(NewSystemMessage(text))
}

// NewConversationRequest creates a new conversation request with default settings
func NewConversationRequest(model FoundationModel, conversation Conversation) *ConversationRequest {
	return &ConversationRequest{
		Model:        model,
		Conversation: conversation,
	}
}

// WithMaxTokens sets the max tokens for the conversation request
func (r *ConversationRequest) WithMaxTokens(maxTokens int) *ConversationRequest {
	r.MaxTokens = maxTokens
	return r
}

// WithTemperature sets the temperature for the conversation request
func (r *ConversationRequest) WithTemperature(temperature float64) *ConversationRequest {
	r.Temperature = temperature
	return r
}

// WithTopP sets the top-p for the conversation request
func (r *ConversationRequest) WithTopP(topP float64) *ConversationRequest {
	r.TopP = topP
	return r
}

// WithStopWords sets the stop words for the conversation request
func (r *ConversationRequest) WithStopWords(stopWords []string) *ConversationRequest {
	r.StopWords = stopWords
	return r
}

// GetModelConfig returns configuration for a specific foundation model
func GetModelConfig(model FoundationModel) ModelConfig {
	switch model {
	case ModelClaude4Sonnet:
		return ModelConfig{
			ModelID:            string(model),
			MaxTokens:          8192,
			Temperature:        0.7,
			TopP:               0.9,
			AnthropicVersion:   "bedrock-2023-05-31",
			SupportsStreaming:  true,
			SupportsSystemRole: true,
		}
	case ModelClaude4Haiku:
		return ModelConfig{
			ModelID:            string(model),
			MaxTokens:          4096,
			Temperature:        0.7,
			TopP:               0.9,
			AnthropicVersion:   "bedrock-2023-05-31",
			SupportsStreaming:  true,
			SupportsSystemRole: true,
		}
	case ModelClaude3Sonnet:
		return ModelConfig{
			ModelID:            string(model),
			MaxTokens:          4096,
			Temperature:        0.7,
			TopP:               0.9,
			AnthropicVersion:   "bedrock-2023-05-31",
			SupportsStreaming:  true,
			SupportsSystemRole: true,
		}
	default:
		// Default to Claude 4 Sonnet
		return GetModelConfig(ModelClaude4Sonnet)
	}
}

// BedrockClient defines the interface for AWS Bedrock operations
type BedrockClient interface {
	InvokeModel(ctx context.Context, req *InvokeModelRequest) (*InvokeModelResponse, error)
	InvokeModelWithStreaming(ctx context.Context, req *InvokeModelRequest) (*StreamingResponse, error)
	InvokeConversation(ctx context.Context, req *ConversationRequest) (*InvokeModelResponse, error)
	InvokeConversationWithStreaming(ctx context.Context, req *ConversationRequest) (*StreamingResponse, error)
	Health(ctx context.Context) error
}

// InvokeModelRequest represents a request to invoke a Bedrock model
type InvokeModelRequest struct {
	ModelID     string                 `json:"model_id"`
	Prompt      string                 `json:"prompt"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	StopWords   []string               `json:"stop_words,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// InvokeModelResponse represents a response from Bedrock model invocation
type InvokeModelResponse struct {
	Content      string                 `json:"content"`
	TokensUsed   int                    `json:"tokens_used"`
	FinishReason string                 `json:"finish_reason"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt  time.Time              `json:"processed_at"`
}

// StreamingResponse represents a streaming response from Bedrock
type StreamingResponse struct {
	Stream chan StreamChunk
	Error  chan error
	Done   chan bool
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content     string                 `json:"content"`
	IsComplete  bool                   `json:"is_complete"`
	TokensUsed  int                    `json:"tokens_used,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// ConversationRequest represents a request to invoke a Bedrock model with conversation
type ConversationRequest struct {
	Model        FoundationModel        `json:"model"`
	Conversation Conversation           `json:"conversation"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TopP         float64                `json:"top_p,omitempty"`
	StopWords    []string               `json:"stop_words,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// bedrockClient implements the BedrockClient interface
type bedrockClient struct {
	client *bedrockruntime.Client
	logger *logger.Logger
	config *BedrockConfig
}

// BedrockConfig holds configuration for Bedrock client
type BedrockConfig struct {
	Region           string
	MaxRetries       int
	RetryDelay       time.Duration
	RequestTimeout   time.Duration
	DefaultModelID   string
	DefaultMaxTokens int
}

// NewBedrockClient creates a new Bedrock client
func NewBedrockClient(cfg *BedrockConfig, logger *logger.Logger) (BedrockClient, error) {
	if cfg == nil {
		cfg = &BedrockConfig{
			Region:           "us-east-1",
			MaxRetries:       3,
			RetryDelay:       time.Second,
			RequestTimeout:   30 * time.Second,
			DefaultModelID:   string(ModelClaude4Sonnet),
			DefaultMaxTokens: 8192,
		}
	}

	// Load AWS configuration
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithRetryMaxAttempts(cfg.MaxRetries),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Bedrock Runtime client
	client := bedrockruntime.NewFromConfig(awsConfig)

	return &bedrockClient{
		client: client,
		logger: logger,
		config: cfg,
	}, nil
}

// InvokeModel invokes a Bedrock model with retry logic
func (c *bedrockClient) InvokeModel(ctx context.Context, req *InvokeModelRequest) (*InvokeModelResponse, error) {
	c.logger.Info("Invoking Bedrock model", "model_id", req.ModelID, "prompt_length", len(req.Prompt))

	// Set defaults
	if req.ModelID == "" {
		req.ModelID = c.config.DefaultModelID
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.DefaultMaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.TopP == 0 {
		req.TopP = 0.9
	}

	// Prepare Claude request body
	claudeRequest := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
	}

	if len(req.StopWords) > 0 {
		claudeRequest["stop_sequences"] = req.StopWords
	}

	requestBody, err := json.Marshal(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Invoke model with retry logic
	var response *bedrockruntime.InvokeModelOutput
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("Retrying Bedrock request", "attempt", attempt, "error", lastErr)
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		response, lastErr = c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(req.ModelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        requestBody,
		})

		if lastErr == nil {
			break
		}

		c.logger.Error("Bedrock request failed", "attempt", attempt, "error", lastErr)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock model after %d attempts: %w", c.config.MaxRetries+1, lastErr)
	}

	// Parse Claude response
	var claudeResponse struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(response.Body, &claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract content
	var content string
	for _, c := range claudeResponse.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	result := &InvokeModelResponse{
		Content:      content,
		TokensUsed:   claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
		FinishReason: claudeResponse.StopReason,
		Metadata: map[string]interface{}{
			"input_tokens":  claudeResponse.Usage.InputTokens,
			"output_tokens": claudeResponse.Usage.OutputTokens,
			"model_id":      req.ModelID,
		},
		ProcessedAt: time.Now(),
	}

	c.logger.Info("Bedrock model invoked successfully",
		"model_id", req.ModelID,
		"tokens_used", result.TokensUsed,
		"content_length", len(result.Content))

	return result, nil
}

// InvokeModelWithStreaming invokes a Bedrock model with streaming response
func (c *bedrockClient) InvokeModelWithStreaming(ctx context.Context, req *InvokeModelRequest) (*StreamingResponse, error) {
	c.logger.Info("Invoking Bedrock model with streaming", "model_id", req.ModelID)

	// Set defaults
	if req.ModelID == "" {
		req.ModelID = c.config.DefaultModelID
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.DefaultMaxTokens
	}

	// Prepare Claude request body
	claudeRequest := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        req.MaxTokens,
		"temperature":       req.Temperature,
		"top_p":             req.TopP,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": req.Prompt,
			},
		},
	}

	requestBody, err := json.Marshal(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create streaming response channels
	streamResp := &StreamingResponse{
		Stream: make(chan StreamChunk, 100),
		Error:  make(chan error, 1),
		Done:   make(chan bool, 1),
	}

	// Start streaming in goroutine
	go func() {
		defer close(streamResp.Stream)
		defer close(streamResp.Error)
		defer close(streamResp.Done)

		// Invoke model with streaming
		response, err := c.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(req.ModelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        requestBody,
		})

		if err != nil {
			streamResp.Error <- fmt.Errorf("failed to invoke streaming model: %w", err)
			return
		}

		// Process streaming response
		stream := response.GetStream()
		for event := range stream.Events() {
			switch e := event.(type) {
			case *types.ResponseStreamMemberChunk:
				// Parse chunk
				var chunkData struct {
					Type  string `json:"type"`
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}

				if err := json.Unmarshal(e.Value.Bytes, &chunkData); err != nil {
					c.logger.Error("Failed to parse chunk", "error", err)
					continue
				}

				if chunkData.Type == "content_block_delta" && chunkData.Delta.Type == "text_delta" {
					chunk := StreamChunk{
						Content:     chunkData.Delta.Text,
						IsComplete:  false,
						ProcessedAt: time.Now(),
					}
					streamResp.Stream <- chunk
				} else if chunkData.Type == "message_stop" {
					// Message completed
					chunk := StreamChunk{
						Content:     "",
						IsComplete:  true,
						ProcessedAt: time.Now(),
					}
					streamResp.Stream <- chunk
				}

			default:
				c.logger.Debug("Received unknown event type", "type", fmt.Sprintf("%T", e))
			}
		}

		if err := stream.Err(); err != nil {
			streamResp.Error <- fmt.Errorf("streaming error: %w", err)
			return
		}

		streamResp.Done <- true
	}()

	return streamResp, nil
}

// Health checks the health of the Bedrock service
func (c *bedrockClient) Health(ctx context.Context) error {
	c.logger.Debug("Checking Bedrock health")

	// Simple health check by making a minimal request
	req := &InvokeModelRequest{
		ModelID:   c.config.DefaultModelID,
		Prompt:    "Hello",
		MaxTokens: 10,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.InvokeModel(ctx, req)
	if err != nil {
		return fmt.Errorf("Bedrock health check failed: %w", err)
	}

	c.logger.Debug("Bedrock health check passed")
	return nil
}

// InvokeConversation invokes a Bedrock model with a conversation
func (c *bedrockClient) InvokeConversation(ctx context.Context, req *ConversationRequest) (*InvokeModelResponse, error) {
	c.logger.Info("Invoking Bedrock model with conversation", "model", req.Model, "message_count", len(req.Conversation.Messages))

	// Get model configuration
	modelConfig := GetModelConfig(req.Model)

	// Set defaults from model config and request
	maxTokens := modelConfig.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	temperature := modelConfig.Temperature
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	topP := modelConfig.TopP
	if req.TopP > 0 {
		topP = req.TopP
	}

	// Convert conversation messages to Claude format
	messages := make([]map[string]interface{}, 0, len(req.Conversation.Messages))
	var systemMessage string

	for _, msg := range req.Conversation.Messages {
		if msg.Role == RoleSystem {
			// Extract system message content
			for _, content := range msg.Content {
				if content.Type == "text" {
					systemMessage += content.Text
				}
			}
			continue
		}

		// Convert message content
		var contentStr string
		for _, content := range msg.Content {
			if content.Type == "text" {
				contentStr += content.Text
			}
		}

		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": contentStr,
		})
	}

	// Prepare Claude request body
	claudeRequest := map[string]interface{}{
		"anthropic_version": modelConfig.AnthropicVersion,
		"max_tokens":        maxTokens,
		"temperature":       temperature,
		"top_p":             topP,
		"messages":          messages,
	}

	// Add system message if present
	if systemMessage != "" {
		claudeRequest["system"] = systemMessage
	}

	// Add stop sequences if provided
	if len(req.StopWords) > 0 {
		claudeRequest["stop_sequences"] = req.StopWords
	}

	requestBody, err := json.Marshal(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	// Invoke model with retry logic
	var response *bedrockruntime.InvokeModelOutput
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Warn("Retrying Bedrock conversation request", "attempt", attempt, "error", lastErr)
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		response, lastErr = c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(modelConfig.ModelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        requestBody,
		})

		if lastErr == nil {
			break
		}

		c.logger.Error("Bedrock conversation request failed", "attempt", attempt, "error", lastErr)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to invoke Bedrock model after %d attempts: %w", c.config.MaxRetries+1, lastErr)
	}

	// Parse Claude response
	var claudeResponse struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		StopReason string `json:"stop_reason"`
	}

	if err := json.Unmarshal(response.Body, &claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract content
	var content string
	for _, c := range claudeResponse.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	result := &InvokeModelResponse{
		Content:      content,
		TokensUsed:   claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens,
		FinishReason: claudeResponse.StopReason,
		Metadata: map[string]interface{}{
			"input_tokens":  claudeResponse.Usage.InputTokens,
			"output_tokens": claudeResponse.Usage.OutputTokens,
			"model_id":      modelConfig.ModelID,
			"model":         string(req.Model),
		},
		ProcessedAt: time.Now(),
	}

	c.logger.Info("Bedrock conversation completed successfully",
		"model", req.Model,
		"tokens_used", result.TokensUsed,
		"content_length", len(result.Content))

	return result, nil
}

// InvokeConversationWithStreaming invokes a Bedrock model with conversation and streaming response
func (c *bedrockClient) InvokeConversationWithStreaming(ctx context.Context, req *ConversationRequest) (*StreamingResponse, error) {
	c.logger.Info("Invoking Bedrock model with conversation and streaming", "model", req.Model, "message_count", len(req.Conversation.Messages))

	// Get model configuration
	modelConfig := GetModelConfig(req.Model)

	// Set defaults from model config and request
	maxTokens := modelConfig.MaxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	temperature := modelConfig.Temperature
	if req.Temperature > 0 {
		temperature = req.Temperature
	}

	topP := modelConfig.TopP
	if req.TopP > 0 {
		topP = req.TopP
	}

	// Convert conversation messages to Claude format
	messages := make([]map[string]interface{}, 0, len(req.Conversation.Messages))
	var systemMessage string

	for _, msg := range req.Conversation.Messages {
		if msg.Role == RoleSystem {
			// Extract system message content
			for _, content := range msg.Content {
				if content.Type == "text" {
					systemMessage += content.Text
				}
			}
			continue
		}

		// Convert message content
		var contentStr string
		for _, content := range msg.Content {
			if content.Type == "text" {
				contentStr += content.Text
			}
		}

		messages = append(messages, map[string]interface{}{
			"role":    string(msg.Role),
			"content": contentStr,
		})
	}

	// Prepare Claude request body
	claudeRequest := map[string]interface{}{
		"anthropic_version": modelConfig.AnthropicVersion,
		"max_tokens":        maxTokens,
		"temperature":       temperature,
		"top_p":             topP,
		"messages":          messages,
	}

	// Add system message if present
	if systemMessage != "" {
		claudeRequest["system"] = systemMessage
	}

	// Add stop sequences if provided
	if len(req.StopWords) > 0 {
		claudeRequest["stop_sequences"] = req.StopWords
	}

	requestBody, err := json.Marshal(claudeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create streaming response channels
	streamResp := &StreamingResponse{
		Stream: make(chan StreamChunk, 100),
		Error:  make(chan error, 1),
		Done:   make(chan bool, 1),
	}

	// Start streaming in goroutine
	go func() {
		defer close(streamResp.Stream)
		defer close(streamResp.Error)
		defer close(streamResp.Done)

		// Invoke model with streaming
		response, err := c.client.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
			ModelId:     aws.String(modelConfig.ModelID),
			ContentType: aws.String("application/json"),
			Accept:      aws.String("application/json"),
			Body:        requestBody,
		})

		if err != nil {
			streamResp.Error <- fmt.Errorf("failed to invoke streaming model: %w", err)
			return
		}

		// Process streaming response
		stream := response.GetStream()
		for event := range stream.Events() {
			switch e := event.(type) {
			case *types.ResponseStreamMemberChunk:
				// Parse chunk
				var chunkData struct {
					Type  string `json:"type"`
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}

				if err := json.Unmarshal(e.Value.Bytes, &chunkData); err != nil {
					c.logger.Error("Failed to parse chunk", "error", err)
					continue
				}

				if chunkData.Type == "content_block_delta" && chunkData.Delta.Type == "text_delta" {
					chunk := StreamChunk{
						Content:     chunkData.Delta.Text,
						IsComplete:  false,
						ProcessedAt: time.Now(),
					}
					streamResp.Stream <- chunk
				} else if chunkData.Type == "message_stop" {
					// Message completed
					chunk := StreamChunk{
						Content:     "",
						IsComplete:  true,
						ProcessedAt: time.Now(),
					}
					streamResp.Stream <- chunk
				}

			default:
				c.logger.Debug("Received unknown event type", "type", fmt.Sprintf("%T", e))
			}
		}

		if err := stream.Err(); err != nil {
			streamResp.Error <- fmt.Errorf("streaming error: %w", err)
			return
		}

		streamResp.Done <- true
	}()

	return streamResp, nil
}
