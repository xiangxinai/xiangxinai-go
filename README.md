# Xiangxin AI Guardrails Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go.svg)](https://pkg.go.dev/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)](https://goreportcard.com/report/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Xiangxin AI Guardrails Go Client - Context-aware AI guardrails based on LLM.

## Overview

An LLM-based context-aware AI guardrail that understands conversation context for security, safety and data leakage detection.

## Core Features

- 🧠 **Context-Aware** - LLM-based conversation understanding, not just simple batch detection
- 🔍 **Prompt Attack Detection** - Identify malicious prompt injection and jailbreak attacks
- 📋 **Content Compliance Detection** - Meet the basic security requirements for generative AI services
- 🔐 **Sensitive Data Leakage Prevention** - Detect and prevent personal/corporate sensitive data leaks
- 🧩 **User-Level Ban Policies** - Support risk identification and ban policies based on user granularity
- 🖼️ **Multimodal Detection** - Support image content safety detection
- 🛠️ **Easy Integration** - Compatible with OpenAI API format, one-line code integration
- ⚡ **OpenAI-style API** - Familiar interface design, quick to get started
- 🚀 **Sync/Async Support** - Support both synchronous and asynchronous calling methods to meet different scenario requirements

## Environment Requirements

- Go 1.18 or higher

## Installation

```bash
go get github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func main() {
    // Initialize client
    client := xiangxinai.NewClient("your-api-key")
    ctx := context.Background()

    // Check user input (optionally pass user ID)
    result, err := client.CheckPrompt(ctx, "User input question", "user-123")  // user-123 is optional
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.OverallRiskLevel) // no_risk/low_risk/medium_risk/high_risk
    fmt.Println(result.SuggestAction)     // pass/reject/replace
    fmt.Println(result.score)           // confidence score

    // Check output content (based on context)
    ctxResult, err := client.CheckResponseCtx(ctx, "Teach me how to cook", "I can teach you some simple home-cooked dishes", "user-123")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(ctxResult.OverallRiskLevel) // no_risk
    fmt.Println(ctxResult.SuggestAction)     // pass
}
```

### Conversation Context Detection (Recommended)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func main() {
    client := xiangxinai.NewClient("your-api-key")
    ctx := context.Background()

    // Check complete conversation context - core functionality
    messages := []*xiangxinai.Message{
        xiangxinai.NewMessage("user", "User's question"),
        xiangxinai.NewMessage("assistant", "AI assistant's answer"),
        xiangxinai.NewMessage("user", "User's follow-up question"),
    }

    result, err := client.CheckConversation(ctx, messages, "user-123")  // user-123 is optional
    if err != nil {
        log.Fatal(err)
    }

    // Check detection result
    if result.IsSafe() {
        fmt.Println("Conversation is safe, can continue")
    } else if result.IsBlocked() {
        fmt.Println("Conversation has risks, recommend blocking")
    } else if result.HasSubstitute() {
        fmt.Printf("Recommend using safe answer: %s\n", *result.SuggestAnswer)
    }
}
```

### Asynchronous Interface (Recommended, Better Performance)

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func main() {
    // Create async client
    asyncClient := xiangxinai.NewAsyncClient("your-api-key")
    defer asyncClient.Close() // Remember to close resources
    
    ctx := context.Background()
    
    // Asynchronously check single prompt
    resultChan := asyncClient.CheckPromptAsync(ctx, "User question")
    select {
    case result := <-resultChan:
        if result.Error != nil {
            log.Printf("Detection failed: %v", result.Error)
        } else {
            fmt.Printf("Async detection completed: %s\n", result.Result.OverallRiskLevel)
        }
    case <-time.After(5 * time.Second):
        fmt.Println("Detection timeout")
    }
    
    // Asynchronous conversation detection
    messages := []*xiangxinai.Message{
        xiangxinai.NewMessage("user", "User question"),
        xiangxinai.NewMessage("assistant", "Assistant answer"),
    }
    conversationChan := asyncClient.CheckConversationAsync(ctx, messages)
    result := <-conversationChan
    if result.Error != nil {
        log.Printf("Conversation detection failed: %v", result.Error)
    } else {
        fmt.Printf("Conversation detection completed: %s\n", result.Result.OverallRiskLevel)
    }
    
    // Batch async detection (high performance)
    contents := []string{"Content 1", "Content 2", "Content 3"}
    batchChan := asyncClient.BatchCheckPrompts(ctx, contents)
    for result := range batchChan {
        if result.Error != nil {
            log.Printf("Batch detection failed: %v", result.Error)
        } else {
            fmt.Printf("Batch detection result: %s\n", result.Result.OverallRiskLevel)
        }
    }
}
```

### Multimodal Image Detection

