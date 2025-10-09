# 象信AI安全护栏 Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go.svg)](https://pkg.go.dev/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)](https://goreportcard.com/report/github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

象信AI安全护栏 Go 客户端 - 基于LLM的上下文感知AI安全护栏。

## 概述

象信AI安全护栏是一个基于大语言模型的上下文感知AI安全护栏系统，能够理解对话上下文进行智能安全检测。不同于传统的关键词匹配，我们的护栏能够理解语言的深层含义和对话的上下文关系。

## 核心特性

- 🧠 **上下文感知** - 基于LLM的对话理解，而不是简单的批量检测
- 🔍 **提示词攻击检测** - 识别恶意提示词注入和越狱攻击
- 📋 **内容合规检测** - 满足生成式人工智能服务安全基本要求
- 🔐 **敏感数据防泄漏** - 检测和防止个人/企业敏感数据泄露
- 🧩 **用户级封禁策略** - 支持基于用户颗粒度的风险识别与封禁策略
- 🖼️ **多模态检测** - 支持图片内容安全检测
- 🛠️ **易于集成** - 兼容OpenAI API格式，一行代码接入
- ⚡ **OpenAI风格API** - 熟悉的接口设计，快速上手
- 🚀 **同步/异步支持** - 支持同步和异步两种调用方式，满足不同场景需求

## 环境要求

- Go 1.18 或更高版本

## 安装

```bash
go get github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go
```

## 快速开始

### 基本用法

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

func main() {
    // 初始化客户端
    client := xiangxinai.NewClient("your-api-key")
    ctx := context.Background()

    // 检测用户输入（可选传入用户ID）
    result, err := client.CheckPrompt(ctx, "用户输入的问题", "user-123")  // user-123是可选参数
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(result.OverallRiskLevel) // 无风险/低风险/中风险/高风险
    fmt.Println(result.SuggestAction)     // 通过/阻断/代答

    // 检测输出内容（基于上下文）
    ctxResult, err := client.CheckResponseCtx(ctx, "教我做饭", "我可以教你做一些简单的家常菜", "user-123")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(ctxResult.OverallRiskLevel) // 无风险
    fmt.Println(ctxResult.SuggestAction)     // 通过
}
```

### 对话上下文检测（推荐）

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

    // 检测完整对话上下文 - 核心功能
    messages := []*xiangxinai.Message{
        xiangxinai.NewMessage("user", "用户的问题"),
        xiangxinai.NewMessage("assistant", "AI助手的回答"),
        xiangxinai.NewMessage("user", "用户的后续问题"),
    }

    result, err := client.CheckConversation(ctx, messages, "user-123")  // user-123是可选参数
    if err != nil {
        log.Fatal(err)
    }

    // 检查检测结果
    if result.IsSafe() {
        fmt.Println("对话安全，可以继续")
    } else if result.IsBlocked() {
        fmt.Println("对话存在风险，建议阻断")
    } else if result.HasSubstitute() {
        fmt.Printf("建议使用安全回答: %s\n", *result.SuggestAnswer)
    }
}
```

### 异步接口（推荐，性能更好）

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
    // 创建异步客户端
    asyncClient := xiangxinai.NewAsyncClient("your-api-key")
    defer asyncClient.Close() // 记住关闭资源
    
    ctx := context.Background()
    
    // 异步检测单个提示词
    resultChan := asyncClient.CheckPromptAsync(ctx, "用户问题")
    select {
    case result := <-resultChan:
        if result.Error != nil {
            log.Printf("检测失败: %v", result.Error)
        } else {
            fmt.Printf("异步检测完成: %s\n", result.Result.OverallRiskLevel)
        }
    case <-time.After(5 * time.Second):
        fmt.Println("检测超时")
    }
    
    // 异步对话检测
    messages := []*xiangxinai.Message{
        xiangxinai.NewMessage("user", "用户问题"),
        xiangxinai.NewMessage("assistant", "助手回答"),
    }
    conversationChan := asyncClient.CheckConversationAsync(ctx, messages)
    result := <-conversationChan
    if result.Error != nil {
        log.Printf("对话检测失败: %v", result.Error)
    } else {
        fmt.Printf("对话检测完成: %s\n", result.Result.OverallRiskLevel)
    }
    
    // 批量异步检测（高性能）
    contents := []string{"内容1", "内容2", "内容3"}
    batchChan := asyncClient.BatchCheckPrompts(ctx, contents)
    for result := range batchChan {
        if result.Error != nil {
            log.Printf("批量检测失败: %v", result.Error)
        } else {
            fmt.Printf("批量检测结果: %s\n", result.Result.OverallRiskLevel)
        }
    }
}
```

### 多模态图片检测

支持多模态检测功能，支持图片内容安全检测，可以结合提示词文本的语义和图片内容语义分析得出是否安全。

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

    // 检测单张图片（本地文件）
    result, err := client.CheckPromptImage(ctx, "这个图片安全吗？", "/path/to/image.jpg")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.OverallRiskLevel)
    fmt.Println(result.SuggestAction)

    // 检测单张图片（网络URL）
    result, err = client.CheckPromptImage(ctx, "", "https://example.com/image.jpg")
    if err != nil {
        log.Fatal(err)
    }

    // 检测多张图片
    images := []string{
        "/path/to/image1.jpg",
        "https://example.com/image2.jpg",
        "/path/to/image3.png",
    }
    result, err = client.CheckPromptImages(ctx, "这些图片都安全吗？", images)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.OverallRiskLevel)
}
```

