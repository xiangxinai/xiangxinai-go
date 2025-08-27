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

	// 示例1: 检测单个提示词
	fmt.Println("=== 示例1: 检测单个提示词 ===")
	result1, err := client.CheckPrompt(ctx, "我想学习编程")
	if err != nil {
		log.Printf("检测失败: %v", err)
	} else {
		fmt.Printf("风险等级: %s\n", result1.OverallRiskLevel)
		fmt.Printf("建议动作: %s\n", result1.SuggestAction)
		fmt.Printf("是否安全: %t\n", result1.IsSafe())
	}

	fmt.Println()

	// 示例2: 检测对话上下文（推荐）
	fmt.Println("=== 示例2: 检测对话上下文 ===")
	messages := []*xiangxinai.Message{
		xiangxinai.NewMessage("user", "你好，我有个问题"),
		xiangxinai.NewMessage("assistant", "您好！我很乐意帮助您解答问题。请问您想了解什么？"),
		xiangxinai.NewMessage("user", "我想了解人工智能的发展历史"),
	}

	result2, err := client.CheckConversation(ctx, messages)
	if err != nil {
		log.Printf("检测失败: %v", err)
	} else {
		fmt.Printf("对话风险等级: %s\n", result2.OverallRiskLevel)
		fmt.Printf("建议动作: %s\n", result2.SuggestAction)
		
		if result2.IsSafe() {
			fmt.Println("✅ 对话安全，可以继续")
		} else if result2.IsBlocked() {
			fmt.Println("❌ 对话存在风险，建议阻断")
			fmt.Printf("风险类别: %v\n", result2.GetAllCategories())
		} else if result2.HasSubstitute() && result2.SuggestAnswer != nil {
			fmt.Printf("💡 建议使用安全回答: %s\n", *result2.SuggestAnswer)
		}
	}

	fmt.Println()

	// 示例3: 健康检查
	fmt.Println("=== 示例3: API健康检查 ===")
	health, err := client.HealthCheck(ctx)
	if err != nil {
		log.Printf("健康检查失败: %v", err)
	} else {
		fmt.Printf("健康状态: %v\n", health)
	}

	fmt.Println()

	// 示例4: 获取可用模型
	fmt.Println("=== 示例4: 获取可用模型 ===")
	models, err := client.GetModels(ctx)
	if err != nil {
		log.Printf("获取模型列表失败: %v", err)
	} else {
		fmt.Printf("可用模型: %v\n", models)
	}

	fmt.Println()

	// 示例5: 错误处理
	fmt.Println("=== 示例5: 错误处理示例 ===")
	_, err = client.CheckPrompt(ctx, "")
	if err != nil {
		handleError(err)
	} else {
		fmt.Println("空内容检测成功")
	}
}

func handleError(err error) {
	switch e := err.(type) {
	case *xiangxinai.AuthenticationError:
		fmt.Printf("❌ 认证失败: %v\n", e)
	case *xiangxinai.RateLimitError:
		fmt.Printf("⏰ 请求频率过高: %v\n", e)
	case *xiangxinai.ValidationError:
		fmt.Printf("📝 输入验证错误: %v\n", e)
	case *xiangxinai.NetworkError:
		fmt.Printf("🌐 网络连接错误: %v\n", e)
	case *xiangxinai.XiangxinAIError:
		fmt.Printf("⚠️ API错误: %v\n", e)
	default:
		fmt.Printf("❓ 未知错误: %v\n", e)
	}
}