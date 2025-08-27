package xiangxinai

import (
	"context"
	"sync"
)

// AsyncResult 异步结果结构
type AsyncResult[T any] struct {
	Result T
	Error  error
}

// AsyncClient 异步客户端包装器
// 提供基于goroutine和channel的异步接口，提供更好的性能和并发控制
//
// 示例用法:
//
//	asyncClient := xiangxinai.NewAsyncClient("your-api-key")
//	defer asyncClient.Close()
//	
//	// 异步检测提示词
//	resultChan := asyncClient.CheckPromptAsync(ctx, "用户问题")
//	select {
//	case result := <-resultChan:
//		if result.Error != nil {
//			log.Printf("检测失败: %v", result.Error)
//		} else {
//			fmt.Printf("检测结果: %s\n", result.Result.OverallRiskLevel)
//		}
//	case <-ctx.Done():
//		fmt.Println("检测超时")
//	}
//	
//	// 批量异步检测
//	contents := []string{"内容1", "内容2", "内容3"}
//	results := asyncClient.BatchCheckPrompts(ctx, contents)
//	for result := range results {
//		if result.Error != nil {
//			log.Printf("检测失败: %v", result.Error)
//		} else {
//			fmt.Printf("批量检测结果: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
type AsyncClient struct {
	client     *Client
	workerPool chan struct{} // 工作池，控制并发数
	wg         sync.WaitGroup
	closed     bool
	closeMu    sync.RWMutex
}

// NewAsyncClient 创建新的异步客户端，使用默认配置
func NewAsyncClient(apiKey string) *AsyncClient {
	return NewAsyncClientWithConfig(&ClientConfig{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
	}, 10) // 默认并发数为10
}

// NewAsyncClientWithConfig 创建新的异步客户端，使用自定义配置
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

// CheckPromptAsync 异步检测提示词的安全性
//
// 参数:
//   - ctx: 上下文
//   - content: 要检测的提示词内容
//
// 返回值:
//   - <-chan AsyncResult[*GuardrailResponse]: 异步结果通道
//
// 示例:
//
//	resultChan := asyncClient.CheckPromptAsync(ctx, "我想学习编程")
//	select {
//	case result := <-resultChan:
//		if result.Error != nil {
//			log.Printf("检测失败: %v", result.Error)
//		} else {
//			fmt.Printf("风险等级: %s\n", result.Result.OverallRiskLevel)
//			fmt.Printf("建议动作: %s\n", result.Result.SuggestAction)
//		}
//	case <-ctx.Done():
//		fmt.Println("检测超时或被取消")
//	}
func (ac *AsyncClient) CheckPromptAsync(ctx context.Context, content string) <-chan AsyncResult[*GuardrailResponse] {
	return ac.CheckPromptWithModelAsync(ctx, content, DefaultModel)
}

// CheckPromptWithModelAsync 异步检测提示词的安全性，指定模型
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
		
		// 获取工作槽
		select {
		case ac.workerPool <- struct{}{}:
			defer func() { <-ac.workerPool }()
		case <-ctx.Done():
			resultChan <- AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
			return
		}
		
		// 执行检测
		result, err := ac.client.CheckPromptWithModel(ctx, content, model)
		resultChan <- AsyncResult[*GuardrailResponse]{Result: result, Error: err}
	}()
	
	return resultChan
}

// CheckConversationAsync 异步检测对话上下文的安全性 - 上下文感知检测
//
// 这是护栏的核心功能，能够理解完整的对话上下文进行安全检测。
//
// 参数:
//   - ctx: 上下文
//   - messages: 对话消息列表
//
// 返回值:
//   - <-chan AsyncResult[*GuardrailResponse]: 异步结果通道
//
// 示例:
//
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "用户问题"),
//		xiangxinai.NewMessage("assistant", "助手回答"),
//	}
//	resultChan := asyncClient.CheckConversationAsync(ctx, messages)
//	result := <-resultChan
//	if result.Error != nil {
//		log.Printf("检测失败: %v", result.Error)
//	} else {
//		fmt.Printf("对话风险等级: %s\n", result.Result.OverallRiskLevel)
//	}
func (ac *AsyncClient) CheckConversationAsync(ctx context.Context, messages []*Message) <-chan AsyncResult[*GuardrailResponse] {
	return ac.CheckConversationWithModelAsync(ctx, messages, DefaultModel)
}

// CheckConversationWithModelAsync 异步检测对话上下文的安全性，指定模型
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
		
		// 获取工作槽
		select {
		case ac.workerPool <- struct{}{}:
			defer func() { <-ac.workerPool }()
		case <-ctx.Done():
			resultChan <- AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
			return
		}
		
		// 执行检测
		result, err := ac.client.CheckConversationWithModel(ctx, messages, model)
		resultChan <- AsyncResult[*GuardrailResponse]{Result: result, Error: err}
	}()
	
	return resultChan
}