### 自定义配置

```go
// 同步客户端
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1", // 可选，默认云端服务
    Timeout:    30,  // 请求超时时间（秒），默认30
    MaxRetries: 3,   // 最大重试次数，默认3
}
client := xiangxinai.NewClientWithConfig(config)

// 异步客户端（自定义并发数）
asyncClient := xiangxinai.NewAsyncClientWithConfig(config, 20) // 最大并发数20
defer asyncClient.Close()
```

## API 参考

### Client（同步客户端）

#### 创建客户端

```go
// 使用默认配置
client := xiangxinai.NewClient("your-api-key")

// 使用自定义配置
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1",
    Timeout:    30,
    MaxRetries: 3,
}
client := xiangxinai.NewClientWithConfig(config)
```

#### 方法

##### CheckPrompt(ctx, content)

检测单个提示词的安全性。

```go
func (c *Client) CheckPrompt(ctx context.Context, content string) (*GuardrailResponse, error)
func (c *Client) CheckPromptWithModel(ctx context.Context, content, model string) (*GuardrailResponse, error)
```

**参数:**
- `ctx` (context.Context): 上下文
- `content` (string): 要检测的内容
- `model` (string, 可选): 模型名称，默认 "Xiangxin-Guardrails-Text"

##### CheckConversation(ctx, messages)

检测对话上下文的安全性（推荐使用）。

```go
func (c *Client) CheckConversation(ctx context.Context, messages []*Message) (*GuardrailResponse, error)
func (c *Client) CheckConversationWithModel(ctx context.Context, messages []*Message, model string) (*GuardrailResponse, error)
```

**参数:**
- `ctx` (context.Context): 上下文
- `messages` ([]*Message): 对话消息列表
- `model` (string, 可选): 模型名称

##### HealthCheck(ctx)

检查API服务健康状态。

```go
func (c *Client) HealthCheck(ctx context.Context) (map[string]interface{}, error)
```

##### GetModels(ctx)

获取可用模型列表。

```go
func (c *Client) GetModels(ctx context.Context) (map[string]interface{}, error)
```

### AsyncClient（异步客户端，推荐）

#### 创建异步客户端

```go
// 使用默认配置（并发数10）
asyncClient := xiangxinai.NewAsyncClient("your-api-key")
defer asyncClient.Close()

// 使用自定义配置和并发数
config := &xiangxinai.ClientConfig{
    APIKey:     "your-api-key",
    BaseURL:    "https://api.xiangxinai.cn/v1",
    Timeout:    30,
    MaxRetries: 3,
}
asyncClient := xiangxinai.NewAsyncClientWithConfig(config, 20) // 最大并发数20
defer asyncClient.Close()
```

#### 异步方法

##### CheckPromptAsync(ctx, content)

异步检测单个提示词的安全性。

```go
func (ac *AsyncClient) CheckPromptAsync(ctx context.Context, content string) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) CheckPromptWithModelAsync(ctx context.Context, content, model string) <-chan AsyncResult[*GuardrailResponse]
```

**返回值:**
- `<-chan AsyncResult[*GuardrailResponse]`: 异步结果通道

**示例:**
```go
resultChan := asyncClient.CheckPromptAsync(ctx, "用户问题")
select {
case result := <-resultChan:
    if result.Error != nil {
        log.Printf("检测失败: %v", result.Error)
    } else {
        fmt.Printf("检测完成: %s\n", result.Result.OverallRiskLevel)
    }
case <-ctx.Done():
    fmt.Println("检测被取消")
}
```

