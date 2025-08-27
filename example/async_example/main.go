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
	defer asyncClient.Close()
	
	ctx := context.Background()
	
	// 示例1: 单个异步检测
	fmt.Println("=== 示例1: 异步检测提示词 ===")
	asyncPromptExample(ctx, asyncClient)
	
	fmt.Println()
	
	// 示例2: 异步对话检测
	fmt.Println("=== 示例2: 异步检测对话上下文 ===")
	asyncConversationExample(ctx, asyncClient)
	
	fmt.Println()
	
	// 示例3: 批量异步检测
	fmt.Println("=== 示例3: 批量异步检测 ===")
	batchAsyncExample(ctx, asyncClient)
	
	fmt.Println()
	
	// 示例4: 带超时的异步检测
	fmt.Println("=== 示例4: 带超时的异步检测 ===")
	timeoutExample(asyncClient)
	
	fmt.Println()
	
	// 示例5: 并发控制示例
	fmt.Println("=== 示例5: 并发控制 ===")
	concurrencyExample(asyncClient)
}

// 异步检测提示词示例
func asyncPromptExample(ctx context.Context, client *xiangxinai.AsyncClient) {
	resultChan := client.CheckPromptAsync(ctx, "我想学习人工智能")
	
	select {
	case result := <-resultChan:
		if result.Error != nil {
			log.Printf("❌ 检测失败: %v", result.Error)
		} else {
			fmt.Printf("✅ 检测完成\n")
			fmt.Printf("   风险等级: %s\n", result.Result.OverallRiskLevel)
			fmt.Printf("   建议动作: %s\n", result.Result.SuggestAction)
			fmt.Printf("   是否安全: %t\n", result.Result.IsSafe())
		}
	case <-time.After(5 * time.Second):
		fmt.Println("⏰ 检测超时")
	}
}

// 异步对话检测示例
func asyncConversationExample(ctx context.Context, client *xiangxinai.AsyncClient) {
	messages := []*xiangxinai.Message{
		xiangxinai.NewMessage("user", "你好，我想了解人工智能"),
		xiangxinai.NewMessage("assistant", "您好！我很乐意为您介绍人工智能的相关知识"),
		xiangxinai.NewMessage("user", "请详细说明一下机器学习的基本概念"),
	}
	
	resultChan := client.CheckConversationAsync(ctx, messages)
	
	select {
	case result := <-resultChan:
		if result.Error != nil {
			log.Printf("❌ 对话检测失败: %v", result.Error)
		} else {
			fmt.Printf("✅ 对话检测完成\n")
			fmt.Printf("   对话风险等级: %s\n", result.Result.OverallRiskLevel)
			fmt.Printf("   建议动作: %s\n", result.Result.SuggestAction)
			
			if result.Result.IsSafe() {
				fmt.Println("   ✅ 对话安全，可以继续")
			} else if result.Result.IsBlocked() {
				fmt.Println("   ❌ 对话存在风险，建议阻断")
				fmt.Printf("   风险类别: %v\n", result.Result.GetAllCategories())
			} else if result.Result.HasSubstitute() && result.Result.SuggestAnswer != nil {
				fmt.Printf("   💡 建议使用安全回答: %s\n", *result.Result.SuggestAnswer)
			}
		}
	case <-time.After(5 * time.Second):
		fmt.Println("⏰ 对话检测超时")
	}
}

// 批量异步检测示例
func batchAsyncExample(ctx context.Context, client *xiangxinai.AsyncClient) {
	contents := []string{
		"我想学习编程",
		"请介绍一下Python语言",
		"如何开始学习机器学习",
		"人工智能的发展历史",
		"深度学习的基本原理",
	}
	
	fmt.Printf("开始批量检测 %d 个内容...\n", len(contents))
	startTime := time.Now()
	
	resultChan := client.BatchCheckPrompts(ctx, contents)
	
	successCount := 0
	failCount := 0
	
	for result := range resultChan {
		if result.Error != nil {
			failCount++
			log.Printf("❌ 批量检测失败: %v", result.Error)
		} else {
			successCount++
			fmt.Printf("✅ 检测结果: %s (动作: %s)\n", 
				result.Result.OverallRiskLevel, result.Result.SuggestAction)
		}
	}
	
	duration := time.Since(startTime)
	fmt.Printf("批量检测完成: 成功 %d 个, 失败 %d 个, 耗时: %v\n", 
		successCount, failCount, duration)
}

// 带超时的异步检测示例
func timeoutExample(client *xiangxinai.AsyncClient) {
	// 创建5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	resultChan := client.CheckPromptAsync(ctx, "测试超时检测")
	
	select {
	case result := <-resultChan:
		if result.Error != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Println("⏰ 检测超时")
			} else {
				log.Printf("❌ 检测失败: %v", result.Error)
			}
		} else {
			fmt.Printf("✅ 检测完成: %s\n", result.Result.OverallRiskLevel)
		}
	case <-ctx.Done():
		fmt.Println("⏰ 上下文超时或取消")
	}
}

// 并发控制示例
func concurrencyExample(client *xiangxinai.AsyncClient) {
	fmt.Printf("当前并发限制: %d\n", client.GetConcurrency())
	fmt.Printf("当前活跃工作线程: %d\n", client.GetActiveWorkers())
	
	// 启动多个并发检测
	ctx := context.Background()
	var results []<-chan xiangxinai.AsyncResult[*xiangxinai.GuardrailResponse]
	
	contents := []string{
		"并发测试1", "并发测试2", "并发测试3", 
		"并发测试4", "并发测试5", "并发测试6",
		"并发测试7", "并发测试8", "并发测试9", "并发测试10",
	}
	
	// 启动所有异步检测
	for _, content := range contents {
		resultChan := client.CheckPromptAsync(ctx, content)
		results = append(results, resultChan)
	}
	
	fmt.Printf("已启动 %d 个并发检测任务\n", len(results))
	fmt.Printf("当前活跃工作线程: %d\n", client.GetActiveWorkers())
	
	// 收集所有结果
	for i, resultChan := range results {
		result := <-resultChan
		if result.Error != nil {
			log.Printf("❌ 并发检测 %d 失败: %v", i+1, result.Error)
		} else {
			fmt.Printf("✅ 并发检测 %d 完成: %s\n", i+1, result.Result.SuggestAction)
		}
	}
	
	fmt.Printf("所有并发检测完成，当前活跃工作线程: %d\n", client.GetActiveWorkers())
}