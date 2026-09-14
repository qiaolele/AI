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

// 默认配置（如果环境变量未设置，则使用此默认值）
const (
	defaultApiUrl = "https://api.deepseek.com/v1/chat/completions"
	defaultModel  = "deepseek-chat"
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
	
	// 如果没有配置密钥，或者配置为 mock，则直接返回一段模拟的优化结果，方便测试流程
	if apiKey == "" || apiKey == "mock" {
		return "【系统提示：当前使用的是模拟数据，因为您没有配置真实的 AI 密钥】\n\n1. 熟练掌握相关专业技能，具备从0到1搭建复杂系统的经验。\n2. 深入理解工程化架构，曾主导团队核心业务系统重构，提升运行效率达40%。\n3. 具有良好的业务抽象设计思维，能有效提升跨团队协同开发效率。\n\n*(如需真实 AI 生成，请前往云托管环境变量配置真实的 AI_API_KEY)*", nil
	}

	// 构建系统提示词
	systemPrompt := `你是一名资深的 HR 和职业规划专家。用户会提供他们的目标岗位和原始工作经历。
请你帮他们扩写和润色工作经历。要求：
1. 提取核心亮点，使用专业术语。
2. 采用 STAR 法则（情境、任务、行动、结果）进行改写。
3. 重点突出对业务的价值和可量化的数据。
4. 返回的内容直接是润色后的经历，不需要寒暄，分点列出即可。`

	// 获取用户自定义的 API 地址和模型名称（支持各种兼容 OpenAI 的 API）
	apiUrl := os.Getenv("AI_API_URL")
	if apiUrl == "" {
		apiUrl = defaultApiUrl
	}
	modelName := os.Getenv("AI_MODEL")
	if modelName == "" {
		modelName = defaultModel
	}

	userPrompt := fmt.Sprintf("我的目标岗位是：%s\n我的原始工作经历是：%s", targetJob, experience)

	reqBody := AIRequest{
		Model: modelName,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
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