##### CheckConversationAsync(ctx, messages)

异步检测对话上下文的安全性。

```go
func (ac *AsyncClient) CheckConversationAsync(ctx context.Context, messages []*Message) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) CheckConversationWithModelAsync(ctx context.Context, messages []*Message, model string) <-chan AsyncResult[*GuardrailResponse]
```

##### BatchCheckPrompts(ctx, contents)

批量异步检测提示词（高性能）。

```go
func (ac *AsyncClient) BatchCheckPrompts(ctx context.Context, contents []string) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) BatchCheckPromptsWithModel(ctx context.Context, contents []string, model string) <-chan AsyncResult[*GuardrailResponse]
```

**示例:**
```go
contents := []string{"内容1", "内容2", "内容3"}
resultChan := asyncClient.BatchCheckPrompts(ctx, contents)
for result := range resultChan {
    if result.Error != nil {
        log.Printf("检测失败: %v", result.Error)
    } else {
        fmt.Printf("批量检测结果: %s\n", result.Result.OverallRiskLevel)
    }
}
```

##### BatchCheckConversations(ctx, conversations)

批量异步检测对话。

```go
func (ac *AsyncClient) BatchCheckConversations(ctx context.Context, conversations [][]*Message) <-chan AsyncResult[*GuardrailResponse]
func (ac *AsyncClient) BatchCheckConversationsWithModel(ctx context.Context, conversations [][]*Message, model string) <-chan AsyncResult[*GuardrailResponse]
```

##### 并发控制方法

```go
func (ac *AsyncClient) GetConcurrency() int        // 获取并发数限制
func (ac *AsyncClient) GetActiveWorkers() int      // 获取当前活跃工作线程数
func (ac *AsyncClient) Close() error               // 关闭异步客户端
```

### 数据结构

#### Message

```go
type Message struct {
    Role    string `json:"role"`    // "user", "system", "assistant"
    Content string `json:"content"` // 消息内容
}

// 创建新消息
func NewMessage(role, content string) *Message
```

#### GuardrailResponse

```go
type GuardrailResponse struct {
    ID                string           `json:"id"`                  // 请求唯一标识
    Result            *GuardrailResult `json:"result"`              // 检测结果详情
    OverallRiskLevel  string           `json:"overall_risk_level"`  // 综合风险等级
    SuggestAction     string           `json:"suggest_action"`      // 建议动作
    SuggestAnswer     *string          `json:"suggest_answer"`      // 建议回答
    Score             *float64         `json:"score"`               // 检测置信度分数 (v2.4.1新增)
}

// 便捷方法
func (r *GuardrailResponse) IsSafe() bool              // 判断是否安全
func (r *GuardrailResponse) IsBlocked() bool           // 判断是否被阻断
func (r *GuardrailResponse) HasSubstitute() bool       // 判断是否有代答
func (r *GuardrailResponse) GetAllCategories() []string // 获取所有风险类别
```

#### GuardrailResult

```go
type GuardrailResult struct {
    Compliance *ComplianceResult `json:"compliance"` // 合规检测结果
    Security   *SecurityResult   `json:"security"`   // 安全检测结果
    Data       *DataResult       `json:"data"`       // 数据防泄漏检测结果（v2.4.0新增）
}
```

#### ComplianceResult / SecurityResult / DataResult

```go
type ComplianceResult struct {
    RiskLevel  string   `json:"risk_level"`  // 风险等级
    Categories []string `json:"categories"`  // 风险类别列表
}

type SecurityResult struct {
    RiskLevel  string   `json:"risk_level"`  // 风险等级
    Categories []string `json:"categories"`  // 风险类别列表
}

type DataResult struct {
    RiskLevel  string   `json:"risk_level"`  // 风险等级
    Categories []string `json:"categories"`  // 检测到的敏感数据类型（v2.4.0新增）
}
```

### 响应格式

```go
{
  "id": "guardrails-xxx",
  "result": {
    "compliance": {
      "risk_level": "无风险",           // 无风险/低风险/中风险/高风险
      "categories": []                  // 合规风险类别
    },
    "security": {
      "risk_level": "无风险",           // 无风险/低风险/中风险/高风险
      "categories": []                  // 安全风险类别
    },
    "data": {
      "risk_level": "无风险",           // 无风险/低风险/中风险/高风险（v2.4.0新增）
      "categories": []                  // 检测到的敏感数据类型（v2.4.0新增）
    }
  },
  "overall_risk_level": "无风险",       // 综合风险等级
  "suggest_action": "通过",             // 通过/阻断/代答
  "suggest_answer": null                // 建议回答（数据防泄漏时包含脱敏后内容）
}
```

