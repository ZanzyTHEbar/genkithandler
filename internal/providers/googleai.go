package providers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ZanzyTHEbar/genkithandler/pkg/domain"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

// RetryConfig defines retry behavior for API calls
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type GoogleAIProviderConfig struct {
	APIKey       string `json:"api_key" mapstructure:"api_key"`
	DefaultModel string `json:"default_model" mapstructure:"default_model"`
}

// GoogleAIProvider represents the Google AI provider configuration using Genkit's built-in plugin
// Store genkit instance
type GoogleAIProvider struct {
	initialized  bool
	retryConfig  RetryConfig
	config       GoogleAIProviderConfig
	logger       domain.Logger
	errorHandler domain.ErrorHandler
	genkit       *genkit.Genkit
}

// NewGoogleAIProvider creates a new Google AI provider instance
func NewGoogleAIProvider(logger domain.Logger, errorHandler domain.ErrorHandler) *GoogleAIProvider {
	// Default retry configuration
	retryConfig := RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}

	return &GoogleAIProvider{
		retryConfig:  retryConfig,
		logger:       logger,
		errorHandler: errorHandler,
		initialized:  false,
	}
}

// Initialize initializes the Google AI provider
func (p *GoogleAIProvider) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Parse configuration
	var googleConfig GoogleAIProviderConfig
	if err := p.parseConfig(config, &googleConfig); err != nil {
		return p.errorHandler.Wrap(err, "failed to parse configuration", map[string]interface{}{
			"config": config,
		})
	}

	// Validate required fields
	if googleConfig.APIKey == "" {
		return p.errorHandler.New("Google AI API key is required", map[string]interface{}{
			"config": config,
		})
	}

	if googleConfig.DefaultModel == "" {
		googleConfig.DefaultModel = "gemini-1.5-flash" // Default model
	}

	p.config = googleConfig

	// Initialize genkit instance first
	genkitInstance, err := genkit.Init(ctx)
	if err != nil {
		return p.errorHandler.Wrap(err, "failed to initialize Genkit", map[string]interface{}{
			"api_key_length": len(googleConfig.APIKey),
		})
	}

	// Initialize Google AI plugin with the genkit instance
	googleAIPlugin := &googlegenai.GoogleAI{
		APIKey: googleConfig.APIKey,
	}
	if err := googleAIPlugin.Init(ctx, genkitInstance); err != nil {
		return p.errorHandler.Wrap(err, "failed to initialize Google AI plugin", map[string]interface{}{
			"api_key_length": len(googleConfig.APIKey),
		})
	}

	p.genkit = genkitInstance
	p.initialized = true

	p.logger.Info("Google AI provider initialized successfully", map[string]interface{}{
		"default_model": googleConfig.DefaultModel,
	})

	return nil
}

// GetModel returns the configured model for Google AI
func (p *GoogleAIProvider) GetModel() string {
	if p.config.DefaultModel == "" {
		return "gemini-1.5-flash"
	}
	return p.config.DefaultModel
}

// GenerateText generates text using the Google AI provider
func (p *GoogleAIProvider) GenerateText(ctx context.Context, prompt string) (string, error) {
	if !p.initialized {
		return "", errors.New("provider not initialized")
	}

	// Use the built-in Genkit generate function with the Google AI model
	response, err := p.withRetry(ctx, func() (string, error) {
		// Get the model reference from the plugin
		model := googlegenai.GoogleAIModel(p.genkit, p.GetModel())

		result, err := genkit.GenerateText(ctx, p.genkit,
			ai.WithModel(model),
			ai.WithPrompt(prompt),
		)

		return result, err
	})

	return response, err
}

// GenerateWithStructuredOutput generates text with structured output using the Google AI provider
func (p *GoogleAIProvider) GenerateWithStructuredOutput(ctx context.Context, prompt string, outputType interface{}) (*ai.ModelResponse, error) {
	if !p.initialized {
		return nil, errors.New("provider not initialized")
	}

	// Use the built-in Genkit generate function with structured output
	return p.withRetryStructured(ctx, func() (*ai.ModelResponse, error) {
		// Get the model reference from the plugin
		model := googlegenai.GoogleAIModel(p.genkit, p.GetModel())

		result, err := genkit.Generate(ctx, p.genkit,
			ai.WithModel(model),
			ai.WithPrompt(prompt),
			ai.WithOutputType(outputType),
		)

		return result, err
	})
}

// withRetry implements retry logic for text generation
func (p *GoogleAIProvider) withRetry(ctx context.Context, fn func() (string, error)) (string, error) {
	var lastErr error

	for attempt := 0; attempt <= p.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := min(p.retryConfig.BaseDelay*time.Duration(1<<(attempt-1)), p.retryConfig.MaxDelay)

			slog.Debug("Retrying Google AI request",
				"attempt", attempt,
				"delay", delay.String(),
				"last_error", lastErr.Error())

			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			break
		}
	}

	return "", fmt.Errorf("google AI request failed after %d attempts: %w", p.retryConfig.MaxRetries+1, lastErr)
}

