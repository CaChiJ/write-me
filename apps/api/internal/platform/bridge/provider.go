package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cachij/write-me/apps/api/internal/modules/writing_sessions"
	"github.com/cachij/write-me/apps/api/internal/shared"
)

type AssistantProvider struct {
	client *Client
}

func NewAssistantProvider(client *Client) *AssistantProvider {
	return &AssistantProvider{client: client}
}

func (p *AssistantProvider) Health(ctx context.Context) (map[shared.ProviderType]bool, error) {
	return p.client.Health(ctx)
}

func (p *AssistantProvider) GenerateDraft(ctx context.Context, provider shared.ProviderType, session writing_sessions.Session, question writing_sessions.DraftQuestion) (writing_sessions.AssistantSuggestion, error) {
	systemPrompt := `당신은 한국어 자기소개서 작성 보조 AI입니다. 자료에 없는 사실은 단정하지 말고, 필요하면 follow_up_questions에 질문을 남기세요. 설명 없이 JSON만 반환하세요.`
	userPrompt := fmt.Sprintf(`
다음 문항에 대한 자기소개서 초안을 작성하세요.

회사: %s
직무: %s
문항 제목: %s
문항 본문: %s
글자 수 제한: %d

사용자 자료:
%s

반환 JSON 스키마:
{
  "assistant_message": "사용자에게 보여줄 짧은 안내",
  "draft_text": "초안 본문",
  "supported_claims": [{"excerpt": "초안 안의 문장 일부", "source_asset_ids": ["자료 id"]}],
  "inferred_claims": [{"excerpt": "자료엔 없지만 AI가 보강한 문장 일부", "reason": "왜 추론했는지"}],
  "follow_up_questions": ["추가로 물어볼 질문"],
  "rationale": ["수정 의도 또는 작성 근거"]
}
`, session.ApplicationSpec.CompanyName, session.ApplicationSpec.RoleName, question.Title, question.PromptText, question.CharLimit, buildAssetContext(session))

	raw, err := p.client.CompleteJSON(ctx, provider, systemPrompt, userPrompt)
	if err != nil {
		return writing_sessions.AssistantSuggestion{}, err
	}
	return decodeSuggestion(raw)
}

func (p *AssistantProvider) SuggestEdit(ctx context.Context, provider shared.ProviderType, session writing_sessions.Session, question writing_sessions.DraftQuestion, scope string, instruction string) (writing_sessions.AssistantSuggestion, error) {
	systemPrompt := `당신은 한국어 자기소개서 편집 AI입니다. 사용자가 지정한 범위를 우선하고, 자료에 없는 사실은 추가하지 마세요. 변경이 꼭 필요하지 않다면 draft_text를 기존과 동일하게 반환하세요. 설명 없이 JSON만 반환하세요.`
	userPrompt := fmt.Sprintf(`
현재 문항과 초안 일부를 바탕으로 수정 제안을 만드세요.

회사: %s
직무: %s
문항 제목: %s
문항 본문: %s
글자 수 제한: %d
적용 범위: %s
사용자 요청: %s

현재 초안:
%s

사용자 자료:
%s

반환 JSON 스키마:
{
  "assistant_message": "수정 내용을 짧게 설명",
  "draft_text": "수정안 전체 본문",
  "supported_claims": [{"excerpt": "수정안 안의 문장 일부", "source_asset_ids": ["자료 id"]}],
  "inferred_claims": [{"excerpt": "자료엔 없지만 AI가 보강한 문장 일부", "reason": "왜 추론했는지"}],
  "follow_up_questions": ["추가 질문"],
  "rationale": ["왜 이렇게 바꿨는지"]
}
`, session.ApplicationSpec.CompanyName, session.ApplicationSpec.RoleName, question.Title, question.PromptText, question.CharLimit, scope, instruction, question.PlainText, buildAssetContext(session))

	raw, err := p.client.CompleteJSON(ctx, provider, systemPrompt, userPrompt)
	if err != nil {
		return writing_sessions.AssistantSuggestion{}, err
	}
	return decodeSuggestion(raw)
}

func (p *AssistantProvider) ReviewDraft(ctx context.Context, provider shared.ProviderType, session writing_sessions.Session, question writing_sessions.DraftQuestion) (writing_sessions.ReviewPayload, error) {
	systemPrompt := `당신은 자기소개서 검토 AI입니다. 질문 적합성, 근거성, 가독성을 중심으로 검토하고, JSON만 반환하세요.`
	userPrompt := fmt.Sprintf(`
다음 초안을 검토하세요.

회사: %s
직무: %s
문항 제목: %s
문항 본문: %s
글자 수 제한: %d
현재 글자 수 기준: %s
초안:
%s

사용자 자료:
%s

반환 JSON 스키마:
{
  "priority_scores": {
    "question_fit": 0,
    "evidence": 0,
    "readability": 0
  },
  "blocking_items": ["바로 고쳐야 할 문제"],
  "top_actions": ["가장 먼저 할 수정"],
  "question_findings": [{"summary": "문항별 검토 요약"}],
  "ready_to_submit": false
}
`, session.ApplicationSpec.CompanyName, session.ApplicationSpec.RoleName, question.Title, question.PromptText, question.CharLimit, session.CountMode, question.PlainText, buildAssetContext(session))

	raw, err := p.client.CompleteJSON(ctx, provider, systemPrompt, userPrompt)
	if err != nil {
		return writing_sessions.ReviewPayload{}, err
	}
	var payload writing_sessions.ReviewPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return writing_sessions.ReviewPayload{}, shared.WrapAppError(http.StatusBadGateway, "bridge_invalid_json", "검토 결과를 해석하지 못했습니다.", err)
	}
	if payload.PriorityScores == nil {
		payload.PriorityScores = map[string]int{}
	}
	return payload, nil
}

func decodeSuggestion(raw string) (writing_sessions.AssistantSuggestion, error) {
	var payload writing_sessions.AssistantSuggestion
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return writing_sessions.AssistantSuggestion{}, shared.WrapAppError(http.StatusBadGateway, "bridge_invalid_json", "AI 응답을 해석하지 못했습니다.", err)
	}
	return payload, nil
}

func buildAssetContext(session writing_sessions.Session) string {
	const maxChars = 16000
	var builder strings.Builder
	used := 0
	for _, asset := range session.Assets {
		segment := fmt.Sprintf("[자료:%s|%s]\n%s\n\n", asset.ID, asset.Title, strings.TrimSpace(asset.ExtractedText))
		if used+len([]rune(segment)) > maxChars {
			break
		}
		builder.WriteString(segment)
		used += len([]rune(segment))
	}
	if builder.Len() == 0 {
		return "연결된 자료 없음"
	}
	return builder.String()
}
