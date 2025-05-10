package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	// OpenRouterCompletionURL is the endpoint for OpenRouter's completion API
	OpenRouterCompletionURL = "https://openrouter.co/v1/completions"
	
	// OpenRouterChatCompletionURL is the endpoint for OpenRouter's chat completion API
	OpenRouterChatCompletionURL = "https://openrouter.co/v1/chat/completions"
	
	// OpenRouterModelsURL is the endpoint for retrieving available models
	OpenRouterModelsURL = "https://openrouter.co/v1/models"
)

// Client is a client for the OpenRouter API
type Client struct {
	apiKey       string
	defaultModel string
	httpClient   *http.Client
}

// NewClient creates a new OpenRouter client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENROUTER_API_KEY environment variable is not set")
	}
	
	defaultModel := os.Getenv("OPENROUTER_DEFAULT_MODEL")
	if defaultModel == "" {
		// Use a default model if not specified
		defaultModel = "openai/gpt-3.5-turbo"
	}
	
	return &Client{
		apiKey:       apiKey,
		defaultModel: defaultModel,
		httpClient:   &http.Client{},
	}, nil
}

// Model represents an AI model available in OpenRouter
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Created     int64  `json:"created"`
	Description string `json:"description"`
	
	Architecture struct {
		InputModalities  []string `json:"input_modalities"`
		OutputModalities []string `json:"output_modalities"`
		Tokenizer        string   `json:"tokenizer"`
	} `json:"architecture"`
	
	TopProvider struct {
		IsModerated bool `json:"is_moderated"`
	} `json:"top_provider"`
	
	Pricing struct {
		Prompt      string `json:"prompt"`
		Completion  string `json:"completion"`
		Image       string `json:"image"`
		Request     string `json:"request"`
		InputCache  string `json:"input_cache"`
		WebSearch   string `json:"web_search"`
		InternalReasoning string `json:"internal_reasoning"`
	} `json:"pricing"`
	
	ContextLength    int                    `json:"context_length"`
	PerRequestLimits map[string]interface{} `json:"per_request_limits"`
}

// ModelsResponse represents the response from the models endpoint
type ModelsResponse struct {
	Data []Model `json:"data"`
}

// GetAvailableModels retrieves the list of available models from OpenRouter
func (c *Client) GetAvailableModels() ([]Model, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", OpenRouterModelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	// Parse response
	var modelsResp ModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	
	return modelsResp.Data, nil
}

// SetModel sets the default model for the client
func (c *Client) SetModel(modelID string) {
	c.defaultModel = modelID
}

// CompletionRequest represents a request to the completion API
type CompletionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

// CompletionResponse represents a response from the completion API
type CompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// GetCompletion sends a completion request to OpenRouter
func (c *Client) GetCompletion(prompt string, options ...any) (*CompletionResponse, error) {
	req := CompletionRequest{
		Model:       c.defaultModel,
		Prompt:      prompt,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	
	// Process optional parameters
	for _, option := range options {
		switch opt := option.(type) {
		case map[string]interface{}:
			if model, ok := opt["model"].(string); ok {
				req.Model = model
			}
			if maxTokens, ok := opt["max_tokens"].(int); ok {
				req.MaxTokens = maxTokens
			}
			if temp, ok := opt["temperature"].(float64); ok {
				req.Temperature = temp
			}
		}
	}
	
	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", OpenRouterCompletionURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	// Parse response
	var completionResp CompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	
	return &completionResp, nil
}

// ChatMessage represents a message in a chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completion API
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

// ChatCompletionResponse represents a response from the chat completion API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// GetChatCompletion sends a chat completion request to OpenRouter
func (c *Client) GetChatCompletion(messages []ChatMessage, options ...any) (*ChatCompletionResponse, error) {
	req := ChatCompletionRequest{
		Model:       c.defaultModel,
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.7,
	}
	
	// Process optional parameters
	for _, option := range options {
		switch opt := option.(type) {
		case map[string]interface{}:
			if model, ok := opt["model"].(string); ok {
				req.Model = model
			}
			if maxTokens, ok := opt["max_tokens"].(int); ok {
				req.MaxTokens = maxTokens
			}
			if temp, ok := opt["temperature"].(float64); ok {
				req.Temperature = temp
			}
		}
	}
	
	// Convert request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", OpenRouterChatCompletionURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Add headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}
	
	// Parse response
	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}
	
	return &chatResp, nil
} 