Supports multimodal detection functionality, supports image content safety detection, can combine prompt text semantics and image content semantic analysis to determine safety.

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func main() {
    client := xiangxinai.NewClient("your-api-key")
    ctx := context.Background()

    // Check single image (local file)
    result, err := client.CheckPromptImage(ctx, "Is this image safe?", "/path/to/image.jpg")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.OverallRiskLevel)
    fmt.Println(result.SuggestAction)

    // Check single image (network URL)
    result, err = client.CheckPromptImage(ctx, "", "https://example.com/image.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // Check multiple images
    images := []string{
        "/path/to/image1.jpg",
        "https://example.com/image2.jpg",
        "/path/to/image3.png",
    }
    result, err = client.CheckPromptImages(ctx, "Are all these images safe?", images)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.OverallRiskLevel)
}
```

### Custom Configuration

```go
// Synchronous client
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1", // Optional, default cloud service
    Timeout:    30,  // Request timeout (seconds), default 30
    MaxRetries: 3,   // Maximum retry count, default 3
}
client := xiangxinai.NewClientWithConfig(config)

// Async client (custom concurrency)
asyncClient := xiangxinai.NewAsyncClientWithConfig(config, 20) // Max concurrency 20
defer asyncClient.Close()
```

## API Reference

### Client (Synchronous Client)

#### Creating Client

```go
// Use default configuration
client := xiangxinai.NewClient("your-api-key")

// Use custom configuration
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1",
    Timeout:    30,
    MaxRetries: 3,
}
client := xiangxinai.NewClientWithConfig(config)
```

#### Methods

##### CheckPrompt(ctx, content)

Check safety of a single prompt.

```go
func (c *Client) CheckPrompt(ctx context.Context, content string) (*GuardrailResponse, error)
func (c *Client) CheckPromptWithModel(ctx context.Context, content, model string) (*GuardrailResponse, error)
```

**Parameters:**
- `ctx` (context.Context): Context
- `content` (string): Content to check
- `model` (string, optional): Model name, default "Xiangxin-Guardrails-Text"

##### CheckConversation(ctx, messages)

Check safety of conversation context (recommended).

```go
func (c *Client) CheckConversation(ctx context.Context, messages []*Message) (*GuardrailResponse, error)
func (c *Client) CheckConversationWithModel(ctx context.Context, messages []*Message, model string) (*GuardrailResponse, error)
```

**Parameters:**
- `ctx` (context.Context): Context
- `messages` ([]*Message): Conversation message list
- `model` (string, optional): Model name

##### HealthCheck(ctx)

Check API service health status.

```go
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error)
```

##### GetModels(ctx)

Get available model list.

```go
func (c *Client) GetModels(ctx context.Context) (map[string]interface{}, error)
```

### AsyncClient (Asynchronous Client, Recommended)

#### Creating Async Client

```go
// Use default configuration (concurrency 10)
asyncClient := xiangxinai.NewAsyncClient("your-api-key")
defer asyncClient.Close()

// Use custom configuration and concurrency
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1",
    Timeout:    30,
    MaxRetries: 3,
}
asyncClient := xiangxinai.NewAsyncClientWithConfig(config, 20) // Max concurrency 20
defer asyncClient.Close()
```

#### Async Methods

##### CheckPromptAsync(ctx, content)

Asynchronously check safety of a single prompt.

```go
func (ac *AsyncClient) CheckPromptAsync(ctx context.Context, content string) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) CheckPromptWithModelAsync(ctx context.Context, content, model string) <-chan AsyncResult[*GuardrailResponse]
```

**Return Value:**
- `<-chan AsyncResult[*GuardrailResponse]`: Asynchronous result channel

**Example:**
```go
resultChan := asyncClient.CheckPromptAsync(ctx, "User question")
select {
case result := <-resultChan:
    if result.Error != nil {
        log.Printf("Detection failed: %v", result.Error)
    } else {
        fmt.Printf("Detection completed: %s\n", result.Result.OverallRiskLevel)
    }
case <-ctx.Done():
    fmt.Println("Detection cancelled")
}
```

##### CheckConversationAsync(ctx, messages)

Asynchronously check safety of conversation context.

```go
func (ac *AsyncClient) CheckConversationAsync(ctx context.Context, messages []*Message) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) CheckConversationWithModelAsync(ctx context.Context, messages []*Message, model string) <-chan AsyncResult[*GuardrailResponse]
```

##### BatchCheckPrompts(ctx, contents)

Batch asynchronous prompt checking (high performance).

```go
func (ac *AsyncClient) BatchCheckPrompts(ctx context.Context, contents []string) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) BatchCheckPromptsWithModel(ctx context.Context, contents []string, model string) <-chan AsyncResult[*GuardrailResponse]
```

**Example:**
```go
contents := []string{"Content 1", "Content 2", "Content 3"}
resultChan := asyncClient.BatchCheckPrompts(ctx, contents)
for result := range resultChan {
    if result.Error != nil {
        log.Printf("Detection failed: %v", result.Error)
    } else {
        fmt.Printf("Batch detection result: %s\n", result.Result.OverallRiskLevel)
    }
}
```

##### BatchCheckConversations(ctx, conversations)

Batch asynchronous conversation checking.

```go
func (ac *AsyncClient) BatchCheckConversations(ctx context.Context, conversations [][]*Message) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) BatchCheckConversationsWithModel(ctx context.Context, conversations [][]*Message, model string) <-chan AsyncResult[*GuardrailResponse]
```

##### Concurrency Control Methods

```go
func (ac *AsyncClient) GetConcurrency() int        // Get concurrency limit
func (ac *AsyncClient) GetActiveWorkers() int      // Get current active worker count
func (ac *AsyncClient) Close() error               // Close async client
```

### Data Structures

#### Message

```go
type Message struct {
    Role    string `json:"role"`    // "user", "system", "assistant"
    Content string `json:"content"` // Message content
}

