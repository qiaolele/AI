package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// 以 DeepSeek 为例 (兼容 OpenAI 格式)
const (
	// 稍后您需要把这里换成您选用的模型 API 地址
	aiApiUrl = "https://api.deepseek.com/v1/chat/completions"
)

type AIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func OptimizeResume(targetJob string, experience string) (string, error) {
	// 推荐将 API_KEY 配置在云托管的环境变量中，而非硬编码
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("AI_API_KEY environment variable is not set")
	}

	// 构建系统提示词
	systemPrompt := `你是一名资深的 HR 和职业规划专家。用户会提供他们的目标岗位和原始工作经历。
请你帮他们扩写和润色工作经历。要求：
1. 提取核心亮点，使用专业术语。
2. 采用 STAR 法则（情境、任务、行动、结果）进行改写。
3. 重点突出对业务的价值和可量化的数据。
4. 返回的内容直接是润色后的经历，不需要寒暄，分点列出即可。`

	userPrompt := fmt.Sprintf("我的目标岗位是：%s\n我的原始工作经历是：%s", targetJob, experience)

	reqBody := AIRequest{
		Model: "deepseek-chat", // 替换为您使用的具体模型名
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", aiApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("AI API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var aiResp AIResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		return "", err
	}

	if len(aiResp.Choices) > 0 {
		return aiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from AI")
}