// withRetryStructured implements retry logic for structured generation
func (p *GoogleAIProvider) withRetryStructured(ctx context.Context, fn func() (*ai.ModelResponse, error)) (*ai.ModelResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= p.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate exponential backoff delay
			delay := min(p.retryConfig.BaseDelay*time.Duration(1<<(attempt-1)), p.retryConfig.MaxDelay)

			slog.Debug("Retrying Google AI structured request",
				"attempt", attempt,
				"delay", delay.String(),
				"last_error", lastErr.Error())

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !p.isRetryable(err) {
			break
		}
	}

	return nil, fmt.Errorf("google AI structured request failed after %d attempts: %w", p.retryConfig.MaxRetries+1, lastErr)
}

// isRetryable determines if an error should trigger a retry
func (p *GoogleAIProvider) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	// Retryable conditions for Google AI API
	retryablePatterns := []string{
		"rate limit",
		"too many requests",
		"quota exceeded",
		"service unavailable",
		"internal error",
		"timeout",
		"connection reset",
		"temporary failure",
		"server error",
		"resource exhausted",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// IsAvailable checks if the Google AI provider is available and configured
func (p *GoogleAIProvider) IsAvailable() bool {
	return p.config.APIKey != "" && p.initialized
}

// SupportsStructuredOutput indicates whether this provider supports structured output
func (p *GoogleAIProvider) SupportsStructuredOutput() bool {
	return true // Google AI/Gemini supports structured output
}

// GetMaxTokens returns the maximum token limit for the configured model
func (p *GoogleAIProvider) GetMaxTokens() int {
	// Return conservative limits for different Gemini models
	model := p.GetModel()
	switch model {
	case "gemini-1.5-pro", "gemini-1.5-pro-latest":
		return 2097152
	case "gemini-1.5-flash", "gemini-1.5-flash-latest":
		return 1048576
	case "gemini-2.0-flash":
		return 1048576
	default:
		return 32768
	}
}

// GenerateStream generates a streaming response (placeholder implementation)
func (p *GoogleAIProvider) GenerateStream(ctx context.Context, g *genkit.Genkit, prompt string) (<-chan StreamChunk, error) {
	if !p.initialized {
		resultChan := make(chan StreamChunk, 1)
		close(resultChan)
		return resultChan, errors.New("provider not initialized")
	}

	resultChan := make(chan StreamChunk, 10)

	// TODO: !!! Use Genkit's streaming capabilities when available !!!

	return resultChan, nil
}

// CallTool executes a tool through the AI model
func (p *GoogleAIProvider) CallTool(ctx context.Context, g *genkit.Genkit, toolName string, params map[string]interface{}) (*ToolCallResult, error) {
	if !p.initialized {
		return nil, errors.New("provider not initialized")
	}

	startTime := time.Now()

	// Validate inputs
	if toolName == "" {
		return &ToolCallResult{
			Result:   nil,
			Success:  false,
			Duration: time.Since(startTime),
			Error:    fmt.Errorf("tool name cannot be empty"),
		}, fmt.Errorf("tool name cannot be empty")
	}

	// Create a prompt for tool calling
	toolPrompt := fmt.Sprintf("Execute tool '%s' with parameters: %v. Return the result in a structured format.", toolName, params)

	// Use the provider's text generation with retry logic
	response, err := p.withRetry(ctx, func() (string, error) {
		// Get the model reference from the plugin
		model := googlegenai.GoogleAIModel(p.genkit, p.GetModel())

		result, err := genkit.Generate(ctx, p.genkit,
			ai.WithModel(model),
			ai.WithPrompt(toolPrompt),
		)

		if err != nil {
			return "", err
		}

		return result.Text(), nil
	})

	duration := time.Since(startTime)

	if err != nil {
		return &ToolCallResult{
			Result:   nil,
			Success:  false,
			Duration: duration,
			Error:    err,
			Metadata: map[string]interface{}{
				"provider":   "googleai",
				"model":      p.GetModel(),
				"tool_name":  toolName,
				"error_type": "generation_failed",
			},
		}, err
	}

	// Create successful result
	result := &ToolCallResult{
		Result:   response,
		Success:  true,
		Duration: duration,
		Metadata: map[string]interface{}{
			"provider":     "googleai",
			"model":        p.GetModel(),
			"tool_name":    toolName,
			"prompt_used":  toolPrompt,
			"response_len": len(response),
		},
	}

	return result, nil
}

// parseConfig parses the configuration map into GoogleAIProviderConfig struct
func (p *GoogleAIProvider) parseConfig(config map[string]interface{}, googleConfig *GoogleAIProviderConfig) error {
	if apiKey, ok := config["api_key"].(string); ok {
		googleConfig.APIKey = apiKey
	}
	if defaultModel, ok := config["default_model"].(string); ok {
		googleConfig.DefaultModel = defaultModel
	}
	return nil
}
