package xiangxinai

import (
	"context"
	"sync"
)

// AsyncResult Async result structure
type AsyncResult[T any] struct {
	Result T
	Error  error
}

// AsyncClient Async client wrapper
// Provide asynchronous interfaces based on goroutines and channels, providing better performance and concurrency control
//
// Example usage:
//
//	asyncClient := xiangxinai.NewAsyncClient("your-api-key")
//	defer asyncClient.Close()
//	
//	// Async check prompt
//	resultChan := asyncClient.CheckPromptAsync(ctx, "User question")
//	select {
//	case result := <-resultChan:
//		if result.Error != nil {
//			log.Printf("Check prompt failed: %v", result.Error)
//		} else {
//			fmt.Printf("Check prompt result: %s\n", result.Result.OverallRiskLevel)
//		}
//	case <-ctx.Done():
//		fmt.Println("Check prompt timeout")
//	}
//	
//	// Batch async check
//	contents := []string{"Content 1", "Content 2", "Content 3"}
//	results := asyncClient.BatchCheckPrompts(ctx, contents)
//	for result := range results {
//		if result.Error != nil {
//			log.Printf("Batch check failed: %v", result.Error)
//		} else {
//			fmt.Printf("Batch check result: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
type AsyncClient struct {
	client     *Client
	workerPool chan struct{} // Worker pool, control concurrency
	wg         sync.WaitGroup
	closed     bool
	closeMu    sync.RWMutex
}

// NewAsyncClient Create new async client, using default configuration
func NewAsyncClient(apiKey string) *AsyncClient {
	return NewAsyncClientWithConfig(&ClientConfig{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
	}, 10) // Default concurrency is 10
}

// NewAsyncClientWithConfig Create new async client, using custom configuration
func NewAsyncClientWithConfig(config *ClientConfig, maxConcurrency int) *AsyncClient {
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}
	
	return &AsyncClient{
		client:     NewClientWithConfig(config),
		workerPool: make(chan struct{}, maxConcurrency),
		closed:     false,
	}
}

// CheckPromptAsync Async check prompt safety
//
// Parameters:
//   - ctx: Context
//   - content: Prompt content to check
//
// Return value:
//   - <-chan AsyncResult[*GuardrailResponse]: Async result channel
//
// Example:
//
//	resultChan := asyncClient.CheckPromptAsync(ctx, "I want to learn programming")
//	select {
//	case result := <-resultChan:
//		if result.Error != nil {
//			log.Printf("Check prompt failed: %v", result.Error)
//		} else {
//			fmt.Printf("Risk level: %s\n", result.Result.OverallRiskLevel)
//			fmt.Printf("Suggest action: %s\n", result.Result.SuggestAction)
//		}
//	case <-ctx.Done():
//		fmt.Println("Check prompt timeout or cancelled")
//	}
func (ac *AsyncClient) CheckPromptAsync(ctx context.Context, content string) <-chan AsyncResult[*GuardrailResponse] {
	return ac.CheckPromptWithModelAsync(ctx, content, DefaultModel)
}

// CheckPromptWithModelAsync Async check prompt safety, specify model
func (ac *AsyncClient) CheckPromptWithModelAsync(ctx context.Context, content, model string) <-chan AsyncResult[*GuardrailResponse] {
	resultChan := make(chan AsyncResult[*GuardrailResponse], 1)
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		resultChan <- AsyncResult[*GuardrailResponse]{Error: NewXiangxinAIError("async client is closed", nil)}
		close(resultChan)
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	ac.wg.Add(1)
	go func() {
		defer ac.wg.Done()
		defer close(resultChan)
		
		// Get worker slot
		select {
		case ac.workerPool <- struct{}{}:
			defer func() { <-ac.workerPool }()
		case <-ctx.Done():
			resultChan <- AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
			return
		}
		
		// Execute detection
		result, err := ac.client.CheckPromptWithModel(ctx, content, model)
		resultChan <- AsyncResult[*GuardrailResponse]{Result: result, Error: err}
	}()
	
	return resultChan
}

