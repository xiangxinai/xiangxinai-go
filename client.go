package xiangxinai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// DefaultBaseURL Default API base URL
	DefaultBaseURL = "https://api.xiangxinai.cn/v1"
	// DefaultModel Default model name
	DefaultModel = "Xiangxin-Guardrails-Text"
	// DefaultTimeout Default request timeout (seconds)
	DefaultTimeout = 30
	// DefaultMaxRetries Default maximum retry count
	DefaultMaxRetries = 3
	// UserAgent User agent
	UserAgent = "xiangxinai-go/2.6.2"
)

// Client Xiangxin AI Guardrails client - Context-aware AI guardrail based on LLM
//
// This client provides a simple interface for interacting with the Xiangxin AI Guardrails API.
// The guardrail uses context-aware technology to understand the conversation context for safety detection.
//
// Example usage:
//
//	client := xiangxinai.NewClient("your-api-key")
//	
//	// Check user input
//	result, err := client.CheckPrompt(context.Background(), "用户问题")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Check output content (based on context)
//	result, err := client.CheckResponseCtx(context.Background(), "用户问题", "助手回答")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Check conversation context
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "问题"),
//		xiangxinai.NewMessage("assistant", "回答"),
//	}
//	result, err := client.CheckConversation(context.Background(), messages)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "high_risk/medium_risk/low_risk/no_risk"
//	fmt.Println(result.SuggestAction)    // "pass/reject/replace"
type Client struct {
	client     *resty.Client
	maxRetries int
}

// NewClient Create new client, using default configuration
func NewClient(apiKey string) *Client {
	return NewClientWithConfig(&ClientConfig{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
	})
}

// NewClientWithConfig Create new client, using custom configuration
func NewClientWithConfig(config *ClientConfig) *Client {
	if config.APIKey == "" {
		panic("API key cannot be empty")
	}
	
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	
	maxRetries := config.MaxRetries
	if maxRetries < 0 {
		maxRetries = DefaultMaxRetries
	}
	
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(time.Duration(timeout) * time.Second)
	client.SetHeader("Authorization", "Bearer "+config.APIKey)
	client.SetHeader("Content-Type", "application/json")
	client.SetHeader("User-Agent", UserAgent)
	
	return &Client{
		client:     client,
		maxRetries: maxRetries,
	}
}

// createSafeResponse Create safe response
func (c *Client) createSafeResponse() *GuardrailResponse {
	return &GuardrailResponse{
		ID: "guardrails-safe-default",
		Result: &GuardrailResult{
			Compliance: &ComplianceResult{
				RiskLevel:  "no_risk",
				Categories: []string{},
			},
			Security: &SecurityResult{
				RiskLevel:  "no_risk",
				Categories: []string{},
			},
		},
		OverallRiskLevel: "no_risk",
		SuggestAction:    "pass",
		SuggestAnswer:    nil,
	}
}

// CheckPrompt Check user input safety
//
// Parameters:
//   - ctx: Context
//   - content: User input content to check
//
// Return value:
//   - *GuardrailResponse: Detection result, format as:
//     {
//       "id": "guardrails-xxx",
//       "result": {
//         "compliance": {
//           "risk_level": "high_risk/medium_risk/low_risk/no_risk",
//           "categories": ["violent crime", "sensitive political topics"]
//         },
//         "security": {
//           "risk_level": "high_risk/medium_risk/low_risk/no_risk",
//           "categories": ["prompt attack"]
//         }
//       },
//       "overall_risk_level": "high_risk/medium_risk/low_risk/no_risk",
//       "suggest_action": "pass/reject/replace",
//       "suggest_answer": "Suggested answer content"
//     }
//   - error: Error information
//
// Possible error types:
//   - ValidationError: Invalid input parameters
//   - AuthenticationError: Authentication failed
//   - RateLimitError: Exceeded rate limit
//   - XiangxinAIError: Other API errors
//
// Example:
//
//	result, err := client.CheckPrompt(ctx, "I want to learn programming")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "no_risk"
//	fmt.Println(result.SuggestAction)    // "pass"
//	fmt.Println(result.Result.Compliance.RiskLevel) // "no_risk"
func (c *Client) CheckPrompt(ctx context.Context, content string, userID ...string) (*GuardrailResponse, error) {
	// If content is an empty string, return no risk
	if strings.TrimSpace(content) == "" {
		return c.createSafeResponse(), nil
	}

	requestData := map[string]interface{}{
		"input": strings.TrimSpace(content),
	}

	// Add optional userID parameter
	if len(userID) > 0 && userID[0] != "" {
		requestData["xxai_app_user_id"] = userID[0]
	}

	return c.makeRequestWithData(ctx, "POST", "/guardrails/input", requestData)
}

