package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type SubtitleEntry struct {
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Text     string  `json:"text"`
	Position string  `json:"position"` // "top" | "bottom"
	Style    string  `json:"style"`    // "normal" | "emphasis" | "caption"
}

type SubtitleSuggestion struct {
	Entries []SubtitleEntry `json:"entries"`
}

const subtitlePrompt = `당신은 유튜브 영상 자막 전문가입니다.
타임스탬프가 있는 자막 원본을 분석해서 최적화된 자막 배치를 JSON으로 반환합니다.

반환 형식:
{
  "entries": [
    {
      "start": 시작초,
      "end": 종료초,
      "text": "자막 내용",
      "position": "bottom",
      "style": "normal"
    }
  ]
}

규칙:
- position: "bottom" (기본), "top" (화면 아래 가릴 때)
- style: "normal" (일반), "emphasis" (강조 - 핵심 내용), "caption" (보충 설명)
- 너무 긴 자막은 2줄로 나눠서 \n 사용
- 한 자막은 최대 20자 이내로 간결하게
- 불필요한 필러("어", "음", "그")는 제거
- 핵심 키워드나 중요 포인트는 emphasis로

응답은 반드시 JSON만 반환하세요.`

func GetSubtitleSuggestions(transcript *Transcript) (*SubtitleSuggestion, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY 환경변수가 필요합니다")
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(cfg)

	userMsg := fmt.Sprintf("다음 자막 원본을 분석해서 최적화된 자막을 만들어주세요:\n\n%s", transcript.FormatForAI())

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "deepseek-chat",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: subtitlePrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMsg},
		},
		Temperature: 0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("자막 분석 실패: %v", err)
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var suggestion SubtitleSuggestion
	if err := json.Unmarshal([]byte(content), &suggestion); err != nil {
		return nil, fmt.Errorf("자막 응답 파싱 실패: %v\n응답: %s", err, content)
	}

	return &suggestion, nil
}

// GenerateSRT writes an SRT file from subtitle entries
func GenerateSRT(entries []SubtitleEntry, outputPath string) error {
	var sb strings.Builder
	for i, e := range entries {
		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		sb.WriteString(fmt.Sprintf("%s --> %s\n", toSRTTime(e.Start), toSRTTime(e.End)))
		sb.WriteString(e.Text + "\n\n")
	}
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// BurnSubtitles burns SRT subtitles into the video using ffmpeg
func BurnSubtitles(videoPath, srtPath, outputPath string) error {
	// Use subtitles filter; emphasis style can't easily be per-entry with basic SRT,
	// so we burn with a clean style
	filter := fmt.Sprintf("subtitles='%s':force_style='FontSize=22,FontName=Arial,PrimaryColour=&H00FFFFFF,OutlineColour=&H00000000,Outline=2,Alignment=2'", srtPath)

	cmd := newFFmpegCmd("-y",
		"-i", videoPath,
		"-vf", filter,
		"-c:a", "copy",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("자막 굽기 실패: %s", string(out))
	}
	return nil
}

func toSRTTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