// BatchCheckPrompts 批量异步检测提示词
//
// 参数:
//   - ctx: 上下文
//   - contents: 要检测的内容列表
//
// 返回值:
//   - <-chan AsyncResult[*GuardrailResponse]: 异步结果通道，按顺序返回结果
//
// 示例:
//
//	contents := []string{"内容1", "内容2", "内容3"}
//	resultChan := asyncClient.BatchCheckPrompts(ctx, contents)
//	for result := range resultChan {
//		if result.Error != nil {
//			log.Printf("检测失败: %v", result.Error)
//		} else {
//			fmt.Printf("批量检测结果: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
func (ac *AsyncClient) BatchCheckPrompts(ctx context.Context, contents []string) <-chan AsyncResult[*GuardrailResponse] {
	return ac.BatchCheckPromptsWithModel(ctx, contents, DefaultModel)
}

// BatchCheckPromptsWithModel 批量异步检测提示词，指定模型
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
		
		// 创建结果收集器，保持顺序
		results := make([]AsyncResult[*GuardrailResponse], len(contents))
		var wg sync.WaitGroup
		
		for i, content := range contents {
			wg.Add(1)
			go func(index int, content string) {
				defer wg.Done()
				
				// 获取工作槽
				select {
				case ac.workerPool <- struct{}{}:
					defer func() { <-ac.workerPool }()
				case <-ctx.Done():
					results[index] = AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
					return
				}
				
				// 执行检测
				result, err := ac.client.CheckPromptWithModel(ctx, content, model)
				results[index] = AsyncResult[*GuardrailResponse]{Result: result, Error: err}
			}(i, content)
		}
		
		wg.Wait()
		
		// 按顺序发送结果
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

// BatchCheckConversations 批量异步检测对话
//
// 参数:
//   - ctx: 上下文
//   - conversations: 对话列表
//
// 返回值:
//   - <-chan AsyncResult[*GuardrailResponse]: 异步结果通道，按顺序返回结果
//
// 示例:
//
//	conversations := [][]*xiangxinai.Message{
//		{xiangxinai.NewMessage("user", "问题1")},
//		{xiangxinai.NewMessage("user", "问题2")},
//	}
//	resultChan := asyncClient.BatchCheckConversations(ctx, conversations)
//	for result := range resultChan {
//		if result.Error != nil {
//			log.Printf("检测失败: %v", result.Error)
//		} else {
//			fmt.Printf("批量对话检测结果: %s\n", result.Result.OverallRiskLevel)
//		}
//	}
func (ac *AsyncClient) BatchCheckConversations(ctx context.Context, conversations [][]*Message) <-chan AsyncResult[*GuardrailResponse] {
	return ac.BatchCheckConversationsWithModel(ctx, conversations, DefaultModel)
}

// BatchCheckConversationsWithModel 批量异步检测对话，指定模型
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
		
		// 创建结果收集器，保持顺序
		results := make([]AsyncResult[*GuardrailResponse], len(conversations))
		var wg sync.WaitGroup
		
		for i, messages := range conversations {
			wg.Add(1)
			go func(index int, messages []*Message) {
				defer wg.Done()
				
				// 获取工作槽
				select {
				case ac.workerPool <- struct{}{}:
					defer func() { <-ac.workerPool }()
				case <-ctx.Done():
					results[index] = AsyncResult[*GuardrailResponse]{Error: ctx.Err()}
					return
				}
				
				// 执行检测
				result, err := ac.client.CheckConversationWithModel(ctx, messages, model)
				results[index] = AsyncResult[*GuardrailResponse]{Result: result, Error: err}
			}(i, messages)
		}
		
		wg.Wait()
		
		// 按顺序发送结果
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

// HealthCheckAsync 异步检查API服务健康状态
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

// GetModelsAsync 异步获取可用模型列表
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

// Close 关闭异步客户端，等待所有正在进行的操作完成
func (ac *AsyncClient) Close() error {
	ac.closeMu.Lock()
	if ac.closed {
		ac.closeMu.Unlock()
		return nil
	}
	ac.closed = true
	ac.closeMu.Unlock()
	
	// 等待所有goroutine完成
	ac.wg.Wait()
	
	// 关闭工作池
	close(ac.workerPool)
	
	return nil
}

// GetConcurrency 获取当前并发数限制
func (ac *AsyncClient) GetConcurrency() int {
	return cap(ac.workerPool)
}

// GetActiveWorkers 获取当前活跃的工作线程数
func (ac *AsyncClient) GetActiveWorkers() int {
	return len(ac.workerPool)
}