// Create new message
func NewMessage(role, content string) *Message
```

#### GuardrailResponse

```go
type GuardrailResponse struct {
    ID                string           `json:"id"`                  // Request unique identifier
    Result            *GuardrailResult `json:"result"`              // Detection result details
    OverallRiskLevel  string           `json:"overall_risk_level"`  // Overall risk level
    SuggestAction     string           `json:"suggest_action"`      // Suggested action
    SuggestAnswer     *string          `json:"suggest_answer"`      // Suggested answer
    Score             *float64         `json:"score"`               // Detection confidence score (added in v2.4.1)
}

// Convenience methods
func (r *GuardrailResponse) IsSafe() bool              // Check if safe
func (r *GuardrailResponse) IsBlocked() bool           // Check if blocked
func (r *GuardrailResponse) HasSubstitute() bool       // Check if has replace answer
func (r *GuardrailResponse) GetAllCategories() []string // Get all risk categories
```

#### GuardrailResult

```go
type GuardrailResult struct {
    Compliance *ComplianceResult `json:"compliance"` // Compliance detection result
    Security   *SecurityResult   `json:"security"`   // Security detection result
    Data       *DataResult       `json:"data"`       // Data leakage prevention detection result (added in v2.4.0)
}
```

#### ComplianceResult / SecurityResult / DataResult

```go
type ComplianceResult struct {
    RiskLevel  string   `json:"risk_level"`  // Risk level
    Categories []string `json:"categories"`  // Risk category list
}

type SecurityResult struct {
    RiskLevel  string   `json:"risk_level"`  // Risk level
    Categories []string `json:"categories"`  // Risk category list
}

