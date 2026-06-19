package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"vidai/internal"

	"github.com/spf13/cobra"
)

var (
	outputPath string
	autoApply  bool
)

var editCmd = &cobra.Command{
	Use:   "edit <영상파일>",
	Short: "AI가 불필요한 구간을 찾아 자동으로 잘라냅니다",
	Long: `영상의 음성을 인식한 뒤 DeepSeek AI가 편집 구간을 제안합니다.

제거 대상:
  - 긴 침묵 / 어색한 정지
  - 반복되는 내용
  - 실수·NG 구간
  - "어...", "음..." 같은 필러가 많은 구간
  - 주제와 관련 없는 잡담

사용 예시:
  acap edit 영상.mp4
  acap edit 영상.mp4 -o 결과.mp4
  acap edit 영상.mp4 --auto`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	editCmd.Flags().StringVarP(&outputPath, "output", "o", "", "출력 파일 경로 (기본값: 원본_edited.mp4)")
	editCmd.Flags().BoolVarP(&autoApply, "auto", "a", false, "확인 없이 자동으로 편집 적용")
}

func runEdit(cmd *cobra.Command, args []string) error {
	videoPath := args[0]

	if _, err := os.Stat(videoPath); err != nil {
		return fmt.Errorf("파일을 찾을 수 없습니다: %s", videoPath)
	}

	if err := internal.CheckFFmpeg(); err != nil {
		return err
	}

	if outputPath == "" {
		ext := filepath.Ext(videoPath)
		base := strings.TrimSuffix(videoPath, ext)
		outputPath = base + "_edited" + ext
	}

	fmt.Println("🎬 vidai - AI 영상 편집 툴")
	fmt.Println(strings.Repeat("─", 40))

	// Step 1: Extract audio
	fmt.Printf("\n[1/3] 오디오 추출 중... ")
	audioPath, err := internal.ExtractAudio(videoPath)
	if err != nil {
		return err
	}
	defer os.Remove(audioPath)
	fmt.Println("완료")

	// Step 2: STT
	fmt.Printf("[2/3] 음성 인식 중 (Whisper)... ")
	transcript, err := internal.Transcribe(audioPath)
	if err != nil {
		return err
	}
	fmt.Printf("완료 (%d개 구간)\n", len(transcript.Segments))

	fmt.Println("\n📝 인식된 자막:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println(transcript.FormatForAI())

	// Step 3: AI analysis
	fmt.Printf("[3/3] AI 편집 분석 중 (DeepSeek)... ")
	suggestion, err := internal.GetEditSuggestions(transcript)
	if err != nil {
		return err
	}
	fmt.Println("완료")

	// Show results
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("🤖 AI 편집 제안")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("📌 요약: %s\n\n", suggestion.Summary)

	if len(suggestion.Highlights) > 0 {
		fmt.Println("✨ 핵심 포인트:")
		for _, h := range suggestion.Highlights {
			fmt.Printf("  • %s\n", h)
		}
		fmt.Println()
	}

	if len(suggestion.CutSegments) > 0 {
		fmt.Printf("✂️  제거 제안 구간 (%d개):\n", len(suggestion.CutSegments))
		for i, seg := range suggestion.CutSegments {
			fmt.Printf("  %d. %.1f초 ~ %.1f초 (%.1f초)\n", i+1, seg.Start, seg.End, seg.End-seg.Start)
		}
		fmt.Printf("\n💡 이유: %s\n", suggestion.Reasoning)
	} else {
		fmt.Println("✅ 제거할 구간 없음 - 영상이 이미 잘 편집되어 있습니다!")
		return nil
	}

	// Confirm
	if !autoApply {
		fmt.Printf("\n편집을 적용할까요? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("취소됨")
			return nil
		}
	}

	// Apply edits
	fmt.Printf("\n영상 편집 적용 중... ")
	if err := internal.CutSegments(videoPath, suggestion.CutSegments, outputPath); err != nil {
		return err
	}
	fmt.Printf("완료!\n\n")
	fmt.Printf("💾 저장됨: %s\n", outputPath)

	return nil
}
