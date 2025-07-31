package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/yourorg/mysteryfactory/pkg/logger"
)

// promptService implements the PromptService interface
type promptService struct {
	prompts    map[string]*Prompt
	catalogPath string
	logger     *logger.Logger
	lastLoaded time.Time
}

// PromptCatalog represents the structure of the YAML prompt catalog
type PromptCatalog struct {
	Version     string             `yaml:"version"`
	Description string             `yaml:"description"`
	UpdatedAt   string             `yaml:"updated_at"`
	Prompts     map[string]*Prompt `yaml:"prompts"`
}

// NewPromptService creates a new prompt service instance
func NewPromptService(catalogPath string, logger *logger.Logger) (PromptService, error) {
	service := &promptService{
		prompts:     make(map[string]*Prompt),
		catalogPath: catalogPath,
		logger:      logger,
	}

	// Load prompts from catalog
	if err := service.loadCatalog(); err != nil {
		return nil, fmt.Errorf("failed to load prompt catalog: %w", err)
	}

	return service, nil
}

// GetPrompt retrieves a prompt by key
func (s *promptService) GetPrompt(ctx context.Context, key string) (*Prompt, error) {
	s.logger.Debug("Getting prompt", "key", key)

	// Check if catalog needs reloading (for development)
	if err := s.reloadIfNeeded(); err != nil {
		s.logger.Warn("Failed to reload catalog", "error", err)
	}

	prompt, exists := s.prompts[key]
	if !exists {
		s.logger.Error("Prompt not found", "key", key)
		return nil, fmt.Errorf("prompt not found: %s", key)
	}

	s.logger.Debug("Prompt retrieved", "key", key, "name", prompt.Name)
	return prompt, nil
}

// ListPrompts retrieves all available prompts
func (s *promptService) ListPrompts(ctx context.Context) ([]*Prompt, error) {
	s.logger.Debug("Listing all prompts")

	// Check if catalog needs reloading
	if err := s.reloadIfNeeded(); err != nil {
		s.logger.Warn("Failed to reload catalog", "error", err)
	}

	prompts := make([]*Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}

	s.logger.Debug("Prompts listed", "count", len(prompts))
	return prompts, nil
}

// GetPromptsByCategory retrieves prompts by category
func (s *promptService) GetPromptsByCategory(ctx context.Context, category string) ([]*Prompt, error) {
	s.logger.Debug("Getting prompts by category", "category", category)

	// Check if catalog needs reloading
	if err := s.reloadIfNeeded(); err != nil {
		s.logger.Warn("Failed to reload catalog", "error", err)
	}

	var prompts []*Prompt
	for _, prompt := range s.prompts {
		if prompt.Category == category {
			prompts = append(prompts, prompt)
		}
	}

	s.logger.Debug("Prompts retrieved by category", "category", category, "count", len(prompts))
	return prompts, nil
}

// CreatePrompt creates a new prompt (runtime creation)
func (s *promptService) CreatePrompt(ctx context.Context, req *CreatePromptRequest) (*Prompt, error) {
	s.logger.Info("Creating prompt", "key", req.Key, "name", req.Name)

	// Validate request
	if err := s.validatePromptRequest(req); err != nil {
		return nil, fmt.Errorf("invalid prompt request: %w", err)
	}

	// Check if prompt already exists
	if _, exists := s.prompts[req.Key]; exists {
		return nil, fmt.Errorf("prompt already exists: %s", req.Key)
	}

	// Create prompt
	prompt := &Prompt{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Template:    req.Template,
		Variables:   req.Variables,
		Metadata:    req.Metadata,
		Version:     "1.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate the template
	if err := s.ValidatePrompt(ctx, prompt); err != nil {
		return nil, fmt.Errorf("prompt validation failed: %w", err)
	}

	// Store in memory (in production, this would also persist to storage)
	s.prompts[req.Key] = prompt

	s.logger.Info("Prompt created successfully", "key", req.Key, "name", req.Name)
	return prompt, nil
}

// UpdatePrompt updates an existing prompt
func (s *promptService) UpdatePrompt(ctx context.Context, key string, req *UpdatePromptRequest) (*Prompt, error) {
	s.logger.Info("Updating prompt", "key", key)

	// Get existing prompt
	prompt, exists := s.prompts[key]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", key)
	}

	// Update fields
	if req.Name != nil {
		prompt.Name = *req.Name
	}
	if req.Description != nil {
		prompt.Description = *req.Description
	}
	if req.Category != nil {
		prompt.Category = *req.Category
	}
	if req.Template != nil {
		prompt.Template = *req.Template
	}
	if req.Variables != nil {
		prompt.Variables = req.Variables
	}
	if req.Metadata != nil {
		prompt.Metadata = req.Metadata
	}
	prompt.UpdatedAt = time.Now()

	// Validate the updated template
	if err := s.ValidatePrompt(ctx, prompt); err != nil {
		return nil, fmt.Errorf("prompt validation failed: %w", err)
	}

	s.logger.Info("Prompt updated successfully", "key", key)
	return prompt, nil
}

