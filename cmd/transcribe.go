package cmd

import (
	"fmt"
	"os"

	"vidai/internal"

	"github.com/spf13/cobra"
)

var transcribeCmd = &cobra.Command{
	Use:   "transcribe <영상파일>",
	Short: "음성을 텍스트로 변환해서 타임스탬프와 함께 출력합니다",
	Long: `로컬 Whisper 모델로 음성을 인식해서 타임스탬프와 함께 출력합니다.
AI 분석 없이 STT 결과만 확인할 때 사용합니다.

사용 예시:
  acap transcribe 영상.mp4`,
	Args: cobra.ExactArgs(1),
	RunE: runTranscribe,
}

func runTranscribe(cmd *cobra.Command, args []string) error {
	videoPath := args[0]

	if _, err := os.Stat(videoPath); err != nil {
		return fmt.Errorf("파일을 찾을 수 없습니다: %s", videoPath)
	}

	if err := internal.CheckFFmpeg(); err != nil {
		return err
	}

	fmt.Printf("오디오 추출 중... ")
	audioPath, err := internal.ExtractAudio(videoPath)
	if err != nil {
		return err
	}
	defer os.Remove(audioPath)
	fmt.Println("완료")

	fmt.Printf("음성 인식 중... ")
	transcript, err := internal.Transcribe(audioPath)
	if err != nil {
		return err
	}
	fmt.Println("완료\n")

	fmt.Println(transcript.FormatForAI())
	return nil
}
