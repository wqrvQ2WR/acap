package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// callAI sends a system+user prompt to the configured OpenAI-compatible model,
// strips any markdown code fence from the reply, and unmarshals the JSON into out.
// All edit/subtitle analysis goes through here so provider settings stay in one place.
func callAI(systemPrompt, userMsg string, out interface{}) error {
	cfg := LoadConfig()

	apiKey := os.Getenv(cfg.APIKeyEnv)
	if apiKey == "" {
		return fmt.Errorf("%s 환경변수가 필요합니다 ('acap api' 로 현재 설정을 확인하세요)", cfg.APIKeyEnv)
	}

	oc := openai.DefaultConfig(apiKey)
	if cfg.BaseURL != "" {
		oc.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(oc)

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: cfg.Model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return fmt.Errorf("AI 호출 실패 (%s/%s): %v", cfg.Provider, cfg.Model, err)
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("AI 응답이 비어 있습니다")
	}

	content := stripCodeFence(resp.Choices[0].Message.Content)
	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("AI 응답 파싱 실패: %v\n응답: %s", err, content)
	}
	return nil
}

// stripCodeFence removes a surrounding ```json ... ``` markdown fence if present.
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
