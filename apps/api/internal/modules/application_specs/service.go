package application_specs

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/shared"
	"github.com/google/uuid"
)

type Question struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	PromptText string `json:"promptText"`
	CharLimit  int    `json:"charLimit"`
}

type Spec struct {
	ID          string     `json:"id"`
	CompanyName string     `json:"companyName"`
	RoleName    string     `json:"roleName"`
	SourceText  string     `json:"sourceText"`
	Warnings    []string   `json:"warnings"`
	Questions   []Question `json:"questions"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type Repository interface {
	CreateSpec(ctx context.Context, spec Spec) (Spec, error)
	GetSpec(ctx context.Context, specID string) (Spec, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ParseAndStore(ctx context.Context, sourceText string) (Spec, error) {
	sourceText = shared.NormalizeText(sourceText)
	if len([]rune(sourceText)) < 30 {
		return Spec{}, shared.NewAppError(http.StatusBadRequest, "posting_too_short", "지원 페이지 텍스트가 너무 짧습니다.")
	}
	spec := ParseText(sourceText)
	return s.repo.CreateSpec(ctx, spec)
}

func (s *Service) CreateManual(ctx context.Context, companyName string, roleName string, questions []Question, sourceText string) (Spec, error) {
	spec := Spec{
		ID:          uuid.NewString(),
		CompanyName: strings.TrimSpace(companyName),
		RoleName:    strings.TrimSpace(roleName),
		SourceText:  shared.NormalizeText(sourceText),
		Warnings:    []string{},
		Questions:   normalizeQuestions(questions),
	}
	if len(spec.Questions) == 0 {
		spec.Questions = []Question{
			{
				ID:         "q1",
				Title:      "자유 자기소개",
				PromptText: "자유 형식 자기소개를 작성하세요.",
				CharLimit:  700,
			},
		}
	}
	return s.repo.CreateSpec(ctx, spec)
}

func (s *Service) Get(ctx context.Context, specID string) (Spec, error) {
	return s.repo.GetSpec(ctx, specID)
}

func ParseText(sourceText string) Spec {
	lines := splitLines(sourceText)
	companyName := findPrefixedValue(lines, []string{"회사", "기업", "company"})
	roleName := findPrefixedValue(lines, []string{"직무", "포지션", "role"})
	questions := parseQuestions(sourceText)
	warnings := []string{}

	if companyName == "" {
		warnings = append(warnings, "회사명을 자동으로 확정하지 못했습니다.")
		companyName = "미확정"
	}
	if roleName == "" {
		warnings = append(warnings, "직무명을 자동으로 확정하지 못했습니다.")
		roleName = "미확정"
	}
	if len(questions) == 0 {
		warnings = append(warnings, "문항을 구조화하지 못해 자유 형식 문항 1개로 생성했습니다.")
		questions = []Question{
			{
				ID:         "q1",
				Title:      "자유 자기소개",
				PromptText: sourceText,
				CharLimit:  detectCharLimit(sourceText),
			},
		}
	}

	return Spec{
		ID:          uuid.NewString(),
		CompanyName: companyName,
		RoleName:    roleName,
		SourceText:  sourceText,
		Warnings:    warnings,
		Questions:   questions,
	}
}

func normalizeQuestions(questions []Question) []Question {
	result := make([]Question, 0, len(questions))
	for idx, question := range questions {
		title := strings.TrimSpace(question.Title)
		prompt := strings.TrimSpace(question.PromptText)
		if title == "" {
			title = fmt.Sprintf("문항 %d", idx+1)
		}
		if prompt == "" {
			prompt = title
		}
		id := strings.TrimSpace(question.ID)
		if id == "" {
			id = fmt.Sprintf("q%d", idx+1)
		}
		result = append(result, Question{
			ID:         id,
			Title:      title,
			PromptText: prompt,
			CharLimit:  question.CharLimit,
		})
	}
	return result
}

func parseQuestions(text string) []Question {
	re := regexp.MustCompile(`(?m)^\s*(\d+)[\.\)]\s+(.+?)(?:\((\d{1,4})\s*자\))?\s*$`)
	matches := re.FindAllStringSubmatch(text, -1)
	var questions []Question
	for idx, match := range matches {
		charLimit := 0
		if len(match) > 3 && match[3] != "" {
			charLimit, _ = strconv.Atoi(match[3])
		}
		if charLimit == 0 {
			charLimit = detectCharLimit(match[2])
		}
		questions = append(questions, Question{
			ID:         fmt.Sprintf("q%d", idx+1),
			Title:      fmt.Sprintf("문항 %d", idx+1),
			PromptText: strings.TrimSpace(match[2]),
			CharLimit:  charLimit,
		})
	}
	return questions
}

func detectCharLimit(text string) int {
	re := regexp.MustCompile(`(\d{2,4})\s*자`)
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		value, _ := strconv.Atoi(match[1])
		if value > 0 {
			return value
		}
	}
	if strings.Contains(text, "2~4문장") {
		return 200
	}
	return 700
}

func findPrefixedValue(lines []string, prefixes []string) string {
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, prefix := range prefixes {
			if !strings.Contains(lower, strings.ToLower(prefix)) {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func splitLines(text string) []string {
	raw := strings.Split(shared.NormalizeText(text), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		clean := strings.TrimSpace(line)
		if clean != "" {
			lines = append(lines, clean)
		}
	}
	return lines
}
