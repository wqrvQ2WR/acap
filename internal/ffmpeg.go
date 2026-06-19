package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func newFFmpegCmd(args ...string) *exec.Cmd {
	return exec.Command("ffmpeg", args...)
}

func CheckFFmpeg() error {
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg가 설치되어 있지 않습니다. brew install ffmpeg 로 설치하세요")
	}
	return nil
}

// ExtractAudio extracts audio from video as mp3
func ExtractAudio(videoPath string) (string, error) {
	audioPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + "_audio.mp3"

	cmd := exec.Command("ffmpeg", "-y",
		"-i", videoPath,
		"-vn",
		"-ar", "16000",
		"-ac", "1",
		"-b:a", "64k",
		audioPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("오디오 추출 실패: %s", string(out))
	}
	return audioPath, nil
}

type Segment struct {
	Start float64
	End   float64
}

// CutSegments cuts specified segments from video and merges the rest
func CutSegments(videoPath string, cutSegments []Segment, outputPath string) error {
	duration, err := getVideoDuration(videoPath)
	if err != nil {
		return err
	}

	// Build keep segments (inverse of cut)
	keepSegments := invertSegments(cutSegments, duration)
	if len(keepSegments) == 0 {
		return fmt.Errorf("자를 구간이 너무 많아서 영상이 남지 않습니다")
	}

	tmpDir, err := os.MkdirTemp("", "vidai-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Extract each keep segment
	var segFiles []string
	for i, seg := range keepSegments {
		segFile := filepath.Join(tmpDir, fmt.Sprintf("seg%d.mp4", i))
		cmd := exec.Command("ffmpeg", "-y",
			"-ss", formatTime(seg.Start),
			"-to", formatTime(seg.End),
			"-i", videoPath,
			"-c", "copy",
			segFile,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("구간 추출 실패 (%d): %s", i, string(out))
		}
		segFiles = append(segFiles, segFile)
	}

	// Create concat list
	listFile := filepath.Join(tmpDir, "list.txt")
	var listContent strings.Builder
	for _, f := range segFiles {
		listContent.WriteString(fmt.Sprintf("file '%s'\n", f))
	}
	if err := os.WriteFile(listFile, []byte(listContent.String()), 0644); err != nil {
		return err
	}

	// Concat
	cmd := exec.Command("ffmpeg", "-y",
		"-f", "concat",
		"-safe", "0",
		"-i", listFile,
		"-c", "copy",
		outputPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("영상 합치기 실패: %s", string(out))
	}

	return nil
}

func getVideoDuration(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("영상 길이 확인 실패: %v", err)
	}
	dur, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("영상 길이 파싱 실패: %v", err)
	}
	return dur, nil
}

func invertSegments(cuts []Segment, duration float64) []Segment {
	var keeps []Segment
	cursor := 0.0
	for _, cut := range cuts {
		if cut.Start > cursor {
			keeps = append(keeps, Segment{Start: cursor, End: cut.Start})
		}
		cursor = cut.End
	}
	if cursor < duration {
		keeps = append(keeps, Segment{Start: cursor, End: duration})
	}
	return keeps
}

func formatTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := seconds - float64(h*3600) - float64(m*60)
	return fmt.Sprintf("%02d:%02d:%06.3f", h, m, s)
}