// DeletePrompt deletes a prompt
func (s *promptService) DeletePrompt(ctx context.Context, key string) error {
	s.logger.Info("Deleting prompt", "key", key)

	if _, exists := s.prompts[key]; !exists {
		return fmt.Errorf("prompt not found: %s", key)
	}

	delete(s.prompts, key)

	s.logger.Info("Prompt deleted successfully", "key", key)
	return nil
}

// ValidatePrompt validates a prompt template and variables
func (s *promptService) ValidatePrompt(ctx context.Context, prompt *Prompt) error {
	s.logger.Debug("Validating prompt", "key", prompt.Key)

	// Validate template syntax
	tmpl, err := template.New(prompt.Key).Parse(prompt.Template)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	// Extract template variables
	templateVars := extractTemplateVariables(prompt.Template)

	// Check if all template variables have corresponding variable definitions
	for _, templateVar := range templateVars {
		found := false
		for _, promptVar := range prompt.Variables {
			if promptVar.Name == templateVar {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("template variable '%s' not defined in variables", templateVar)
		}
	}

	// Check if all defined variables are used in template
	for _, promptVar := range prompt.Variables {
		if !contains(templateVars, promptVar.Name) {
			s.logger.Warn("Variable defined but not used in template", "variable", promptVar.Name, "prompt", prompt.Key)
		}
	}

	// Validate variable definitions
	for _, variable := range prompt.Variables {
		if err := s.validateVariable(variable); err != nil {
			return fmt.Errorf("invalid variable '%s': %w", variable.Name, err)
		}
	}

	// Test template execution with default values
	testData := make(map[string]interface{})
	for _, variable := range prompt.Variables {
		if variable.Default != nil {
			testData[variable.Name] = variable.Default
		} else if variable.Required {
			// Use a test value for required variables
			testData[variable.Name] = getTestValue(variable.Type)
		}
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, testData); err != nil {
		return fmt.Errorf("template execution failed: %w", err)
	}

	s.logger.Debug("Prompt validation successful", "key", prompt.Key)
	return nil
}

// TestPrompt tests a prompt with provided data
func (s *promptService) TestPrompt(ctx context.Context, key string, testData map[string]interface{}) (*PromptTestResult, error) {
	s.logger.Debug("Testing prompt", "key", key)

	startTime := time.Now()

	// Get prompt
	prompt, err := s.GetPrompt(ctx, key)
	if err != nil {
		return &PromptTestResult{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(startTime),
			TestedAt: time.Now(),
		}, nil
	}

	// Validate test data against prompt variables
	if err := s.validateTestData(prompt, testData); err != nil {
		return &PromptTestResult{
			Success:  false,
			Error:    fmt.Sprintf("invalid test data: %v", err),
			Duration: time.Since(startTime),
			TestedAt: time.Now(),
		}, nil
	}

	// Execute template
	tmpl, err := template.New(key).Parse(prompt.Template)
	if err != nil {
		return &PromptTestResult{
			Success:  false,
			Error:    fmt.Sprintf("template parse error: %v", err),
			Duration: time.Since(startTime),
			TestedAt: time.Now(),
		}, nil
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, testData); err != nil {
		return &PromptTestResult{
			Success:  false,
			Error:    fmt.Sprintf("template execution error: %v", err),
			Duration: time.Since(startTime),
			TestedAt: time.Now(),
		}, nil
	}

	result := buf.String()
	duration := time.Since(startTime)

	// Calculate estimated token usage (rough approximation)
	tokenCount := len(strings.Fields(result))

	s.logger.Debug("Prompt test completed", "key", key, "duration", duration, "tokens", tokenCount)

	return &PromptTestResult{
		Success:    true,
		Result:     result,
		Duration:   duration,
		TokensUsed: tokenCount,
		Cost:       float64(tokenCount) * 0.0001, // Mock cost calculation
		Metadata: map[string]interface{}{
			"template_variables": len(prompt.Variables),
			"result_length":      len(result),
		},
		TestedAt: time.Now(),
	}, nil
}

// RenderPrompt renders a prompt with the given data
func (s *promptService) RenderPrompt(ctx context.Context, key string, data map[string]interface{}) (string, error) {
	s.logger.Debug("Rendering prompt", "key", key)

	// Get prompt
	prompt, err := s.GetPrompt(ctx, key)
	if err != nil {
		return "", err
	}

	// Validate data
	if err := s.validateTestData(prompt, data); err != nil {
		return "", fmt.Errorf("invalid data: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New(key).Parse(prompt.Template)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// Private methods

// loadCatalog loads prompts from the YAML catalog file
func (s *promptService) loadCatalog() error {
	s.logger.Info("Loading prompt catalog", "path", s.catalogPath)

	// Read catalog file
	data, err := os.ReadFile(s.catalogPath)
	if err != nil {
		return fmt.Errorf("failed to read catalog file: %w", err)
	}

	// Parse YAML
	var catalog PromptCatalog
	if err := yaml.Unmarshal(data, &catalog); err != nil {
		return fmt.Errorf("failed to parse catalog YAML: %w", err)
	}

	// Load prompts
	s.prompts = make(map[string]*Prompt)
	for key, prompt := range catalog.Prompts {
		prompt.Key = key // Ensure key is set
		s.prompts[key] = prompt
	}

	s.lastLoaded = time.Now()
	s.logger.Info("Prompt catalog loaded successfully", "prompts_count", len(s.prompts), "version", catalog.Version)

	return nil
}

// reloadIfNeeded reloads the catalog if the file has been modified
func (s *promptService) reloadIfNeeded() error {
	// Check file modification time
	info, err := os.Stat(s.catalogPath)
	if err != nil {
		return err
	}

	if info.ModTime().After(s.lastLoaded) {
		s.logger.Info("Catalog file modified, reloading", "path", s.catalogPath)
		return s.loadCatalog()
	}

	return nil
}

// validatePromptRequest validates a create prompt request
func (s *promptService) validatePromptRequest(req *CreatePromptRequest) error {
	if req.Key == "" {
		return fmt.Errorf("key is required")
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Description == "" {
		return fmt.Errorf("description is required")
	}
	if req.Category == "" {
		return fmt.Errorf("category is required")
	}
	if req.Template == "" {
		return fmt.Errorf("template is required")
	}

	// Validate key format (should be category/name)
	if !strings.Contains(req.Key, "/") {
		return fmt.Errorf("key should be in format 'category/name'")
	}

	return nil
}

// validateVariable validates a prompt variable definition
func (s *promptService) validateVariable(variable PromptVariable) error {
	if variable.Name == "" {
		return fmt.Errorf("variable name is required")
	}
	if variable.Type == "" {
		return fmt.Errorf("variable type is required")
	}

	// Validate type
	validTypes := []string{"string", "integer", "float", "boolean", "array", "object"}
	if !contains(validTypes, variable.Type) {
		return fmt.Errorf("invalid variable type: %s", variable.Type)
	}

	return nil
}

// validateTestData validates test data against prompt variables
func (s *promptService) validateTestData(prompt *Prompt, data map[string]interface{}) error {
	// Check required variables
	for _, variable := range prompt.Variables {
		if variable.Required {
			if _, exists := data[variable.Name]; !exists {
				return fmt.Errorf("required variable '%s' is missing", variable.Name)
			}
		}
	}

	// Validate data types (basic validation)
	for key, value := range data {
		// Find variable definition
		var variable *PromptVariable
		for _, v := range prompt.Variables {
			if v.Name == key {
				variable = &v
				break
			}
		}

		if variable != nil {
			if err := s.validateDataType(value, variable.Type); err != nil {
				return fmt.Errorf("invalid type for variable '%s': %w", key, err)
			}
		}
	}

	return nil
}

// validateDataType validates that a value matches the expected type
func (s *promptService) validateDataType(value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "integer":
		switch value.(type) {
		case int, int32, int64:
			// Valid
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "float":
		switch value.(type) {
		case float32, float64:
			// Valid
		default:
			return fmt.Errorf("expected float, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// Helper functions

// extractTemplateVariables extracts variable names from a template string
func extractTemplateVariables(templateStr string) []string {
	var variables []string
	
	// Simple regex-like extraction for {{variable}} patterns
	parts := strings.Split(templateStr, "{{")
	for i := 1; i < len(parts); i++ {
		if closingIndex := strings.Index(parts[i], "}}"); closingIndex != -1 {
			variable := strings.TrimSpace(parts[i][:closingIndex])
			if variable != "" && !contains(variables, variable) {
				variables = append(variables, variable)
			}
		}
	}
	
	return variables
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getTestValue returns a test value for a given type
func getTestValue(varType string) interface{} {
	switch varType {
	case "string":
		return "test_value"
	case "integer":
		return 42
	case "float":
		return 3.14
	case "boolean":
		return true
	case "array":
		return []interface{}{"item1", "item2"}
	case "object":
		return map[string]interface{}{"key": "value"}
	default:
		return "unknown"
	}
}

// GetCatalogPath returns the default catalog path
func GetCatalogPath() string {
	// Try to find catalog in various locations
	possiblePaths := []string{
		"prompts/catalog.yaml",
		"./prompts/catalog.yaml",
		"/etc/mysteryfactory/prompts/catalog.yaml",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Default to relative path
	return "prompts/catalog.yaml"
}