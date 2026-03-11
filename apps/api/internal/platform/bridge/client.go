package bridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/platform/config"
	"github.com/cachij/write-me/apps/api/internal/shared"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	token      string
	models     map[shared.ProviderType]string
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 90 * time.Second},
		baseURL:    cfg.BridgeBaseURL,
		token:      cfg.BridgeToken,
		models: map[shared.ProviderType]string{
			shared.ProviderCodexCLI:  cfg.CodexModelAlias,
			shared.ProviderGeminiCLI: cfg.GeminiModelAlias,
			shared.ProviderClaudeCLI: cfg.ClaudeModelAlias,
		},
	}
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Temperature float64       `json:"temperature,omitempty"`
	Messages    []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *Client) Health(ctx context.Context) (map[shared.ProviderType]bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bridge health failed: %s", strings.TrimSpace(string(body)))
	}

	var payload ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	available := make(map[shared.ProviderType]bool, len(c.models))
	seen := map[string]struct{}{}
	for _, model := range payload.Data {
		seen[model.ID] = struct{}{}
	}
	for provider, alias := range c.models {
		_, ok := seen[alias]
		available[provider] = ok
	}
	return available, nil
}

func (c *Client) CompleteJSON(ctx context.Context, provider shared.ProviderType, systemPrompt string, userPrompt string) (string, error) {
	modelAlias, ok := c.models[provider]
	if !ok {
		return "", shared.NewAppError(http.StatusBadRequest, "provider_not_supported", "지원하지 않는 provider입니다.")
	}

	requestPayload := ChatCompletionRequest{
		Model:       modelAlias,
		Temperature: 0.2,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	content, err := c.complete(ctx, requestPayload)
	if err == nil && json.Valid([]byte(content)) {
		return content, nil
	}

	repairPayload := ChatCompletionRequest{
		Model:       modelAlias,
		Temperature: 0,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
			{Role: "assistant", Content: content},
			{Role: "user", Content: "이전 응답이 JSON 스키마를 만족하지 않았습니다. 설명 없이 유효한 JSON 객체만 다시 반환하세요."},
		},
	}
	repaired, repairErr := c.complete(ctx, repairPayload)
	if repairErr != nil {
		if err != nil {
			return "", err
		}
		return "", repairErr
	}
	if !json.Valid([]byte(repaired)) {
		return "", shared.NewAppError(http.StatusBadGateway, "bridge_invalid_json", "bridge가 유효한 JSON을 반환하지 못했습니다.")
	}
	return repaired, nil
}

func (c *Client) complete(ctx context.Context, payload ChatCompletionRequest) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", shared.WrapAppError(http.StatusBadGateway, "bridge_unavailable", "bridge에 연결하지 못했습니다.", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", shared.WrapAppError(http.StatusBadGateway, "bridge_request_failed", strings.TrimSpace(string(raw)), nil)
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", shared.NewAppError(http.StatusBadGateway, "bridge_empty_response", "bridge 응답이 비어 있습니다.")
	}
	return flattenContent(parsed.Choices[0].Message.Content), nil
}

func flattenContent(content any) string {
	switch value := content.(type) {
	case string:
		return strings.TrimSpace(value)
	case []any:
		var builder strings.Builder
		for _, item := range value {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok {
				builder.WriteString(text)
			}
		}
		return strings.TrimSpace(builder.String())
	default:
		return ""
	}
}
