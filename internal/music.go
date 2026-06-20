package internal

import (
	"fmt"
	"os/exec"
	"strings"
)

// hasAudioStream reports whether the file contains at least one audio stream.
func hasAudioStream(path string) bool {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		path,
	).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// AddBackgroundMusic adds a background music track to a video.
//   - volume: music volume multiplier (0.0 ~ 1.0+)
//   - replace: replace the original audio entirely instead of mixing under it
//   - loop: loop the music to fill the whole video (trimmed to video length)
//
// The video stream is copied (no re-encode); only audio is processed.
func AddBackgroundMusic(videoPath, musicPath, outputPath string, volume float64, replace, loop bool) error {
	args := []string{"-y", "-i", videoPath}

	// -stream_loop는 바로 다음 -i 입력에 적용되므로 음악 입력 앞에 둔다.
	if loop {
		args = append(args, "-stream_loop", "-1")
	}
	args = append(args, "-i", musicPath)

	// 원본에 오디오가 없으면 믹스할 대상이 없으므로 항상 교체 동작.
	mix := !replace && hasAudioStream(videoPath)

	var filter string
	if mix {
		// 원본 음성은 그대로 두고(normalize=0) 음악만 volume 조절해서 깐다.
		filter = fmt.Sprintf(
			"[1:a]volume=%.3f[m];[0:a][m]amix=inputs=2:duration=first:dropout_transition=0:normalize=0[aout]",
			volume,
		)
	} else {
		// 교체(또는 원본 무음) 모드: 음악이 영상보다 짧을 때 -shortest가 영상까지
		// 자르지 않도록 apad로 무음을 채운다. -shortest가 영상 길이에서 끊는다.
		filter = fmt.Sprintf("[1:a]volume=%.3f,apad[aout]", volume)
	}

	args = append(args,
		"-filter_complex", filter,
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-shortest",
		outputPath,
	)

	cmd := newFFmpegCmd(args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("배경음악 추가 실패: %s", string(out))
	}
	return nil
}
