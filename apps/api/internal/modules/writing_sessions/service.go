package writing_sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/modules/application_specs"
	"github.com/cachij/write-me/apps/api/internal/modules/assets"
	"github.com/cachij/write-me/apps/api/internal/shared"
	"github.com/google/uuid"
)

type DraftQuestion struct {
	ID                    string          `json:"id"`
	QuestionID            string          `json:"questionId"`
	Title                 string          `json:"title"`
	PromptText            string          `json:"promptText"`
	CharLimit             int             `json:"charLimit"`
	Document              shared.Document `json:"document"`
	PlainText             string          `json:"plainText"`
	InferredCount         int             `json:"inferredCount"`
	ResolvedInferredCount int             `json:"resolvedInferredCount"`
	UpdatedAt             time.Time       `json:"updatedAt"`
}

type ChatMessage struct {
	ID          string         `json:"id"`
	QuestionID  string         `json:"questionId"`
	Role        string         `json:"role"`
	MessageType string         `json:"messageType"`
	Content     string         `json:"content"`
	Meta        map[string]any `json:"meta"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type Suggestion struct {
	ID                 string          `json:"id"`
	QuestionID         string          `json:"questionId"`
	Source             string          `json:"source"`
	Scope              string          `json:"scope"`
	Rationale          []string        `json:"rationale"`
	OriginalDocument   shared.Document `json:"originalDocument"`
	OriginalPlainText  string          `json:"originalPlainText"`
	SuggestedDocument  shared.Document `json:"suggestedDocument"`
	SuggestedPlainText string          `json:"suggestedPlainText"`
	Status             string          `json:"status"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type ReviewReport struct {
	ID               string            `json:"id"`
	SessionID        string            `json:"sessionId"`
	QuestionID       string            `json:"questionId,omitempty"`
	PriorityScores   map[string]int    `json:"priorityScores"`
	BlockingItems    []string          `json:"blockingItems"`
	TopActions       []string          `json:"topActions"`
	QuestionFindings []map[string]any  `json:"questionFindings"`
	UnresolvedClaims []shared.ClaimTag `json:"unresolvedClaims"`
	ReadyToSubmit    bool              `json:"readyToSubmit"`
	Raw              map[string]any    `json:"raw"`
	CreatedAt        time.Time         `json:"createdAt"`
}

type Session struct {
	ID              string                       `json:"id"`
	Title           string                       `json:"title"`
	Status          shared.SessionStatus         `json:"status"`
	CurrentProvider shared.ProviderType          `json:"currentProvider"`
	ApplyMode       shared.ApplyMode             `json:"applyMode"`
	ReviewMode      shared.ReviewMode            `json:"reviewMode"`
	AutoReview      bool                         `json:"autoReview"`
	AutoApply       bool                         `json:"autoApply"`
	CountMode       shared.CountMode             `json:"countMode"`
	ApplicationSpec application_specs.Spec       `json:"applicationSpec"`
	Assets          []assets.Asset               `json:"assets"`
	Drafts          []DraftQuestion              `json:"drafts"`
	ChatMessages    []ChatMessage                `json:"chatMessages"`
	LatestReview    *ReviewReport                `json:"latestReview,omitempty"`
	Pending         map[string]*Suggestion       `json:"pendingSuggestions"`
	ProviderHealth  map[shared.ProviderType]bool `json:"providerHealth"`
	CreatedAt       time.Time                    `json:"createdAt"`
	UpdatedAt       time.Time                    `json:"updatedAt"`
}

type DashboardSummary struct {
	RecentSessions []SessionCard `json:"recentSessions"`
	AssetCount     int           `json:"assetCount"`
	ReadyCount     int           `json:"readyCount"`
}

type SessionCard struct {
	ID              string               `json:"id"`
	Title           string               `json:"title"`
	CompanyName     string               `json:"companyName"`
	RoleName        string               `json:"roleName"`
	Status          shared.SessionStatus `json:"status"`
	UpdatedAt       time.Time            `json:"updatedAt"`
	UnresolvedCount int                  `json:"unresolvedCount"`
}

type CreateSessionInput struct {
	Title             string
	Provider          shared.ProviderType
	ApplicationSpecID string
	SelectedAssetIDs  []string
	InitialDraftText  string
}

type UpdateSessionInput struct {
	CurrentProvider shared.ProviderType
	AutoReview      bool
	AutoApply       bool
	CountMode       shared.CountMode
	ApplyMode       shared.ApplyMode
}

type UpdateDraftInput struct {
	SessionID  string
	QuestionID string
	PlainText  string
	Document   *shared.Document
	Reason     string
}

type AssistantSuggestion struct {
	AssistantMessage  string           `json:"assistant_message"`
	DraftText         string           `json:"draft_text"`
	SupportedClaims   []shared.AIClaim `json:"supported_claims"`
	InferredClaims    []shared.AIClaim `json:"inferred_claims"`
	FollowUpQuestions []string         `json:"follow_up_questions"`
	Rationale         []string         `json:"rationale"`
}

type ReviewPayload struct {
	PriorityScores   map[string]int   `json:"priority_scores"`
	BlockingItems    []string         `json:"blocking_items"`
	TopActions       []string         `json:"top_actions"`
	QuestionFindings []map[string]any `json:"question_findings"`
	ReadyToSubmit    bool             `json:"ready_to_submit"`
}

type AssistantProvider interface {
	Health(ctx context.Context) (map[shared.ProviderType]bool, error)
	GenerateDraft(ctx context.Context, provider shared.ProviderType, session Session, question DraftQuestion) (AssistantSuggestion, error)
	SuggestEdit(ctx context.Context, provider shared.ProviderType, session Session, question DraftQuestion, scope string, instruction string) (AssistantSuggestion, error)
	ReviewDraft(ctx context.Context, provider shared.ProviderType, session Session, question DraftQuestion) (ReviewPayload, error)
}

type Repository interface {
	CreateSession(ctx context.Context, input CreateSessionInput, spec application_specs.Spec) (Session, error)
	GetSession(ctx context.Context, sessionID string) (Session, error)
	ListDashboard(ctx context.Context) (DashboardSummary, error)
	UpdateDraft(ctx context.Context, input UpdateDraftInput) (DraftQuestion, error)
	CreateVersionSnapshot(ctx context.Context, sessionID string, questionID string, reason string, draft DraftQuestion) error
	CreateSuggestion(ctx context.Context, sessionID string, suggestion Suggestion) (Suggestion, error)
	GetLatestSuggestion(ctx context.Context, sessionID string, questionID string) (*Suggestion, error)
	ApplySuggestion(ctx context.Context, suggestionID string) (DraftQuestion, error)
	CreateChatMessage(ctx context.Context, sessionID string, message ChatMessage) (ChatMessage, error)
	CreateToolAction(ctx context.Context, sessionID string, questionID string, toolName string, promptText string, chatMessageID string) error
	CreateReview(ctx context.Context, sessionID string, report ReviewReport) (ReviewReport, error)
	GetLatestReview(ctx context.Context, sessionID string) (*ReviewReport, error)
	ResolveClaimTag(ctx context.Context, sessionID string, questionID string, claimID string) (DraftQuestion, error)
	FinalizeSession(ctx context.Context, sessionID string) error
	UpdateSession(ctx context.Context, sessionID string, input UpdateSessionInput) (Session, error)
}

type Service struct {
	repo      Repository
	assistant AssistantProvider
}

func NewService(repo Repository, assistant AssistantProvider) *Service {
	return &Service{repo: repo, assistant: assistant}
}

func (s *Service) CreateSession(ctx context.Context, input CreateSessionInput, spec application_specs.Spec) (Session, error) {
	if input.Provider == "" {
		input.Provider = shared.ProviderCodexCLI
	}
	session, err := s.repo.CreateSession(ctx, input, spec)
	if err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(input.InitialDraftText) == "" || len(session.Drafts) == 0 {
		return session, nil
	}
	firstDraft := session.Drafts[0]
	_, err = s.repo.UpdateDraft(ctx, UpdateDraftInput{
		SessionID:  session.ID,
		QuestionID: firstDraft.QuestionID,
		PlainText:  input.InitialDraftText,
		Reason:     "initial_seed",
	})
	if err != nil {
		return Session{}, err
	}
	return s.repo.GetSession(ctx, session.ID)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (Session, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}
	if health, healthErr := s.assistant.Health(ctx); healthErr == nil {
		session.ProviderHealth = health
	} else {
		session.ProviderHealth = map[shared.ProviderType]bool{}
	}
	return session, nil
}

func (s *Service) Dashboard(ctx context.Context) (DashboardSummary, error) {
	return s.repo.ListDashboard(ctx)
}

func (s *Service) UpdateDraft(ctx context.Context, input UpdateDraftInput) (DraftQuestion, error) {
	return s.repo.UpdateDraft(ctx, input)
}

func (s *Service) UpdateSession(ctx context.Context, sessionID string, input UpdateSessionInput) (Session, error) {
	return s.repo.UpdateSession(ctx, sessionID, input)
}

func (s *Service) GenerateDraft(ctx context.Context, sessionID string, questionID string) (*Suggestion, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	question, err := findDraft(session.Drafts, questionID)
	if err != nil {
		return nil, err
	}
	result, err := s.assistant.GenerateDraft(ctx, session.CurrentProvider, session, question)
	if err != nil {
		return nil, err
	}
	return s.storeSuggestion(ctx, session, question, result, "generate", "QUESTION")
}

func (s *Service) SendChatTurn(ctx context.Context, sessionID string, questionID string, scope string, userMessage string) (ChatMessage, *Suggestion, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return ChatMessage{}, nil, err
	}
	question, err := findDraft(session.Drafts, questionID)
	if err != nil {
		return ChatMessage{}, nil, err
	}

	chatMessage := ChatMessage{
		ID:          uuid.NewString(),
		QuestionID:  questionID,
		Role:        "user",
		MessageType: "user",
		Content:     strings.TrimSpace(userMessage),
		Meta: map[string]any{
			"scope": scope,
		},
	}
	if _, err := s.repo.CreateChatMessage(ctx, sessionID, chatMessage); err != nil {
		return ChatMessage{}, nil, err
	}

	result, err := s.assistant.SuggestEdit(ctx, session.CurrentProvider, session, question, scope, userMessage)
	if err != nil {
		return ChatMessage{}, nil, err
	}

	assistantMessage := ChatMessage{
		ID:          uuid.NewString(),
		QuestionID:  questionID,
		Role:        "assistant",
		MessageType: "assistant",
		Content:     strings.TrimSpace(result.AssistantMessage),
		Meta: map[string]any{
			"followUpQuestions": result.FollowUpQuestions,
		},
	}
	createdAssistantMessage, err := s.repo.CreateChatMessage(ctx, sessionID, assistantMessage)
	if err != nil {
		return ChatMessage{}, nil, err
	}

	var suggestion *Suggestion
	if strings.TrimSpace(result.DraftText) != "" && shared.NormalizeText(result.DraftText) != shared.NormalizeText(question.PlainText) {
		suggestion, err = s.storeSuggestion(ctx, session, question, result, "chat", scope)
		if err != nil {
			return ChatMessage{}, nil, err
		}
	}
	return createdAssistantMessage, suggestion, nil
}

