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
	musicOutput  string
	musicVolume  float64
	musicReplace bool
	musicNoLoop  bool
)

var musicCmd = &cobra.Command{
	Use:   "music <영상파일> <음악파일>",
	Short: "영상에 배경음악을 넣습니다",
	Long: `영상에 배경음악(mp3, m4a, wav 등)을 추가합니다.
기본적으로 원본 소리 위에 음악을 깔며, 음악은 영상 길이에 맞춰 자동 반복됩니다.

사용 예시:
  acap music 영상.mp4 bgm.mp3                  # 원본 소리 위에 음악 믹스
  acap music 영상.mp4 bgm.mp3 --volume 0.15    # 음악을 더 작게
  acap music 영상.mp4 bgm.mp3 --replace        # 원본 소리 제거하고 음악만
  acap music 영상.mp4 bgm.mp3 --no-loop        # 음악을 반복하지 않음 (끝나면 무음)
  acap music 영상.mp4 bgm.mp3 -o 결과.mp4`,
	Args: cobra.ExactArgs(2),
	RunE: runMusic,
}

func init() {
	musicCmd.Flags().StringVarP(&musicOutput, "output", "o", "", "출력 파일 경로")
	musicCmd.Flags().Float64VarP(&musicVolume, "volume", "v", 0.3, "음악 음량 (0.0 ~ 1.0+)")
	musicCmd.Flags().BoolVar(&musicReplace, "replace", false, "원본 소리를 제거하고 음악으로 교체")
	musicCmd.Flags().BoolVar(&musicNoLoop, "no-loop", false, "음악을 영상 길이에 맞춰 반복하지 않음")
}

func runMusic(cmd *cobra.Command, args []string) error {
	videoPath := args[0]
	musicPath := args[1]

	if _, err := os.Stat(videoPath); err != nil {
		return fmt.Errorf("영상 파일을 찾을 수 없습니다: %s", videoPath)
	}
	if _, err := os.Stat(musicPath); err != nil {
		return fmt.Errorf("음악 파일을 찾을 수 없습니다: %s", musicPath)
	}
	if musicVolume < 0 {
		return fmt.Errorf("음량은 0 이상이어야 합니다: %.3f", musicVolume)
	}
	if err := internal.CheckFFmpeg(); err != nil {
		return err
	}

	if musicOutput == "" {
		ext := filepath.Ext(videoPath)
		base := strings.TrimSuffix(videoPath, ext)
		musicOutput = base + "_bgm" + ext
	}

	mode := "원본 소리 위에 믹스"
	if musicReplace {
		mode = "원본 소리 교체"
	}
	fmt.Printf("배경음악 추가 중 (%s, 음량 %.2f)... ", mode, musicVolume)

	if err := internal.AddBackgroundMusic(videoPath, musicPath, musicOutput, musicVolume, musicReplace, !musicNoLoop); err != nil {
		return err
	}

	fmt.Printf("완료!\n💾 저장됨: %s\n", musicOutput)
	return nil
}
