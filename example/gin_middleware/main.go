package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xiangxinai/xiangxin-guardrails/client/xiangxinai-go"
)

// 护栏中间件
func GuardrailMiddleware(client *xiangxinai.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Content string `json:"content" binding:"required"`
		}
		
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
			c.Abort()
			return
		}
		
		// 进行安全检测
		result, err := client.CheckPrompt(context.Background(), req.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Guardrail check failed",
				"detail": err.Error(),
			})
			c.Abort()
			return
		}
		
		// 检查是否被阻断
		if result.IsBlocked() {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Content blocked by guardrail",
				"risk_level": result.OverallRiskLevel,
				"categories": result.GetAllCategories(),
				"suggest_action": result.SuggestAction,
			})
			c.Abort()
			return
		}
		
		// 如果有代答建议，添加到响应头
		if result.HasSubstitute() && result.SuggestAnswer != nil {
			c.Header("X-Suggested-Answer", *result.SuggestAnswer)
		}
		
		// 将检测结果添加到上下文
		c.Set("guardrail_result", result)
		c.Next()
	}
}

func main() {
	// 从环境变量获取API密钥
	apiKey := os.Getenv("XIANGXINAI_API_KEY")
	if apiKey == "" {
		panic("XIANGXINAI_API_KEY environment variable is required")
	}
	
	// 初始化护栏客户端
	client := xiangxinai.NewClient(apiKey)
	
	// 创建Gin路由器
	r := gin.Default()
	
	// 应用护栏中间件
	r.Use(GuardrailMiddleware(client))
	
	// 定义API端点
	r.POST("/chat", func(c *gin.Context) {
		// 获取检测结果
		result, exists := c.Get("guardrail_result")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Guardrail result not found"})
			return
		}
		
		guardrailResult := result.(*xiangxinai.GuardrailResponse)
		
		// 模拟聊天响应
		response := gin.H{
			"message": "您的消息已通过安全检测",
			"risk_level": guardrailResult.OverallRiskLevel,
			"safe": guardrailResult.IsSafe(),
		}
		
		// 如果有建议回答，使用建议回答
		if guardrailResult.HasSubstitute() && guardrailResult.SuggestAnswer != nil {
			response["suggested_response"] = *guardrailResult.SuggestAnswer
		}
		
		c.JSON(http.StatusOK, response)
	})
	
	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		health, err := client.HealthCheck(context.Background())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error": err.Error(),
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"guardrail_service": health,
		})
	})
	
	// 启动服务器
	r.Run(":8080")
}