func (s *Service) ExecuteToolAction(ctx context.Context, sessionID string, questionID string, toolName string, scope string) (ChatMessage, *Suggestion, error) {
	prompt, ok := toolPrompts[toolName]
	if !ok {
		return ChatMessage{}, nil, shared.NewAppError(http.StatusBadRequest, "tool_not_supported", "지원하지 않는 도구입니다.")
	}
	message, suggestion, err := s.SendChatTurn(ctx, sessionID, questionID, scope, prompt)
	if err == nil {
		_ = s.repo.CreateToolAction(ctx, sessionID, questionID, toolName, prompt, message.ID)
	}
	return message, suggestion, err
}

func (s *Service) GetLatestSuggestion(ctx context.Context, sessionID string, questionID string) (*Suggestion, error) {
	return s.repo.GetLatestSuggestion(ctx, sessionID, questionID)
}

func (s *Service) ApplySuggestion(ctx context.Context, suggestionID string) (DraftQuestion, error) {
	return s.repo.ApplySuggestion(ctx, suggestionID)
}

func (s *Service) RunReview(ctx context.Context, sessionID string, questionID string) (ReviewReport, error) {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return ReviewReport{}, err
	}

	drafts := session.Drafts
	if questionID != "" {
		target, err := findDraft(session.Drafts, questionID)
		if err != nil {
			return ReviewReport{}, err
		}
		drafts = []DraftQuestion{target}
	}

	report := buildLocalReview(session, drafts)
	if len(drafts) == 1 {
		if aiReport, err := s.assistant.ReviewDraft(ctx, session.CurrentProvider, session, drafts[0]); err == nil {
			report.PriorityScores = aiReport.PriorityScores
			report.BlockingItems = aiReport.BlockingItems
			report.TopActions = aiReport.TopActions
			report.QuestionFindings = aiReport.QuestionFindings
			report.ReadyToSubmit = aiReport.ReadyToSubmit && len(report.UnresolvedClaims) == 0
			report.Raw = map[string]any{
				"assistant": aiReport,
			}
		}
	}

	return s.repo.CreateReview(ctx, sessionID, report)
}

