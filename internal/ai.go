package internal

import (
	"fmt"
)

type EditSuggestion struct {
	CutSegments []Segment `json:"cut_segments"`
	Summary     string    `json:"summary"`
	Highlights  []string  `json:"highlights"`
	Reasoning   string    `json:"reasoning"`
}

const systemPrompt = `당신은 유튜브 영상 편집 전문가입니다.
타임스탬프가 있는 자막을 분석해서 편집 제안을 JSON으로 반환합니다.

반환 형식:
{
  "cut_segments": [
    {"start": 시작초, "end": 종료초}
  ],
  "summary": "이 영상은 ~에 관한 내용입니다",
  "highlights": ["핵심 포인트 1", "핵심 포인트 2"],
  "reasoning": "편집 이유 설명"
}

cut_segments에는 제거할 구간을 넣으세요:
- 긴 침묵 또는 어색한 정지 구간
- 반복되는 내용
- 실수나 NG 구간 (잘못 말한 부분)
- 내용과 관련 없는 잡담
- "어...", "음..." 같은 필러가 많은 구간

응답은 반드시 JSON만 반환하세요.`

func GetEditSuggestions(transcript *Transcript) (*EditSuggestion, error) {
	userMsg := fmt.Sprintf("다음 자막을 분석해서 편집 제안을 해주세요:\n\n%s", transcript.FormatForAI())

	var suggestion EditSuggestion
	if err := callAI(systemPrompt, userMsg, &suggestion); err != nil {
		return nil, err
	}
	return &suggestion, nil
}
