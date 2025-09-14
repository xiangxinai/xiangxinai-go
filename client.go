package xiangxinai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// DefaultBaseURL 默认API基础URL
	DefaultBaseURL = "https://api.xiangxinai.cn/v1"
	// DefaultModel 默认模型名称
	DefaultModel = "Xiangxin-Guardrails-Text"
	// DefaultTimeout 默认请求超时时间（秒）
	DefaultTimeout = 30
	// DefaultMaxRetries 默认最大重试次数
	DefaultMaxRetries = 3
	// UserAgent 用户代理
	UserAgent = "xiangxinai-go/2.0.0"
)

// Client 象信AI安全护栏客户端 - 基于LLM的上下文感知AI安全护栏
//
// 这个客户端提供了与象信AI安全护栏API交互的简单接口。
// 护栏采用上下文感知技术，能够理解对话上下文进行安全检测。
//
// 示例用法:
//
//	client := xiangxinai.NewClient("your-api-key")
//	
//	// 检测用户输入
//	result, err := client.CheckPrompt(context.Background(), "用户问题")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 检测输出内容（基于上下文）
//	result, err := client.CheckResponseCtx(context.Background(), "用户问题", "助手回答")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 检测对话上下文
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "问题"),
//		xiangxinai.NewMessage("assistant", "回答"),
//	}
//	result, err := client.CheckConversation(context.Background(), messages)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "高风险/中风险/低风险/无风险"
//	fmt.Println(result.SuggestAction)    // "通过/阻断/代答"
type Client struct {
	client     *resty.Client
	maxRetries int
}

// NewClient 创建新的客户端，使用默认配置
func NewClient(apiKey string) *Client {
	return NewClientWithConfig(&ClientConfig{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
	})
}

// NewClientWithConfig 创建新的客户端，使用自定义配置
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

// createSafeResponse 创建无风险的默认响应
func (c *Client) createSafeResponse() *GuardrailResponse {
	return &GuardrailResponse{
		ID: "guardrails-safe-default",
		Result: &GuardrailResult{
			Compliance: &ComplianceResult{
				RiskLevel:  "无风险",
				Categories: []string{},
			},
			Security: &SecurityResult{
				RiskLevel:  "无风险",
				Categories: []string{},
			},
		},
		OverallRiskLevel: "无风险",
		SuggestAction:    "通过",
		SuggestAnswer:    nil,
	}
}

// CheckPrompt 检测用户输入的安全性
//
// 参数:
//   - ctx: 上下文
//   - content: 要检测的用户输入内容
//
// 返回值:
//   - *GuardrailResponse: 检测结果，格式为:
//     {
//       "id": "guardrails-xxx",
//       "result": {
//         "compliance": {
//           "risk_level": "高风险/中风险/低风险/无风险",
//           "categories": ["暴力犯罪", "敏感政治话题"]
//         },
//         "security": {
//           "risk_level": "高风险/中风险/低风险/无风险",
//           "categories": ["提示词攻击"]
//         }
//       },
//       "overall_risk_level": "高风险/中风险/低风险/无风险",
//       "suggest_action": "通过/阻断/代答",
//       "suggest_answer": "建议回答内容"
//     }
//   - error: 错误信息
//
// 可能的错误类型:
//   - ValidationError: 输入参数无效
//   - AuthenticationError: 认证失败
//   - RateLimitError: 超出速率限制
//   - XiangxinAIError: 其他API错误
//
// 示例:
//
//	result, err := client.CheckPrompt(ctx, "我想学习编程")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "无风险"
//	fmt.Println(result.SuggestAction)    // "通过"
//	fmt.Println(result.Result.Compliance.RiskLevel) // "无风险"
func (c *Client) CheckPrompt(ctx context.Context, content string) (*GuardrailResponse, error) {
	// 如果content是空字符串，直接返回无风险
	if strings.TrimSpace(content) == "" {
		return c.createSafeResponse(), nil
	}

	requestData := map[string]string{
		"input": strings.TrimSpace(content),
	}

	return c.makeRequestWithData(ctx, "POST", "/guardrails/input", requestData)
}

