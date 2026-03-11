package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cachij/write-me/apps/api/internal/shared"
)

type Config struct {
	Port            int
	DatabaseURL     string
	AppOrigin       string
	APIPublicOrigin string
	SessionSecret   string
	AdminEmail      string
	AdminPassword   string

	BridgeBaseURL string
	BridgeToken   string

	DefaultProvider  shared.ProviderType
	CodexModelAlias  string
	GeminiModelAlias string
	ClaudeModelAlias string
}

func Load() (Config, error) {
	cfg := Config{
		Port:             getEnvInt("API_PORT", 8080),
		DatabaseURL:      strings.TrimSpace(os.Getenv("DATABASE_URL")),
		AppOrigin:        getEnv("APP_ORIGIN", "http://localhost:3000"),
		APIPublicOrigin:  getEnv("API_PUBLIC_ORIGIN", "http://localhost:8080"),
		SessionSecret:    strings.TrimSpace(os.Getenv("SESSION_SECRET")),
		AdminEmail:       strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		AdminPassword:    strings.TrimSpace(os.Getenv("ADMIN_PASSWORD")),
		BridgeBaseURL:    strings.TrimRight(getEnv("LLM_BRIDGE_BASE_URL", "http://host.docker.internal:43110"), "/"),
		BridgeToken:      strings.TrimSpace(os.Getenv("LLM_BRIDGE_TOKEN")),
		DefaultProvider:  shared.ProviderType(getEnv("DEFAULT_PROVIDER", string(shared.ProviderCodexCLI))),
		CodexModelAlias:  getEnv("CODEX_MODEL_ALIAS", "write-me-codex"),
		GeminiModelAlias: getEnv("GEMINI_MODEL_ALIAS", "write-me-gemini"),
		ClaudeModelAlias: getEnv("CLAUDE_MODEL_ALIAS", "write-me-claude"),
	}

	var missing []string
	for key, value := range map[string]string{
		"DATABASE_URL":     cfg.DatabaseURL,
		"SESSION_SECRET":   cfg.SessionSecret,
		"ADMIN_EMAIL":      cfg.AdminEmail,
		"ADMIN_PASSWORD":   cfg.AdminPassword,
		"LLM_BRIDGE_TOKEN": cfg.BridgeToken,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
