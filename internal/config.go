package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Config holds the AI provider settings used for edit/subtitle analysis.
type Config struct {
	Provider  string `json:"provider"`    // 프리셋 이름 또는 "custom"
	Model     string `json:"model"`       // 모델 ID (예: deepseek-chat, gpt-4o)
	BaseURL   string `json:"base_url"`    // OpenAI 호환 엔드포인트
	APIKeyEnv string `json:"api_key_env"` // API 키가 담긴 환경변수 이름
}

// Presets are built-in OpenAI-compatible providers.
var Presets = map[string]Config{
	"deepseek": {
		Provider:  "deepseek",
		Model:     "deepseek-chat",
		BaseURL:   "https://api.deepseek.com/v1",
		APIKeyEnv: "DEEPSEEK_API_KEY",
	},
	"openai": {
		Provider:  "openai",
		Model:     "gpt-4o",
		BaseURL:   "https://api.openai.com/v1",
		APIKeyEnv: "OPENAI_API_KEY",
	},
}

// DefaultConfig returns the fallback config when nothing is saved yet.
func DefaultConfig() Config {
	return Presets["deepseek"]
}

// PresetNames returns the available preset names, sorted.
func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for n := range Presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "acap", "config.json"), nil
}

// LoadConfig reads the saved config, falling back to the default if missing.
func LoadConfig() Config {
	path, err := configPath()
	if err != nil {
		return DefaultConfig()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}
	// 손상되거나 비어있는 필드는 기본값으로 보정
	if cfg.Model == "" || cfg.APIKeyEnv == "" {
		return DefaultConfig()
	}
	return cfg
}

// SaveConfig persists the config to ~/.config/acap/config.json.
func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// String renders a human-readable summary of the config.
func (c Config) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  제공자(provider) : %s\n", c.Provider))
	sb.WriteString(fmt.Sprintf("  모델(model)      : %s\n", c.Model))
	sb.WriteString(fmt.Sprintf("  엔드포인트       : %s\n", c.BaseURL))
	keyState := "❌ 미설정"
	if os.Getenv(c.APIKeyEnv) != "" {
		keyState = "✅ 설정됨"
	}
	sb.WriteString(fmt.Sprintf("  API 키 환경변수  : %s (%s)", c.APIKeyEnv, keyState))
	return sb.String()
}