// CheckConversation 检测对话上下文的安全性 - 上下文感知检测
//
// 这是护栏的核心功能，能够理解完整的对话上下文进行安全检测。
// 不是分别检测每条消息，而是分析整个对话的安全性。
//
// 参数:
//   - ctx: 上下文
//   - messages: 对话消息列表，包含用户和助手的完整对话，每个消息包含role('user'或'assistant')和content
//
// 返回值:
//   - *GuardrailResponse: 基于对话上下文的检测结果，格式与CheckPrompt相同
//   - error: 错误信息
//
// 示例:
//
//	// 检测用户问题和助手回答的对话安全性
//	messages := []*xiangxinai.Message{
//		xiangxinai.NewMessage("user", "用户问题"),
//		xiangxinai.NewMessage("assistant", "助手回答"),
//	}
//	result, err := client.CheckConversation(ctx, messages)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "无风险"
//	fmt.Println(result.SuggestAction)    // 基于对话上下文的建议
func (c *Client) CheckConversation(ctx context.Context, messages []*Message) (*GuardrailResponse, error) {
	return c.CheckConversationWithModel(ctx, messages, DefaultModel)
}

// CheckConversationWithModel 检测对话上下文的安全性，指定模型
func (c *Client) CheckConversationWithModel(ctx context.Context, messages []*Message, model string) (*GuardrailResponse, error) {
	if len(messages) == 0 {
		return nil, NewValidationError("messages cannot be empty")
	}
	
	// 验证消息格式
	var validatedMessages []*Message
	allEmpty := true // 标记是否所有content都为空
	
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
		// 检查是否有非空content
		if content != "" {
			allEmpty = false
			// 只添加非空消息到validatedMessages
			validatedMessages = append(validatedMessages, &Message{
				Role:    msg.Role,
				Content: content,
			})
		}
	}
	
	// 如果所有messages的content都是空的，直接返回无风险
	if allEmpty {
		return c.createSafeResponse(), nil
	}
	
	// 确保至少有一条消息
	if len(validatedMessages) == 0 {
		return c.createSafeResponse(), nil
	}
	
	request := &GuardrailRequest{
		Model:    model,
		Messages: validatedMessages,
	}
	
	return c.makeRequest(ctx, "POST", "/guardrails", request)
}

// CheckResponseCtx 检测用户输入和模型输出的安全性 - 上下文感知检测
//
// 这是护栏的核心功能，能够理解用户输入和模型输出的上下文进行安全检测。
// 护栏会基于用户问题的上下文来检测模型输出是否安全合规。
//
// 参数:
//   - ctx: 上下文
//   - prompt: 用户输入的文本内容，用于让护栏理解上下文语意
//   - response: 模型输出的文本内容，实际检测对象
//
// 返回值:
//   - *GuardrailResponse: 基于上下文的检测结果，格式与CheckPrompt相同
//   - error: 错误信息
//
// 可能的错误类型:
//   - ValidationError: 输入参数无效
//   - AuthenticationError: 认证失败
//   - RateLimitError: 超出速率限制
//   - XiangxinAIError: 其他API错误
//
// 示例:
//
//	result, err := client.CheckResponseCtx(ctx, "教我做饭", "我可以教你做一些简单的家常菜")
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(result.OverallRiskLevel) // "无风险"
//	fmt.Println(result.SuggestAction)    // "通过"
func (c *Client) CheckResponseCtx(ctx context.Context, prompt, response string) (*GuardrailResponse, error) {
	// 如果prompt和response都是空字符串，直接返回无风险
	if strings.TrimSpace(prompt) == "" && strings.TrimSpace(response) == "" {
		return c.createSafeResponse(), nil
	}

	requestData := map[string]string{
		"input":  strings.TrimSpace(prompt),
		"output": strings.TrimSpace(response),
	}

	return c.makeRequestWithData(ctx, "POST", "/guardrails/output", requestData)
}

// HealthCheck 检查API服务健康状态
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

// GetModels 获取可用模型列表
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

// makeRequest 发送HTTP请求
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, requestData *GuardrailRequest) (*GuardrailResponse, error) {
	return c.makeRequestWithData(ctx, method, endpoint, requestData)
}

// makeRequestWithData 发送HTTP请求（通用版本）
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
		
		// 处理HTTP错误状态码
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
				// 指数退避重试
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

// handleErrorResponse 处理错误响应
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

// calculateBackoff 计算指数退避等待时间
func (c *Client) calculateBackoff(attempt int) time.Duration {
	base := time.Second
	backoff := time.Duration(math.Pow(2, float64(attempt))) * base
	return backoff + time.Second
}

// sleep 等待指定时间，支持上下文取消
func (c *Client) sleep(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(duration):
		return
	}
}