// CheckConversationAsync Async check conversation context safety - context-aware detection
//
// This is the core functionality of the guardrail, capable of understanding the complete conversation context for safety detection.
//
// Parameters:
//   - ctx: Context
//   - messages: Conversation message list
//
// Return value:
//   - <-chan AsyncResult[*GuardrailResponse]: Async result channel
//
// Example:
//
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "User question"),
//		xiangxinai.NewMessage("assistant", "Assistant answer"),
//	}
//	resultChan := asyncClient.CheckConversationAsync(ctx, messages)
//	result := <-resultChan
//	if result.Error != nil {
//		log.Printf("Check conversation failed: %v", result.Error)
//	} else {
//		fmt.Printf("Conversation risk level: %s\n", result.Result.OverallRiskLevel)
//	}
func (ac *AsyncClient) CheckConversationAsync(ctx context.Context, messages []*Message) <-chan AsyncResult[*GuardrailResponse] {
	return ac.CheckConversationWithModelAsync(ctx, messages, DefaultModel)
}

// CheckConversationWithModelAsync Async check conversation context safety, specify model
func (ac *AsyncClient) CheckConversationWithModelAsync(ctx context.Context, messages []*Message, model string) <-chan AsyncResult[*GuardrailResponse] {
	resultChan := make(chan AsyncResult[*GuardrailResponse], 1)
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		resultChan <- AsyncResult[*GuardrailResponse]{Error: NewXiangxinAIError("async client is closed", nil)}
		close(resultChan)
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	ac.wg.Add(1)
	go func() {
		defer ac.wg.Done()
		defer close(resultChan)
		
		// Get worker slot
		select {
		case ac.workerPool <- struct{}{}:
			defer func() { <-ac.workerPool }()
		case <-ctx.Done():
			resultChan <- AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
			return
		}
		
		// Execute detection
		result, err := ac.client.CheckConversationWithModel(ctx, messages, model)
		resultChan <- AsyncResult[*GuardrailResponse]{Result: result, Error: err}
	}()
	
	return resultChan
}

// BatchCheckPrompts Batch async check prompt
//
// Parameters:
//   - ctx: Context
//   - contents: Content list to check
//
// Return value:
//   - <-chan AsyncResult[*GuardrailResponse]: Async result channel, return results in order
//
// Example:
//
//	contents := []string{"Content 1", "Content 2", "Content 3"}
//	resultChan := asyncClient.BatchCheckPrompts(ctx, contents)
//	for result := range resultChan {
//		if result.Error != nil {
//			log.Printf("Batch check failed: %v", result.Error)
//		} else {
//			fmt.Printf("Batch check result: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
func (ac *AsyncClient) BatchCheckPrompts(ctx context.Context, contents []string) <-chan AsyncResult[*GuardrailResponse] {
	return ac.BatchCheckPromptsWithModel(ctx, contents, DefaultModel)
}