// CheckConversation Check conversation context safety - context-aware detection
//
// This is the core functionality of the guardrail, capable of understanding the complete conversation context for safety detection.
// Instead of checking each message separately, it analyzes the overall conversation safety.
//
// Parameters:
//   - ctx: Context
//   - messages: Conversation message list, containing the complete conversation between user and assistant, each message contains role('user' or 'assistant') and content
//   - userID: Optional parameter, tenant AI application user ID, used for user-level risk control and audit tracking
//
// Return value:
//   - *GuardrailResponse: Detection result based on conversation context, format as CheckPrompt
//   - error: Error information
//
// Example:
//
//	// Check conversation safety between user question and assistant answer
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "User question"),
//		xiangxinai.NewMessage("assistant", "Assistant answer"),
//	}
//	result, err := client.CheckConversation(ctx, messages, "user-123")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "no_risk"
//	fmt.Println(result.SuggestAction)    // "pass"
func (c *Client) CheckConversation(ctx context.Context, messages []*Message, userID ...string) (*GuardrailResponse, error) {
	return c.CheckConversationWithModel(ctx, messages, DefaultModel, userID...)
}

// CheckConversationWithModel Check conversation context safety, specify model
func (c *Client) CheckConversationWithModel(ctx context.Context, messages []*Message, model string, userID ...string) (*GuardrailResponse, error) {
	if len(messages) == 0 {
		return nil, NewValidationError("messages cannot be empty")
	}
	
	// Validate message format
	var validatedMessages []*Message
	allEmpty := true // Mark whether all content are empty
	
	for _, msg := range messages {
		if msg == nil {
			return nil, NewValidationError("message cannot be nil")
		}
		
		if msg.Role != "user" && msg.Role != "system" && msg.Role != "assistant" {
			return nil, NewValidationError("message role must be one of: user, system, assistant")
		}
		
		if len(msg.Content) > 1000000 {
			return nil, NewValidationError("content too long (max 1000000 characters)")
		}
		
		content := strings.TrimSpace(msg.Content)
		// Check if there is non-empty content
		if content != "" {
			allEmpty = false
			// Only add non-empty messages to validatedMessages
			validatedMessages = append(validatedMessages, &Message{
				Role:    msg.Role,
				Content: content,
			})
		}
	}
	
	// If all messages' content are empty, return no risk
	if allEmpty {
		return c.createSafeResponse(), nil
	}
	
	// Ensure at least one message
	if len(validatedMessages) == 0 {
		return c.createSafeResponse(), nil
	}
	
	request := &GuardrailRequest{
		Model:    model,
		Messages: validatedMessages,
	}

	// Add optional userID parameter
	if len(userID) > 0 && userID[0] != "" {
		if request.ExtraBody == nil {
			request.ExtraBody = make(map[string]interface{})
		}
		request.ExtraBody["xxai_app_user_id"] = userID[0]
	}

	return c.makeRequest(ctx, "POST", "/guardrails", request)
}

// CheckResponseCtx Check user input and model output safety - context-aware detection
//
// This is the core functionality of the guardrail, capable of understanding the user input and model output context for safety detection.
// The guardrail will detect whether the model output is safe and compliant based on the user question context.
//
// Parameters:
//   - ctx: Context
//   - prompt: User input text content, used to help the guardrail understand the context semantics
//   - response: Model output text content, actual detection object
//
// Return value:
//   - *GuardrailResponse: Detection result based on context, format as CheckPrompt
//   - error: Error information
//
// Possible error types:
//   - ValidationError: Invalid input parameters
//   - AuthenticationError: Authentication failed
//   - RateLimitError: Exceeded rate limit
//   - XiangxinAIError: Other API errors
//
// Example:
//
//	result, err := client.CheckResponseCtx(ctx, "I want to learn cooking", "I can teach you some simple home-cooked dishes")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "no_risk"
//	fmt.Println(result.SuggestAction)    // "pass"
func (c *Client) CheckResponseCtx(ctx context.Context, prompt, response string, userID ...string) (*GuardrailResponse, error) {
	// If prompt and response are empty strings, return no risk
	if strings.TrimSpace(prompt) == "" && strings.TrimSpace(response) == "" {
		return c.createSafeResponse(), nil
	}

	requestData := map[string]interface{}{
		"input":  strings.TrimSpace(prompt),
		"output": strings.TrimSpace(response),
	}

	// Add optional userID parameter
	if len(userID) > 0 && userID[0] != "" {
		requestData["xxai_app_user_id"] = userID[0]
	}

	return c.makeRequestWithData(ctx, "POST", "/guardrails/output", requestData)
}

