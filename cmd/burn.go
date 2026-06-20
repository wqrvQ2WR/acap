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
	burnOutput string
	burnSoft   bool
)

var burnCmd = &cobra.Command{
	Use:   "burn <영상파일> <자막파일.srt>",
	Short: "기존 SRT 자막 파일을 영상에 넣습니다",
	Long: `이미 가지고 있는 .srt 자막 파일을 영상에 합칩니다.
AI 분석 없이 자막 파일만 영상에 넣을 때 사용합니다.

두 가지 방식:
  기본       자막을 영상에 직접 구워넣음 (어디서나 보임, libass 필요)
  --soft     자막 트랙으로 삽입 (플레이어에서 켜고 끌 수 있음, 재인코딩 없이 즉시)

사용 예시:
  acap burn 영상.mp4 자막.srt                # 영상에 자막 굽기
  acap burn 영상.mp4 자막.srt -o 결과.mp4
  acap burn 영상.mp4 자막.srt --soft         # 끌 수 있는 자막 트랙으로 삽입`,
	Args: cobra.ExactArgs(2),
	RunE: runBurn,
}

func init() {
	burnCmd.Flags().StringVarP(&burnOutput, "output", "o", "", "출력 파일 경로")
	burnCmd.Flags().BoolVar(&burnSoft, "soft", false, "자막을 트랙으로 삽입 (켜고 끌 수 있음, 재인코딩 없음)")
}

func runBurn(cmd *cobra.Command, args []string) error {
	videoPath := args[0]
	srtPath := args[1]

	if _, err := os.Stat(videoPath); err != nil {
		return fmt.Errorf("영상 파일을 찾을 수 없습니다: %s", videoPath)
	}
	if _, err := os.Stat(srtPath); err != nil {
		return fmt.Errorf("자막 파일을 찾을 수 없습니다: %s", srtPath)
	}
	if !strings.EqualFold(filepath.Ext(srtPath), ".srt") {
		return fmt.Errorf("자막 파일은 .srt 형식이어야 합니다: %s", srtPath)
	}
	if err := internal.CheckFFmpeg(); err != nil {
		return err
	}

	if burnOutput == "" {
		ext := filepath.Ext(videoPath)
		base := strings.TrimSuffix(videoPath, ext)
		burnOutput = base + "_subtitled" + ext
	}

	if burnSoft {
		fmt.Printf("자막 트랙 삽입 중... ")
		if err := internal.EmbedSubtitlesSoft(videoPath, srtPath, burnOutput); err != nil {
			return err
		}
		fmt.Printf("완료!\n💾 저장됨: %s\n", burnOutput)
		fmt.Println("(플레이어에서 자막을 켜야 보입니다)")
		return nil
	}

	fmt.Printf("자막 굽는 중... ")
	if err := internal.BurnSubtitles(videoPath, srtPath, burnOutput); err != nil {
		return err
	}
	fmt.Printf("완료!\n💾 저장됨: %s\n", burnOutput)
	return nil
}
