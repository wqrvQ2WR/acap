package cmd

import (
	"fmt"
	"strings"

	"vidai/internal"

	"github.com/spf13/cobra"
)

var (
	apiProvider string
	apiModel    string
	apiBaseURL  string
	apiKeyEnv   string
	apiList     bool
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "AI 모델/제공자를 설정합니다 (DeepSeek, OpenAI 등)",
	Long: `edit·subtitle 분석에 사용할 AI 제공자와 모델을 설정합니다.
OpenAI 호환 엔드포인트면 어떤 모델이든 쓸 수 있습니다.

설정은 ~/.config/acap/config.json 에 저장됩니다.

사용 예시:
  acap api                                  # 현재 설정 보기
  acap api --list                           # 사용 가능한 프리셋 보기
  acap api --provider openai                # OpenAI(gpt-4o)로 전환
  acap api --provider openai --model gpt-4o-mini
  acap api --provider deepseek              # DeepSeek로 전환
  acap api --base-url https://my-host/v1 --model my-model --key-env MY_API_KEY  # 커스텀`,
	Args: cobra.NoArgs,
	RunE: runAPI,
}

func init() {
	apiCmd.Flags().StringVarP(&apiProvider, "provider", "p", "", "프리셋 제공자 (deepseek, openai)")
	apiCmd.Flags().StringVarP(&apiModel, "model", "m", "", "모델 ID")
	apiCmd.Flags().StringVar(&apiBaseURL, "base-url", "", "OpenAI 호환 엔드포인트 URL")
	apiCmd.Flags().StringVar(&apiKeyEnv, "key-env", "", "API 키가 담긴 환경변수 이름")
	apiCmd.Flags().BoolVarP(&apiList, "list", "l", false, "사용 가능한 프리셋 목록 보기")
}

func runAPI(cmd *cobra.Command, args []string) error {
	if apiList {
		fmt.Println("사용 가능한 프리셋:")
		for _, name := range internal.PresetNames() {
			p := internal.Presets[name]
			fmt.Printf("  • %-10s → 모델 %s, 키 환경변수 %s\n", name, p.Model, p.APIKeyEnv)
		}
		return nil
	}

	// 아무 플래그도 없으면 현재 설정만 출력
	if apiProvider == "" && apiModel == "" && apiBaseURL == "" && apiKeyEnv == "" {
		fmt.Println("현재 AI 설정:")
		fmt.Println(internal.LoadConfig())
		fmt.Println("\n변경하려면: acap api --provider openai  (또는 --help 참고)")
		return nil
	}

	// 프리셋을 기준(base)으로 잡고, 나머지 플래그로 덮어쓴다.
	var cfg internal.Config
	if apiProvider != "" {
		preset, ok := internal.Presets[strings.ToLower(apiProvider)]
		if !ok {
			cfg = internal.LoadConfig()
			cfg.Provider = apiProvider // 커스텀 제공자 이름 허용
		} else {
			cfg = preset
		}
	} else {
		cfg = internal.LoadConfig()
	}

	if apiModel != "" {
		cfg.Model = apiModel
	}
	if apiBaseURL != "" {
		cfg.BaseURL = apiBaseURL
	}
	if apiKeyEnv != "" {
		cfg.APIKeyEnv = apiKeyEnv
	}

	if err := internal.SaveConfig(cfg); err != nil {
		return fmt.Errorf("설정 저장 실패: %v", err)
	}

	fmt.Println("✅ 설정이 저장되었습니다:")
	fmt.Println(cfg)
	return nil
}