// encodeBase64FromPath Encode image to base64 format
func (c *Client) encodeBase64FromPath(imagePath string) (string, error) {
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		// Get image from URL
		resp, err := http.Get(imagePath)
		if err != nil {
			return "", fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read image data: %w", err)
		}

		return base64.StdEncoding.EncodeToString(data), nil
	}

	// Read image from local file
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// CheckPromptImage Check text prompt and image safety - multi-modal detection
//
// Combine text semantics and image content for safety detection.
//
// Parameters:
//   - ctx: Context
//   - prompt: Text prompt (can be empty)
//   - image: Local path or HTTP(S) link of image file (cannot be empty)
//
// Return value:
//   - *GuardrailResponse: Detection result
//   - error: Error information
//
// Example:
//
//	// Check local image
//	result, err := client.CheckPromptImage(ctx, "Is this image safe?", "/path/to/image.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Check network image
//	result, err := client.CheckPromptImage(ctx, "", "https://example.com/image.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel)
func (c *Client) CheckPromptImage(ctx context.Context, prompt, image string, userID ...string) (*GuardrailResponse, error) {
	return c.CheckPromptImageWithModel(ctx, prompt, image, "Xiangxin-Guardrails-VL", userID...)
}

// CheckPromptImageWithModel Check text prompt and image safety, specify model
func (c *Client) CheckPromptImageWithModel(ctx context.Context, prompt, image, model string, userID ...string) (*GuardrailResponse, error) {
	if image == "" {
		return nil, NewValidationError("image path cannot be empty")
	}

	// Encode image
	imageBase64, err := c.encodeBase64FromPath(image)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewValidationError(fmt.Sprintf("image file not found: %s", image))
		}
		return nil, NewXiangxinAIError(fmt.Sprintf("failed to encode image: %v", err), err)
	}

	// Build message content
	content := []interface{}{}
	if strings.TrimSpace(prompt) != "" {
		content = append(content, map[string]string{
			"type": "text",
			"text": strings.TrimSpace(prompt),
		})
	}
	content = append(content, map[string]interface{}{
		"type": "image_url",
		"image_url": map[string]string{
			"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
		},
	})

	messages := []*Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	request := &GuardrailRequest{
		Model:    model,
		Messages: messages,
	}

	// Add optional userID parameter
	if len(userID) > 0 && userID[0] != "" {
		if request.ExtraBody == nil {
			request.ExtraBody = make(map[string]interface{})
		}
		request.ExtraBody["xxai_app_user_id"] = userID[0]
	}

	return c.makeRequest(ctx, "POST", "/guardrails", request)
}

// CheckPromptImages Check text prompt and multiple images safety - multi-modal detection
//
// Combine text semantics and multiple image content for safety detection.
//
// Parameters:
//   - ctx: Context
//   - prompt: Text prompt (can be empty)
//   - images: Local path or HTTP(S) link list of image file (cannot be empty)
//
// Return value:
//   - *GuardrailResponse: Detection result
//   - error: Error information
//
// Example:
//
//	images := []string{"/path/to/image1.jpg", "https://example.com/image2.jpg"}
//	result, err := client.CheckPromptImages(ctx, "Are these images safe?", images)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel)
func (c *Client) CheckPromptImages(ctx context.Context, prompt string, images []string, userID ...string) (*GuardrailResponse, error) {
	return c.CheckPromptImagesWithModel(ctx, prompt, images, "Xiangxin-Guardrails-VL", userID...)
}

// CheckPromptImagesWithModel Check text prompt and multiple images safety, specify model
func (c *Client) CheckPromptImagesWithModel(ctx context.Context, prompt string, images []string, model string, userID ...string) (*GuardrailResponse, error) {
	if len(images) == 0 {
		return nil, NewValidationError("images list cannot be empty")
	}

	// Build message content
	content := []interface{}{}
	if strings.TrimSpace(prompt) != "" {
		content = append(content, map[string]string{
			"type": "text",
			"text": strings.TrimSpace(prompt),
		})
	}

	// Encode all images
	for _, imagePath := range images {
		imageBase64, err := c.encodeBase64FromPath(imagePath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, NewValidationError(fmt.Sprintf("image file not found: %s", imagePath))
			}
			return nil, NewXiangxinAIError(fmt.Sprintf("failed to encode image %s: %v", imagePath, err), err)
		}

		content = append(content, map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64),
			},
		})
	}

	messages := []*Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	request := &GuardrailRequest{
		Model:    model,
		Messages: messages,
	}

	// Add optional userID parameter
	if len(userID) > 0 && userID[0] != "" {
		if request.ExtraBody == nil {
			request.ExtraBody = make(map[string]interface{})
		}
		request.ExtraBody["xxai_app_user_id"] = userID[0]
	}

	return c.makeRequest(ctx, "POST", "/guardrails", request)
}

