# acap

AI-powered video editing CLI tool. Transcribes speech with local Whisper and uses DeepSeek AI to suggest edits and subtitles.

로컬 Whisper로 음성을 인식하고 DeepSeek AI가 영상 편집을 도와주는 CLI 툴입니다.

---

## Features / 기능

| Command | Description |
|---------|-------------|
| `acap edit` | AI detects and removes unnecessary segments (silence, fillers, NG takes) |
| `acap subtitle` | AI suggests subtitle content, position, and style — generates SRT or burns into video |
| `acap transcribe` | Transcribes speech to text with timestamps |

| 명령어 | 설명 |
|--------|------|
| `acap edit` | AI가 불필요한 구간(침묵, 필러, NG)을 찾아 자동으로 잘라냄 |
| `acap subtitle` | AI가 자막 내용·위치·스타일을 제안하고 SRT 생성 또는 영상에 직접 구워줌 |
| `acap transcribe` | 음성을 텍스트로 변환해서 타임스탬프와 함께 출력 |

---

## Installation / 설치

### Prerequisites / 사전 준비

```bash
# ffmpeg
brew install ffmpeg

# Whisper (local STT, free / 로컬 STT, 무료)
pip install openai-whisper

# DeepSeek API key → https://platform.deepseek.com
export DEEPSEEK_API_KEY="sk-..."
```

### Build / 빌드

```bash
git clone https://github.com/wqrvQ2WR/acap
cd acap
go build -o acap .
cp acap /usr/local/bin/acap
```

---

## Usage / 사용법

### edit — AI Auto Edit / AI 자동 편집

Cuts unnecessary segments from your video.
영상의 불필요한 구간을 잘라냅니다.

```bash
acap edit video.mp4
acap edit video.mp4 -o output.mp4   # specify output path / 출력 경로 지정
acap edit video.mp4 --auto          # apply without confirmation / 확인 없이 자동 적용
```

**What gets removed / 제거 대상:**
- Long silences and awkward pauses / 긴 침묵 / 어색한 정지
- Repeated content / 반복되는 내용
- Mistakes and NG takes / 실수 · NG 구간
- Filler-heavy segments ("uh", "um", "어", "음") / 필러가 많은 구간
- Off-topic tangents / 주제와 관련 없는 잡담

**Example output / 실행 예시:**
```
🎬 acap - AI Video Editor
────────────────────────────────────────
[1/3] Extracting audio... done
[2/3] Transcribing (Whisper)... done (24 segments)
[3/3] AI analysis (DeepSeek)... done

📌 Summary: A React tutorial covering component basics

✂️  Suggested cuts (3):
  1. 4.0s ~ 7.2s  (3.2s)
  2. 43.5s ~ 48.0s (4.5s)
  3. 112.0s ~ 115.3s (3.3s)

💡 Reason: Opening silence, mid-video NG take, off-topic closing remarks

Apply edits? [y/N]
```

---

### subtitle — AI Subtitle Generation / AI 자막 생성

AI suggests what subtitle to show, where, and in what style.
AI가 자막 내용, 위치, 스타일을 제안합니다.

```bash
acap subtitle video.mp4                     # interactive / 대화형
acap subtitle video.mp4 --srt-only          # generate SRT file only / SRT 파일만 생성
acap subtitle video.mp4 --burn              # burn subtitles into video / 영상에 직접 굽기
acap subtitle video.mp4 --burn -o output.mp4
```

**What AI suggests / AI가 제안하는 것:**
- Exact timestamps for each subtitle / 몇 초 ~ 몇 초에 어떤 자막을 달지
- Position: top or bottom / 위치: 상단 / 하단
- Style: normal / ★ emphasis (key points) / 💬 caption (supplementary) / 스타일: 일반 / 강조 / 보충
- Auto-removes fillers ("uh", "um", "어", "음") / 필러 자동 제거
- Auto line-breaks for long sentences / 긴 문장 자동 줄 나누기

---

### transcribe — Speech to Text / 음성 → 텍스트

Runs STT only, without any AI analysis.
AI 분석 없이 STT 결과만 확인할 때 사용합니다.

```bash
acap transcribe video.mp4
```

```
[0.0s ~ 2.4s] Hello, today we're going to look at React.
[2.4s ~ 5.1s] Let's start by understanding what a component is.
...
```

---

## Project Structure / 구조

```
.
├── main.go
├── cmd/
│   ├── root.go        # CLI entry point
│   ├── edit.go        # edit command
│   ├── subtitle.go    # subtitle command
│   └── transcribe.go  # transcribe command
└── internal/
    ├── ffmpeg.go      # audio extraction / video editing
    ├── stt.go         # Whisper STT
    ├── ai.go          # DeepSeek edit analysis
    └── subtitle.go    # DeepSeek subtitle generation / SRT / burn
```

## Tech Stack / 기술 스택

- **Go** — CLI build
- **Cobra** — CLI framework
- **OpenAI Whisper** — local STT (free)
- **DeepSeek API** — AI edit & subtitle analysis
- **ffmpeg** — audio extraction / video editing / subtitle burning
