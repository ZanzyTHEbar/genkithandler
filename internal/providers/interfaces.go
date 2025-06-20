package providers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/firebase/genkit/go/ai"
)

// AIProvider represents a simplified AI provider interface
type AIProvider interface {
	// Initialize sets up the provider with configuration
	Initialize(ctx context.Context, config map[string]interface{}) error

	// GenerateText generates a simple text response
	GenerateText(ctx context.Context, prompt string) (string, error)

	// GenerateWithStructuredOutput generates a response with structured output
	GenerateWithStructuredOutput(ctx context.Context, prompt string, outputType interface{}) (*ai.ModelResponse, error)

	// GetModel returns the configured model name
	GetModel() string

	// IsAvailable checks if the provider is properly configured and available
	IsAvailable() bool
}

const (
	ProviderTypeOpenAI    ProviderType = "openai"
	ProviderTypeGoogleAI  ProviderType = "googleai"
	ProviderTypeOllamaAI  ProviderType = "ollama"
	ProviderTypeAnthropic ProviderType = "anthropic"
	ProviderTypeAzureAI   ProviderType = "azureai"
)

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content  string                 `json:"content"`
	Delta    map[string]interface{} `json:"delta,omitempty"`
	Done     bool                   `json:"done"`
	Error    error                  `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCallResult represents the result of a tool execution
type ToolCallResult struct {
	Result   interface{}            `json:"result"`
	Success  bool                   `json:"success"`
	Error    error                  `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// StructuredResponse represents a parsed structured response
type StructuredResponse struct {
	Data     interface{}            `json:"data"`
	Schema   string                 `json:"schema,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderManager manages multiple AI providers with fallback support
type LegacyProviderManager struct {
	providers map[string]AIProvider
	primary   string
	fallback  string
}

// ExtendedAIProvider provides additional capabilities beyond the basic interface
type ExtendedAIProvider interface {
	AIProvider

	// SupportsStructuredOutput returns whether this provider supports structured output
	SupportsStructuredOutput() bool

	// GetMaxTokens returns the maximum token limit for the current model
	GetMaxTokens() int
}

// ProviderType represents the type of AI provider
type ProviderType string

// GenerateRequest represents a request to generate AI content
type GenerateRequest struct {
	// Model specifies which AI model to use
	Model string `json:"model"`

	// Prompt is the input prompt for the AI
	Prompt string `json:"prompt"`

	// Schema defines the expected response structure (for structured output)
	Schema *ResponseSchema `json:"schema,omitempty"`

	// Tools available for the AI to call
	Tools []ToolDefinition `json:"tools,omitempty"`

	// Temperature controls randomness (0.0 to 1.0)
	Temperature *float64 `json:"temperature,omitempty"`

	// MaxTokens limits response length
	MaxTokens *int `json:"max_tokens,omitempty"`

	// Stream indicates if streaming response is desired
	Stream bool `json:"stream,omitempty"`

	// Context provides additional context for the request
	Context map[string]interface{} `json:"context,omitempty"`
}

// GenerateResponse represents a structured response from an AI model
type GenerateResponse struct {
	// Content is the main response content
	Content string `json:"content"`

	// StructuredData contains the parsed structured response (if schema was provided)
	StructuredData interface{} `json:"structured_data,omitempty"`

	// ToolCalls contains any tool calls made by the AI
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// Metadata contains additional response information
	Metadata ResponseMetadata `json:"metadata"`

	// Provider indicates which provider generated this response
	Provider string `json:"provider"`

	// Model indicates which model was used
	Model string `json:"model"`
}

// ResponseSchema defines the structure for structured output
type ResponseSchema struct {
	// Type specifies the schema type (e.g., "object", "array")
	Type string `json:"type"`

	// Properties defines object properties (for type "object")
	Properties map[string]SchemaProperty `json:"properties,omitempty"`

	// Items defines array item schema (for type "array")
	Items *ResponseSchema `json:"items,omitempty"`

	// Required lists required properties
	Required []string `json:"required,omitempty"`

	// Description provides schema documentation
	Description string `json:"description,omitempty"`
}

// SchemaProperty defines a property in a response schema
type SchemaProperty struct {
	Type        string                    `json:"type"`
	Description string                    `json:"description,omitempty"`
	Properties  map[string]SchemaProperty `json:"properties,omitempty"`
	Items       *ResponseSchema           `json:"items,omitempty"`
	Required    []string                  `json:"required,omitempty"`
}

// ToolRequest represents a request to execute a tool
type ToolRequest struct {
	// ToolName identifies the tool to execute
	ToolName string `json:"tool_name"`

	// Parameters contains the tool input parameters
	Parameters map[string]interface{} `json:"parameters"`

	// Context provides additional execution context
	Context map[string]interface{} `json:"context,omitempty"`
}

// ToolResponse represents the result of tool execution
type ToolResponse struct {
	// Result contains the tool execution result
	Result interface{} `json:"result"`

	// Success indicates if the tool executed successfully
	Success bool `json:"success"`

	// Error contains any execution error
	Error error `json:"error,omitempty"`

	// Metadata contains additional response information
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents an AI model's call to a tool
type ToolCall struct {
	// ID uniquely identifies this tool call
	ID string `json:"id"`

	// Name is the tool name
	Name string `json:"name"`

	// Parameters contains the tool input
	Parameters json.RawMessage `json:"parameters"`

	// Result contains the tool execution result (if executed)
	Result *ToolResponse `json:"result,omitempty"`
}

// ResponseMetadata contains additional information about the response
type ResponseMetadata struct {
	// TokenUsage tracks token consumption
	TokenUsage TokenUsage `json:"token_usage"`

	// Duration tracks request processing time
	Duration int64 `json:"duration_ms"`

	// RequestID for tracing and debugging
	RequestID string `json:"request_id"`

	// FinishReason indicates why the response ended
	FinishReason string `json:"finish_reason,omitempty"`

	// Additional provider-specific metadata
	ProviderMetadata map[string]interface{} `json:"provider_metadata,omitempty"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig contains configuration for AI providers
type ProviderConfig struct {
	// Primary provider to use
	Primary string `json:"primary"`

	// Fallback provider if primary fails
	Fallback string `json:"fallback,omitempty"`

	// Provider-specific configurations
	Providers map[string]ProviderSettings `json:"providers"`

	// Default model for each provider
	DefaultModels map[string]string `json:"default_models,omitempty"`
}

// ProviderSettings contains settings for a specific provider
type ProviderSettings struct {
	// APIKey for authentication
	APIKey string `json:"api_key,omitempty"`

	// BaseURL for custom endpoints
	BaseURL string `json:"base_url,omitempty"`

	// Timeout for requests
	TimeoutSeconds int `json:"timeout_seconds,omitempty"`

	// RetryAttempts for failed requests
	RetryAttempts int `json:"retry_attempts,omitempty"`

	// Additional provider-specific settings
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// ToolDefinition represents a tool that can be called by AI
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}