// HealthCheck Check API service health status
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		Get("/guardrails/health")
	
	if err != nil {
		return nil, NewNetworkError("health check failed", err)
	}
	
	if resp.IsError() {
		return nil, c.handleErrorResponse(resp)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, NewXiangxinAIError("failed to parse response", err)
	}
	
	return result, nil
}

// GetModels Get available model list
func (c *Client) GetModels(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		Get("/guardrails/models")
	
	if err != nil {
		return nil, NewNetworkError("get models failed", err)
	}
	
	if resp.IsError() {
		return nil, c.handleErrorResponse(resp)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, NewXiangxinAIError("failed to parse response", err)
	}
	
	return result, nil
}

// makeRequest Send HTTP request
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, requestData *GuardrailRequest) (*GuardrailResponse, error) {
	return c.makeRequestWithData(ctx, method, endpoint, requestData)
}

// makeRequestWithData Send HTTP request (generic version)
func (c *Client) makeRequestWithData(ctx context.Context, method, endpoint string, requestData interface{}) (*GuardrailResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		resp, err := c.client.R().
			SetContext(ctx).
			SetBody(requestData).
			Post(endpoint)
		
		if err != nil {
			lastErr = NewNetworkError("request failed", err)
			if attempt < c.maxRetries {
				c.sleep(ctx, c.calculateBackoff(attempt))
				continue
			}
			return nil, lastErr
		}
		
		if resp.IsSuccess() {
			var result GuardrailResponse
			if err := json.Unmarshal(resp.Body(), &result); err != nil {
				return nil, NewXiangxinAIError("failed to parse response", err)
			}
			return &result, nil
		}
		
		// Handle HTTP error status code
		switch resp.StatusCode() {
		case 401:
			return nil, NewAuthenticationError("invalid API key")
		case 422:
			var errorResp map[string]interface{}
			json.Unmarshal(resp.Body(), &errorResp)
			detail := "validation error"
			if d, ok := errorResp["detail"]; ok {
				if s, ok := d.(string); ok {
					detail = s
				}
			}
			return nil, NewValidationError(fmt.Sprintf("validation error: %s", detail))
		case 429:
			if attempt < c.maxRetries {
				// Exponential backoff retry
				backoff := c.calculateBackoff(attempt)
				c.sleep(ctx, backoff)
				continue
			}
			return nil, NewRateLimitError("rate limit exceeded")
		default:
			errorMsg := string(resp.Body())
			var errorResp map[string]interface{}
			if json.Unmarshal(resp.Body(), &errorResp) == nil {
				if detail, ok := errorResp["detail"].(string); ok {
					errorMsg = detail
				}
			}
			lastErr = NewXiangxinAIError(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode(), errorMsg), nil)
			if attempt < c.maxRetries {
				c.sleep(ctx, c.calculateBackoff(attempt))
				continue
			}
			return nil, lastErr
		}
	}
	
	return nil, lastErr
}

// handleErrorResponse Handle error response
func (c *Client) handleErrorResponse(resp *resty.Response) error {
	switch resp.StatusCode() {
	case 401:
		return NewAuthenticationError("invalid API key")
	case 422:
		var errorResp map[string]interface{}
		json.Unmarshal(resp.Body(), &errorResp)
		detail := "validation error"
		if d, ok := errorResp["detail"]; ok {
			if s, ok := d.(string); ok {
				detail = s
			}
		}
		return NewValidationError(fmt.Sprintf("validation error: %s", detail))
	case 429:
		return NewRateLimitError("rate limit exceeded")
	default:
		errorMsg := string(resp.Body())
		var errorResp map[string]interface{}
		if json.Unmarshal(resp.Body(), &errorResp) == nil {
			if detail, ok := errorResp["detail"].(string); ok {
				errorMsg = detail
			}
		}
		return NewXiangxinAIError(fmt.Sprintf("API request failed with status %d: %s", resp.StatusCode(), errorMsg), nil)
	}
}

// calculateBackoff Calculate exponential backoff waiting time
func (c *Client) calculateBackoff(attempt int) time.Duration {
	base := time.Second
	backoff := time.Duration(math.Pow(2, float64(attempt))) * base
	return backoff + time.Second
}

// sleep Wait for specified time, support context cancellation
func (c *Client) sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}