package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "acap",
	Short: "AI 기반 영상 편집 CLI 툴",
	Long: `acap - AI 기반 영상 편집 CLI 툴

로컬 Whisper로 음성을 인식하고 DeepSeek AI가 편집을 제안합니다.

사전 준비:
  1. ffmpeg 설치:       brew install ffmpeg
  2. Whisper 설치:      pip install openai-whisper
  3. DeepSeek API 키:   export DEEPSEEK_API_KEY="sk-..."

명령어:
  acap edit <영상>        AI가 불필요한 구간을 찾아 자동으로 잘라냅니다
  acap subtitle <영상>    AI가 자막 내용·위치·스타일을 제안하고 영상에 추가합니다
  acap transcribe <영상>  음성을 텍스트로 변환해서 타임스탬프와 함께 출력합니다`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(transcribeCmd)
	rootCmd.AddCommand(subtitleCmd)

	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		fmt.Fprintln(os.Stderr, "경고: DEEPSEEK_API_KEY 환경변수가 설정되지 않았습니다")
	}
}
