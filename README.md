# acap

로컬 Whisper로 음성을 인식하고 DeepSeek AI가 영상 편집을 도와주는 CLI 툴입니다.

## 기능

| 명령어 | 설명 |
|--------|------|
| `acap edit` | AI가 불필요한 구간(침묵, 필러, NG)을 찾아 자동으로 잘라냄 |
| `acap subtitle` | AI가 자막 내용·위치·스타일을 제안하고 SRT 생성 또는 영상에 직접 구워줌 |
| `acap transcribe` | 음성을 텍스트로 변환해서 타임스탬프와 함께 출력 |

## 설치

### 사전 준비

```bash
# ffmpeg
brew install ffmpeg

# Whisper (로컬 STT, 무료)
pip install openai-whisper

# DeepSeek API 키 발급 → https://platform.deepseek.com
export DEEPSEEK_API_KEY="sk-..."
```

### 빌드

```bash
git clone https://github.com/wqrvQ2WR/acap
cd acap
go build -o acap .
cp acap /usr/local/bin/acap
```

## 사용법

### edit — AI 자동 편집

영상의 불필요한 구간을 잘라냅니다.

```bash
acap edit 영상.mp4
acap edit 영상.mp4 -o 결과.mp4   # 출력 경로 지정
acap edit 영상.mp4 --auto         # 확인 없이 자동 적용
```

**제거 대상:**
- 긴 침묵 / 어색한 정지
- 반복되는 내용
- 실수 · NG 구간
- "어...", "음..." 같은 필러가 많은 구간
- 주제와 관련 없는 잡담

**실행 예시:**
```
🎬 acap - AI 영상 편집 툴
────────────────────────────────────────
[1/3] 오디오 추출 중... 완료
[2/3] 음성 인식 중 (Whisper)... 완료 (24개 구간)
[3/3] AI 편집 분석 중 (DeepSeek)... 완료

📌 요약: 코딩 튜토리얼 영상으로 React 기초를 설명합니다

✂️  제거 제안 구간 (3개):
  1. 4.0초 ~ 7.2초 (3.2초)
  2. 43.5초 ~ 48.0초 (4.5초)
  3. 112.0초 ~ 115.3초 (3.3초)

💡 이유: 시작 부분 침묵, 중간 NG 구간, 마무리 잡담

편집을 적용할까요? [y/N]
```

---

### subtitle — AI 자막 생성

AI가 자막 내용, 위치, 스타일을 제안합니다.

```bash
acap subtitle 영상.mp4                    # 대화형으로 진행
acap subtitle 영상.mp4 --srt-only         # SRT 파일만 생성
acap subtitle 영상.mp4 --burn             # 자막을 영상에 직접 굽기
acap subtitle 영상.mp4 --burn -o 결과.mp4
```

**AI가 제안하는 것:**
- 몇 초 ~ 몇 초에 어떤 자막을 달지
- 위치: 상단 / 하단
- 스타일: 일반 / ★ 강조 (핵심 내용) / 💬 보충 설명
- 필러("어", "음") 자동 제거
- 긴 문장 자동 줄 나누기

---

### transcribe — 음성 → 텍스트

AI 분석 없이 STT 결과만 확인할 때 사용합니다.

```bash
acap transcribe 영상.mp4
```

```
[0.0초 ~ 2.4초] 안녕하세요, 오늘은 React를 알아보겠습니다.
[2.4초 ~ 5.1초] 먼저 컴포넌트가 뭔지부터 설명드릴게요.
...
```

## 구조

```
.
├── main.go
├── cmd/
│   ├── root.go        # CLI 진입점
│   ├── edit.go        # edit 명령어
│   ├── subtitle.go    # subtitle 명령어
│   └── transcribe.go  # transcribe 명령어
└── internal/
    ├── ffmpeg.go      # 오디오 추출 / 영상 편집
    ├── stt.go         # Whisper STT
    ├── ai.go          # DeepSeek 편집 분석
    └── subtitle.go    # DeepSeek 자막 생성 / SRT / 굽기
```

## 기술 스택

- **Go** — CLI 빌드
- **Cobra** — CLI 프레임워크
- **OpenAI Whisper** — 로컬 STT (무료)
- **DeepSeek API** — AI 편집 · 자막 분석
- **ffmpeg** — 오디오 추출 / 영상 편집 / 자막 굽기
