package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TranscriptSegment struct {
	Start float64
	End   float64
	Text  string
}

type Transcript struct {
	Segments []TranscriptSegment
	FullText string
}

type whisperSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type whisperOutput struct {
	Text     string           `json:"text"`
	Segments []whisperSegment `json:"segments"`
}

func Transcribe(audioPath string) (*Transcript, error) {
	tmpDir, err := os.MkdirTemp("", "vidai-whisper-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	cmd := exec.Command("whisper", audioPath,
		"--model", "base",
		"--language", "ko",
		"--output_format", "json",
		"--output_dir", tmpDir,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("Whisper 실패: %s", string(out))
	}

	base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	jsonPath := filepath.Join(tmpDir, base+".json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("Whisper 결과 파일을 찾을 수 없습니다: %v", err)
	}

	var wo whisperOutput
	if err := json.Unmarshal(data, &wo); err != nil {
		return nil, fmt.Errorf("Whisper 결과 파싱 실패: %v", err)
	}

	t := &Transcript{FullText: wo.Text}
	for _, seg := range wo.Segments {
		t.Segments = append(t.Segments, TranscriptSegment{
			Start: seg.Start,
			End:   seg.End,
			Text:  strings.TrimSpace(seg.Text),
		})
	}
	return t, nil
}

func (t *Transcript) FormatForAI() string {
	var sb string
	for _, seg := range t.Segments {
		sb += fmt.Sprintf("[%.1f초 ~ %.1f초] %s\n", seg.Start, seg.End, seg.Text)
	}
	return sb
}
