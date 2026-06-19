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
	subtitleOutput string
	burnIn         bool
	srtOnly        bool
)

var subtitleCmd = &cobra.Command{
	Use:   "subtitle <영상파일>",
	Short: "AI가 자막 내용·위치·스타일을 제안하고 영상에 추가합니다",
	Long: `음성을 인식한 뒤 DeepSeek AI가 최적화된 자막을 제안합니다.

AI가 제안하는 것:
  - 몇 초 ~ 몇 초에 어떤 자막을 달지
  - 위치: 상단 / 하단
  - 스타일: 일반 / ★ 강조 (핵심 내용) / 💬 보충 설명
  - 필러("어", "음") 자동 제거
  - 긴 문장 자동 줄 나누기

사용 예시:
  acap subtitle 영상.mp4                   # SRT 생성 여부 확인 후 진행
  acap subtitle 영상.mp4 --srt-only        # SRT 파일만 생성
  acap subtitle 영상.mp4 --burn            # 자막을 영상에 직접 굽기
  acap subtitle 영상.mp4 --burn -o 결과.mp4`,
	Args: cobra.ExactArgs(1),
	RunE: runSubtitle,
}

func init() {
	subtitleCmd.Flags().StringVarP(&subtitleOutput, "output", "o", "", "출력 파일 경로")
	subtitleCmd.Flags().BoolVarP(&burnIn, "burn", "b", false, "자막을 영상에 직접 굽기 (기본: SRT 파일만 생성)")
	subtitleCmd.Flags().BoolVar(&srtOnly, "srt-only", false, "SRT 파일만 생성하고 영상 편집 안 함")
}

func runSubtitle(cmd *cobra.Command, args []string) error {
	videoPath := args[0]

	if _, err := os.Stat(videoPath); err != nil {
		return fmt.Errorf("파일을 찾을 수 없습니다: %s", videoPath)
	}
	if err := internal.CheckFFmpeg(); err != nil {
		return err
	}

	ext := filepath.Ext(videoPath)
	base := strings.TrimSuffix(videoPath, ext)
	srtPath := base + ".srt"

	if subtitleOutput == "" {
		subtitleOutput = base + "_subtitled" + ext
	}

	fmt.Println("🎬 acap subtitle - AI 자막 생성")
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

	// Step 3: AI subtitle suggestions
	fmt.Printf("[3/3] AI 자막 분석 중 (DeepSeek)... ")
	suggestion, err := internal.GetSubtitleSuggestions(transcript)
	if err != nil {
		return err
	}
	fmt.Printf("완료 (%d개 자막)\n", len(suggestion.Entries))

	// Show preview
	fmt.Println("\n" + strings.Repeat("─", 40))
	fmt.Println("📝 AI 자막 제안")
	fmt.Println(strings.Repeat("─", 40))
	for _, e := range suggestion.Entries {
		styleTag := ""
		if e.Style == "emphasis" {
			styleTag = " ★"
		} else if e.Style == "caption" {
			styleTag = " 💬"
		}
		pos := "하단"
		if e.Position == "top" {
			pos = "상단"
		}
		fmt.Printf("  [%.1f초~%.1f초] (%s)%s %s\n", e.Start, e.End, pos, styleTag, e.Text)
	}

	if srtOnly {
		if err := internal.GenerateSRT(suggestion.Entries, srtPath); err != nil {
			return err
		}
		fmt.Printf("\n💾 SRT 저장됨: %s\n", srtPath)
		return nil
	}

	fmt.Printf("\n영상에 자막을 추가할까요? (--burn 옵션으로 영상에 직접 굽기) [y/N] ")
	if !burnIn {
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) == "y" {
			burnIn = true
		}
	}

	// Generate SRT
	if err := internal.GenerateSRT(suggestion.Entries, srtPath); err != nil {
		return err
	}
	fmt.Printf("\n💾 SRT 저장됨: %s\n", srtPath)

	if burnIn {
		fmt.Printf("자막 굽는 중... ")
		if err := internal.BurnSubtitles(videoPath, srtPath, subtitleOutput); err != nil {
			return err
		}
		fmt.Printf("완료!\n💾 영상 저장됨: %s\n", subtitleOutput)
	} else {
		fmt.Println("SRT 파일만 생성됨. 영상에 굽으려면 --burn 옵션 사용.")
	}

	return nil
}
