package internal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	userMsg := fmt.Sprintf("다음 자막 원본을 분석해서 최적화된 자막을 만들어주세요:\n\n%s", transcript.FormatForAI())

	var suggestion SubtitleSuggestion
	if err := callAI(subtitlePrompt, userMsg, &suggestion); err != nil {
		return nil, err
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
	// so we burn with a clean style.
	// 필터그래프 파서가 콤마(,)를 필터 구분자로 보기 때문에 force_style 안의 콤마는
	// 반드시 \, 로 이스케이프해야 한다. 경로의 특수문자도 따옴표 대신 이스케이프한다.
	if err := checkSubtitlesFilter(); err != nil {
		return err
	}

	style := "FontSize=22,FontName=Arial,PrimaryColour=&H00FFFFFF,OutlineColour=&H00000000,Outline=2,Alignment=2"
	filter := fmt.Sprintf("subtitles=%s:force_style=%s",
		escapeFilterArg(srtPath),
		escapeFilterArg(style),
	)

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

// checkSubtitlesFilter verifies that the ffmpeg build includes the subtitles
// filter (libass). Many homebrew builds ship without --enable-libass, in which
// case burning subtitles is impossible and ffmpeg fails with a cryptic
// "No such filter" / "No option name" error.
func checkSubtitlesFilter() error {
	out, err := exec.Command("ffmpeg", "-hide_banner", "-filters").Output()
	if err != nil {
		return fmt.Errorf("ffmpeg 필터 목록 확인 실패: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, " subtitles ") {
			return nil
		}
	}
	return fmt.Errorf("설치된 ffmpeg에 자막(subtitles) 필터가 없습니다 (libass 미포함).\n" +
		"  → brew reinstall ffmpeg  로 재설치하거나\n" +
		"  → --srt-only 옵션으로 SRT 파일만 생성하세요")
}

// escapeFilterArg escapes characters that have special meaning inside an
// ffmpeg filtergraph argument (path or option value): backslash, colon,
// single quote, comma, and the filtergraph separators.
func escapeFilterArg(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`:`, `\:`,
		`'`, `\'`,
		`,`, `\,`,
		`[`, `\[`,
		`]`, `\]`,
		`;`, `\;`,
	)
	return r.Replace(s)
}

func toSRTTime(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}