type DataResult struct {
    RiskLevel  string   `json:"risk_level"`  // Risk level
    Categories []string `json:"categories"`  // Detected sensitive data types (added in v2.4.0)
}
```

### Response Format

```go
{
  "id": "guardrails-xxx",
  "result": {
    "compliance": {
      "risk_level": "no_risk",           // no_risk/low_risk/medium_risk/high_risk
      "categories": []                  // Compliance risk categories
    },
    "security": {
      "risk_level": "no_risk",           // no_risk/low_risk/medium_risk/high_risk
      "categories": []                  // Security risk categories
    },
    "data": {
      "risk_level": "no_risk",           // no_risk/low_risk/medium_risk/high_risk (added in v2.4.0)
      "categories": []                  // Detected sensitive data types (added in v2.4.0)
    }
  },
  "overall_risk_level": "no_risk",       // Overall risk level
  "suggest_action": "pass",             // pass/reject/replace
  "suggest_answer": null                // Suggested answer (contains desensitized content when data leakage prevention is triggered)
}
```

## Error Handling

```go
import (
    "errors"
    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

result, err := client.CheckPrompt(ctx, "test content")
if err != nil {
    var authErr *xiangxinai.AuthenticationError
    var rateErr *xiangxinai.RateLimitError
    var validationErr *xiangxinai.ValidationError
    var networkErr *xiangxinai.NetworkError
    
    switch {
    case errors.As(err, &authErr):
        fmt.Printf("Authentication failed, please check API key: %v\n", err)
    case errors.As(err, &rateErr):
        fmt.Printf("Request rate too high, please try again later: %v\n", err)
    case errors.As(err, &validationErr):
        fmt.Printf("Invalid input parameters: %v\n", err)
    case errors.As(err, &networkErr):
        fmt.Printf("Network connection error: %v\n", err)
    default:
        fmt.Printf("API error: %v\n", err)
    }
    return
}

fmt.Println(result)
```

### Error Types

- `XiangxinAIError` - Base error class
- `AuthenticationError` - Authentication failure
- `RateLimitError` - Rate limit exceeded
- `ValidationError` - Input validation error
- `NetworkError` - Network connection error
- `ServerError` - Server error

## Usage Scenarios

### 1. Content Moderation

```go
func moderateContent(client *xiangxinai.Client, userContent string) error {
    ctx := context.Background()
    result, err := client.CheckPrompt(ctx, userContent)
    if err != nil {
        return err
    }
    
    if !result.IsSafe() {
        categories := result.GetAllCategories()
        fmt.Printf("Content contains risks: %v\n", categories)
        return fmt.Errorf("content moderation failed: %s", result.OverallRiskLevel)
    }
    
    return nil
}
```

### 2. Chat System Protection

```go
func safeChatResponse(client *xiangxinai.Client, conversation []*xiangxinai.Message) (string, error) {
    ctx := context.Background()
    result, err := client.CheckConversation(ctx, conversation)
    if err != nil {
        return "", err
    }
    
    if result.SuggestAction == "replace" && result.SuggestAnswer != nil {
        // Use safe replace answer
        return *result.SuggestAnswer, nil
    } else if result.IsBlocked() {
        // Block unsafe conversation
        return "Sorry, I cannot answer this question", nil
    }
    
    // Conversation is safe, continue normal process
    return "", nil
}
```

### 3. Middleware Integration

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func GuardrailMiddleware(client *xiangxinai.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req struct {
            Content string `json:"content"`
        }
        
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            c.Abort()
            return
        }
        
        result, err := client.CheckPrompt(context.Background(), req.Content)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            c.Abort()
            return
        }
        
        if result.IsBlocked() {
            c.JSON(http.StatusForbidden, gin.H{
                "error": "Content blocked",
                "risk_level": result.OverallRiskLevel,
                "categories": result.GetAllCategories(),
            })
            c.Abort()
            return
        }
        
        c.Set("guardrail_result", result)
        c.Next()
    }
}
```

### 4. Concurrent Detection

```go
package main

import (
    "context"
    "fmt"
    "sync"
    
    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func batchCheck(client *xiangxinai.Client, contents []string) {
    var wg sync.WaitGroup
    results := make(chan *xiangxinai.GuardrailResponse, len(contents))
    
    for _, content := range contents {
        wg.Add(1)
        go func(content string) {
            defer wg.Done()
            
            result, err := client.CheckPrompt(context.Background(), content)
            if err != nil {
                fmt.Printf("Detection failed: %v\n", err)
                return
            }
            
            results <- result
        }(content)
    }
    
    wg.Wait()
    close(results)
    
    for result := range results {
        fmt.Printf("Content: %s, Risk Level: %s\n", 
            result.ID, result.OverallRiskLevel)
    }
}
```

### 5. Context Cancellation

```go
func checkWithTimeout(client *xiangxinai.Client, content string, timeout time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    result, err := client.CheckPrompt(ctx, content)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("Detection timeout")
        } else {
            fmt.Printf("Detection failed: %v\n", err)
        }
        return
    }
    
    fmt.Printf("Detection result: %s\n", result.SuggestAction)
}
```

## Best Practices

1. **Use Conversation Context Detection**: Recommend using `CheckConversation` instead of `CheckPrompt`, as context awareness provides more accurate detection results.

2. **Context Management**: Properly use `context.Context` for timeout control and cancellation operations.

3. **Error Handling**: Implement appropriate error handling and retry mechanisms.

4. **Client Reuse**: Reuse the same `Client` instance in your application, avoid frequent creation.

5. **Concurrency Safety**: `Client` is concurrency-safe and can be used simultaneously in multiple goroutines.

6. **Resource Management**: `Client` internally uses connection pooling and typically doesn't require manual closing.

## Performance Considerations

- Default configuration optimized for most usage scenarios
- Supports connection reuse and keep-alive
- Automatic retry and exponential backoff
- Context cancellation support

## License

Apache 2.0

## Technical Support

- Website: https://xiangxinai.cn
- Documentation: https://docs.xiangxinai.cn
- Issue Reporting: https://github.com/xiangxinai/xiangxin-guardrails/issues
- Email: wanglei@xiangxinai.cn

## Contribution Guide

Welcome to submit Issues and Pull Requests!

## Changelog
### v2.0.0
- Added check_response_ctx(prompt, response) interface, to be used with check_prompt(prompt) for convenience.

### v1.1.1
- Adjusted maximum detection content length from 10000 to 1M

### v1.1.0
- Initial version release
- Support prompt detection and conversation context detection
- Complete error handling and retry mechanism
- Concurrency-safe client implementation