func (s *Service) ResolveClaimTag(ctx context.Context, sessionID string, questionID string, claimID string) (DraftQuestion, error) {
	return s.repo.ResolveClaimTag(ctx, sessionID, questionID, claimID)
}

func (s *Service) GetLatestReview(ctx context.Context, sessionID string) (*ReviewReport, error) {
	return s.repo.GetLatestReview(ctx, sessionID)
}

func (s *Service) FinalizeSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	for _, draft := range session.Drafts {
		if draft.InferredCount > draft.ResolvedInferredCount {
			return shared.NewAppError(http.StatusBadRequest, "unresolved_inferred_claims", "AI 추론 문장을 모두 확인해야 최종 확정할 수 있습니다.")
		}
	}
	return s.repo.FinalizeSession(ctx, sessionID)
}

func (s *Service) storeSuggestion(ctx context.Context, session Session, question DraftQuestion, result AssistantSuggestion, source string, scope string) (*Suggestion, error) {
	originalDocument := question.Document
	suggestedDocument := shared.NewDocumentFromText(result.DraftText)
	suggestedDocument.TagClaims(shared.ClaimTagSupported, result.SupportedClaims)
	suggestedDocument.TagClaims(shared.ClaimTagInferred, result.InferredClaims)

	suggestion := Suggestion{
		ID:                 uuid.NewString(),
		QuestionID:         question.QuestionID,
		Source:             source,
		Scope:              scope,
		Rationale:          result.Rationale,
		OriginalDocument:   originalDocument,
		OriginalPlainText:  question.PlainText,
		SuggestedDocument:  suggestedDocument,
		SuggestedPlainText: suggestedDocument.PlainText(),
		Status:             "PENDING",
		CreatedAt:          time.Now(),
	}
	created, err := s.repo.CreateSuggestion(ctx, session.ID, suggestion)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func buildLocalReview(session Session, drafts []DraftQuestion) ReviewReport {
	report := ReviewReport{
		ID:               uuid.NewString(),
		SessionID:        session.ID,
		PriorityScores:   map[string]int{"question_fit": 70, "evidence": 70, "readability": 70},
		BlockingItems:    []string{},
		TopActions:       []string{},
		QuestionFindings: []map[string]any{},
		UnresolvedClaims: []shared.ClaimTag{},
		Raw:              map[string]any{},
		CreatedAt:        time.Now(),
	}

	if len(drafts) == 1 {
		report.QuestionID = drafts[0].QuestionID
	}

	for _, draft := range drafts {
		charCount := draft.Document.CharCount(session.CountMode)
		if draft.CharLimit > 0 && charCount > draft.CharLimit {
			report.BlockingItems = append(report.BlockingItems, fmt.Sprintf("%s 글자 수가 제한을 초과했습니다.", draft.Title))
		}
		if draft.InferredCount > draft.ResolvedInferredCount {
			report.BlockingItems = append(report.BlockingItems, fmt.Sprintf("%s에 미확정 AI 추론 문장이 남아 있습니다.", draft.Title))
		}

		findings := map[string]any{
			"questionId":    draft.QuestionID,
			"title":         draft.Title,
			"charCount":     charCount,
			"charLimit":     draft.CharLimit,
			"inferredCount": draft.InferredCount,
		}
		report.QuestionFindings = append(report.QuestionFindings, findings)

		for _, block := range draft.Document.Blocks {
			for _, tag := range block.ClaimTags {
				if tag.Status == shared.ClaimTagInferred && !tag.Resolved {
					report.UnresolvedClaims = append(report.UnresolvedClaims, tag)
				}
			}
		}
	}

	if len(report.BlockingItems) == 0 {
		report.TopActions = append(report.TopActions, "문항별 핵심 강점이 첫 문장에 드러나는지 다시 확인하세요.")
		report.TopActions = append(report.TopActions, "중복되는 표현을 줄이고 숫자나 결과를 한 번 더 보강하세요.")
	} else {
		report.TopActions = append(report.TopActions, "차단 이슈부터 해결한 뒤 다시 검토하세요.")
	}

	report.PriorityScores["question_fit"] = clampScore(80 - len(report.BlockingItems)*8)
	report.PriorityScores["evidence"] = clampScore(85 - len(report.UnresolvedClaims)*10)
	report.PriorityScores["readability"] = clampScore(78)
	report.ReadyToSubmit = len(report.BlockingItems) == 0 && len(report.UnresolvedClaims) == 0
	return report
}

func findDraft(drafts []DraftQuestion, questionID string) (DraftQuestion, error) {
	for _, draft := range drafts {
		if draft.QuestionID == questionID {
			return draft, nil
		}
	}
	return DraftQuestion{}, shared.NewAppError(http.StatusNotFound, "question_not_found", "문항을 찾을 수 없습니다.")
}

func clampScore(value int) int {
	return int(math.Max(0, math.Min(100, float64(value))))
}

func ParseStoredReview(raw []byte) (ReviewReport, error) {
	var report ReviewReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return ReviewReport{}, err
	}
	sort.SliceStable(report.QuestionFindings, func(i, j int) bool {
		return fmt.Sprint(report.QuestionFindings[i]["questionId"]) < fmt.Sprint(report.QuestionFindings[j]["questionId"])
	})
	return report, nil
}

var toolPrompts = map[string]string{
	"다른 소재 찾기":  "현재 문항과 선택 범위를 유지한 채, 자료에 있는 경험 중 대체 가능한 소재 3개를 제안하세요. 자료에 없는 사실은 쓰지 말고, 없으면 사용자에게 질문하세요.",
	"질문 적합성 개선": "선택 범위만 수정하고, 현재 문항이 요구하는 역량에 더 직접 답하도록 재작성하세요. 범위 밖 문장은 건드리지 마세요.",
	"가독성 개선":    "선택 범위만 유지한 채 문장을 더 짧고 명확하게 다듬으세요. 의미가 바뀌면 안 됩니다.",
	"분량 맞추기":    "현재 문항의 글자 수 제한 안에 들어오도록 선택 범위를 압축하거나 확장하세요. 핵심 근거는 유지하세요.",
	"톤 조정":      "과장 없이 자신감 있는 톤으로 선택 범위를 다듬으세요. 추상 표현보다 검증 가능한 표현을 우선하세요.",
}