## 错误处理

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
        fmt.Printf("认证失败，请检查API密钥: %v\n", err)
    case errors.As(err, &rateErr):
        fmt.Printf("请求频率过高，请稍后重试: %v\n", err)
    case errors.As(err, &validationErr):
        fmt.Printf("输入参数无效: %v\n", err)
    case errors.As(err, &networkErr):
        fmt.Printf("网络连接错误: %v\n", err)
    default:
        fmt.Printf("API错误: %v\n", err)
    }
    return
}

fmt.Println(result)
```

### 错误类型

- `XiangxinAIError` - 基础错误类
- `AuthenticationError` - 认证失败
- `RateLimitError` - 超出速率限制
- `ValidationError` - 输入验证错误
- `NetworkError` - 网络连接错误
- `ServerError` - 服务器错误

## 使用场景

### 1. 内容审核

```go
func moderateContent(client *xiangxinai.Client, userContent string) error {
    ctx := context.Background()
    result, err := client.CheckPrompt(ctx, userContent)
    if err != nil {
        return err
    }
    
    if !result.IsSafe() {
        categories := result.GetAllCategories()
        fmt.Printf("内容包含风险: %v\n", categories)
        return fmt.Errorf("content moderation failed: %s", result.OverallRiskLevel)
    }
    
    return nil
}
```

### 2. 对话系统防护

```go
func safeChatResponse(client *xiangxinai.Client, conversation []*xiangxinai.Message) (string, error) {
    ctx := context.Background()
    result, err := client.CheckConversation(ctx, conversation)
    if err != nil {
        return "", err
    }
    
    if result.SuggestAction == "代答" && result.SuggestAnswer != nil {
        // 使用安全的代答内容
        return *result.SuggestAnswer, nil
    } else if result.IsBlocked() {
        // 阻断不安全的对话
        return "抱歉，我无法回答这个问题", nil
    }
    
    // 对话安全，继续正常流程
    return "", nil
}
```

### 3. 中间件集成

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

### 4. 并发检测

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
                fmt.Printf("检测失败: %v\n", err)
                return
            }
            
            results <- result
        }(content)
    }
    
    wg.Wait()
    close(results)
    
    for result := range results {
        fmt.Printf("内容: %s, 风险等级: %s\n", 
            result.ID, result.OverallRiskLevel)
    }
}
```

### 5. 上下文取消

```go
func checkWithTimeout(client *xiangxinai.Client, content string, timeout time.Duration) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    result, err := client.CheckPrompt(ctx, content)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            fmt.Println("检测超时")
        } else {
            fmt.Printf("检测失败: %v\n", err)
        }
        return
    }
    
    fmt.Printf("检测结果: %s\n", result.SuggestAction)
}
```

## 最佳实践

1. **使用对话上下文检测**: 推荐使用 `CheckConversation` 而不是 `CheckPrompt`，因为上下文感知能提供更准确的检测结果。

2. **上下文管理**: 合理使用 `context.Context` 进行超时控制和取消操作。

3. **错误处理**: 实现适当的错误处理和重试机制。

4. **客户端复用**: 在应用中复用同一个 `Client` 实例，避免频繁创建。

5. **并发安全**: `Client` 是并发安全的，可以在多个 goroutine 中同时使用。

6. **资源管理**: `Client` 内部使用连接池，通常不需要手动关闭。

## 性能考虑

- 默认配置已针对大多数使用场景优化
- 支持连接复用和keep-alive
- 自动重试和指数退避
- 上下文取消支持

## 许可证

Apache 2.0

## 技术支持

- 官网: https://xiangxinai.cn
- 文档: https://docs.xiangxinai.cn
- 问题反馈: https://github.com/xiangxinai/xiangxin-guardrails/issues
- 邮箱: wanglei@xiangxinai.cn

## 贡献指南

欢迎提交 Issue 和 Pull Request！

## 更新日志
### v2.0.0
- 新增 check_response_ctx(prompt, resposne)接口，与check_prompt(prmopt)配合使用，方便使用。

### v1.1.1
- 将最大检测内容长度从10000调整到1M

### v1.1.0
- 初始版本发布
- 支持提示词检测和对话上下文检测
- 完整的错误处理和重试机制
- 并发安全的客户端实现