// BatchCheckPromptsWithModel Batch async check prompt, specify model
func (ac *AsyncClient) BatchCheckPromptsWithModel(ctx context.Context, contents []string, model string) <-chan AsyncResult[*GuardrailResponse] {
	resultChan := make(chan AsyncResult[*GuardrailResponse])
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		go func() {
			defer close(resultChan)
			for range contents {
				resultChan <- AsyncResult[*GuardrailResponse]{Error: NewXiangxinAIError("async client is closed", nil)}
			}
		}()
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	go func() {
		defer close(resultChan)
		
		// Create result collector, keep order
		results := make([]AsyncResult[*GuardrailResponse], len(contents))
		var wg sync.WaitGroup
		
		for i, content := range contents {
			wg.Add(1)
			go func(index int, content string) {
				defer wg.Done()
				
				// Get worker slot
				select {
				case ac.workerPool <- struct{}{}:
					defer func() { <-ac.workerPool }()
				case <-ctx.Done():
					results[index] = AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
					return
				}
				
				// Execute detection
				result, err := ac.client.CheckPromptWithModel(ctx, content, model)
				results[index] = AsyncResult[*GuardrailResponse]{Result: result, Error: err}
			}(i, content)
		}
		
		wg.Wait()
		
		// Send results in order
		for _, result := range results {
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return resultChan
}

// BatchCheckConversations Batch async check conversation
//
// Parameters:
//   - ctx: Context
//   - conversations: Conversation list
//
// Return value:
//   - <-chan AsyncResult[*GuardrailResponse]: Async result channel, return results in order
//
// Example:
//
//	conversations := [][]*xiangxinai.Message{
//		{xiangxinai.NewMessage("user", "Question 1")},
//		{xiangxinai.NewMessage("user", "Question 2")},
//	}
//	resultChan := asyncClient.BatchCheckConversations(ctx, conversations)
//	for result := range resultChan {
//		if result.Error != nil {
//			log.Printf("Batch check failed: %v", result.Error)
//		} else {
//			fmt.Printf("Batch check result: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
func (ac *AsyncClient) BatchCheckConversations(ctx context.Context, conversations [][]*Message) <-chan AsyncResult[*GuardrailResponse] {
	return ac.BatchCheckConversationsWithModel(ctx, conversations, DefaultModel)
}

// BatchCheckConversationsWithModel Batch async check conversation, specify model
func (ac *AsyncClient) BatchCheckConversationsWithModel(ctx context.Context, conversations [][]*Message, model string) <-chan AsyncResult[*GuardrailResponse] {
	resultChan := make(chan AsyncResult[*GuardrailResponse])
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		go func() {
			defer close(resultChan)
			for range conversations {
				resultChan <- AsyncResult[*GuardrailResponse]{Error: NewXiangxinAIError("async client is closed", nil)}
			}
		}()
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	go func() {
		defer close(resultChan)
		
		// Create result collector, keep order
		results := make([]AsyncResult[*GuardrailResponse], len(conversations))
		var wg sync.WaitGroup
		
		for i, messages := range conversations {
			wg.Add(1)
			go func(index int, messages []*Message) {
				defer wg.Done()
				
				// Get worker slot
				select {
				case ac.workerPool <- struct{}{}:
					defer func() { <-ac.workerPool }()
				case <-ctx.Done():
					results[index] = AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
					return
				}
				
				// Execute detection
				result, err := ac.client.CheckConversationWithModel(ctx, messages, model)
				results[index] = AsyncResult[*GuardrailResponse]{Result: result, Error: err}
			}(i, messages)
		}
		
		wg.Wait()
		
		// Send results in order
		for _, result := range results {
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return resultChan
}

// HealthCheckAsync Async check API service health status
func (ac *AsyncClient) HealthCheckAsync(ctx context.Context) <-chan AsyncResult[map[string]interface{}] {
	resultChan := make(chan AsyncResult[map[string]interface{}], 1)
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		resultChan <- AsyncResult[map[string]interface{}]{Error: NewXiangxinAIError("async client is closed", nil)}
		close(resultChan)
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	ac.wg.Add(1)
	go func() {
		defer ac.wg.Done()
		defer close(resultChan)
		
		result, err := ac.client.HealthCheck(ctx)
		resultChan <- AsyncResult[map[string]interface{}]{Result: result, Error: err}
	}()
	
	return resultChan
}

// GetModelsAsync Async get available model list
func (ac *AsyncClient) GetModelsAsync(ctx context.Context) <-chan AsyncResult[map[string]interface{}] {
	resultChan := make(chan AsyncResult[map[string]interface{}], 1)
	
	ac.closeMu.RLock()
	if ac.closed {
		ac.closeMu.RUnlock()
		resultChan <- AsyncResult[map[string]interface{}]{Error: NewXiangxinAIError("async client is closed", nil)}
		close(resultChan)
		return resultChan
	}
	ac.closeMu.RUnlock()
	
	ac.wg.Add(1)
	go func() {
		defer ac.wg.Done()
		defer close(resultChan)
		
		result, err := ac.client.GetModels(ctx)
		resultChan <- AsyncResult[map[string]interface{}]{Result: result, Error: err}
	}()
	
	return resultChan
}

// Close async client, wait for all ongoing operations to complete
func (ac *AsyncClient) Close() error {
	ac.closeMu.Lock()
	if ac.closed {
		ac.closeMu.Unlock()
		return nil
	}
	ac.closed = true
	ac.closeMu.Unlock()
	
	// Wait for all goroutines to complete
	ac.wg.Wait()
	
	// Close worker pool
	close(ac.workerPool)
	
	return nil
}

// GetConcurrency Get current concurrency limit
func (ac *AsyncClient) GetConcurrency() int {
	return cap(ac.workerPool)
}

// GetActiveWorkers Get current active worker count
func (ac *AsyncClient) GetActiveWorkers() int {
	return len(ac.workerPool)
}