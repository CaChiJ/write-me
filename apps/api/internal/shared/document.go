package shared

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type ProviderType string

const (
	ProviderCodexCLI  ProviderType = "CODEX_CLI"
	ProviderGeminiCLI ProviderType = "GEMINI_CLI"
	ProviderClaudeCLI ProviderType = "CLAUDE_CLI"
)

type CountMode string

const (
	CountModeIncludeSpaces CountMode = "INCLUDE_SPACES"
	CountModeExcludeSpaces CountMode = "EXCLUDE_SPACES"
)

type ClaimTagStatus string

const (
	ClaimTagSupported ClaimTagStatus = "SUPPORTED"
	ClaimTagInferred  ClaimTagStatus = "INFERRED"
	ClaimTagResolved  ClaimTagStatus = "RESOLVED"
)

type ApplyMode string

const (
	ApplyModeManual ApplyMode = "MANUAL_REVIEW"
	ApplyModeAuto   ApplyMode = "AUTO_APPLY"
)

type ReviewMode string

const (
	ReviewModeManual ReviewMode = "MANUAL"
	ReviewModeAuto   ReviewMode = "AUTO"
	ReviewModeBoth   ReviewMode = "BOTH"
)

type SessionStatus string

const (
	SessionStatusDraft     SessionStatus = "DRAFT"
	SessionStatusFinalized SessionStatus = "FINALIZED"
)

type Document struct {
	Version int             `json:"version"`
	Blocks  []DocumentBlock `json:"blocks"`
}

type DocumentBlock struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	ClaimTags []ClaimTag `json:"claimTags"`
}

type ClaimTag struct {
	ID             string         `json:"id"`
	Start          int            `json:"start"`
	End            int            `json:"end"`
	Status         ClaimTagStatus `json:"status"`
	Excerpt        string         `json:"excerpt"`
	Reason         string         `json:"reason,omitempty"`
	SourceAssetIDs []string       `json:"sourceAssetIds,omitempty"`
	Resolved       bool           `json:"resolved"`
}

type AIClaim struct {
	Excerpt        string   `json:"excerpt"`
	Reason         string   `json:"reason,omitempty"`
	SourceAssetIDs []string `json:"source_asset_ids,omitempty"`
}

func NewDocumentFromText(text string) Document {
	normalized := NormalizeText(text)
	if normalized == "" {
		return Document{
			Version: 1,
			Blocks: []DocumentBlock{
				{
					ID:        uuid.NewString(),
					Text:      "",
					ClaimTags: []ClaimTag{},
				},
			},
		}
	}

	rawBlocks := splitParagraphs(normalized)
	blocks := make([]DocumentBlock, 0, len(rawBlocks))
	for _, blockText := range rawBlocks {
		blocks = append(blocks, DocumentBlock{
			ID:        uuid.NewString(),
			Text:      blockText,
			ClaimTags: []ClaimTag{},
		})
	}

	return Document{
		Version: 1,
		Blocks:  blocks,
	}
}

func NormalizeText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	return text
}

func (d Document) PlainText() string {
	parts := make([]string, 0, len(d.Blocks))
	for _, block := range d.Blocks {
		parts = append(parts, strings.TrimSpace(block.Text))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func (d Document) CharCount(mode CountMode) int {
	text := d.PlainText()
	if mode == CountModeExcludeSpaces {
		text = strings.ReplaceAll(text, " ", "")
		text = strings.ReplaceAll(text, "\n", "")
		text = strings.ReplaceAll(text, "\t", "")
	}
	return len([]rune(text))
}

func (d *Document) TagClaims(status ClaimTagStatus, claims []AIClaim) {
	for _, claim := range claims {
		excerpt := strings.TrimSpace(claim.Excerpt)
		if excerpt == "" {
			continue
		}
		for blockIdx := range d.Blocks {
			block := &d.Blocks[blockIdx]
			start := strings.Index(block.Text, excerpt)
			if start < 0 {
				continue
			}
			block.ClaimTags = append(block.ClaimTags, ClaimTag{
				ID:             uuid.NewString(),
				Start:          start,
				End:            start + len([]rune(excerpt)),
				Status:         status,
				Excerpt:        excerpt,
				Reason:         claim.Reason,
				SourceAssetIDs: claim.SourceAssetIDs,
				Resolved:       status != ClaimTagInferred,
			})
			break
		}
	}
}

func (d *Document) ResolveClaim(claimID string) bool {
	for blockIdx := range d.Blocks {
		for tagIdx := range d.Blocks[blockIdx].ClaimTags {
			tag := &d.Blocks[blockIdx].ClaimTags[tagIdx]
			if tag.ID != claimID {
				continue
			}
			tag.Status = ClaimTagResolved
			tag.Resolved = true
			return true
		}
	}
	return false
}

func (d Document) CountClaimStatus(status ClaimTagStatus) int {
	count := 0
	for _, block := range d.Blocks {
		for _, tag := range block.ClaimTags {
			if tag.Status == status {
				count++
			}
		}
	}
	return count
}

func (d Document) CountResolvedInferred() int {
	count := 0
	for _, block := range d.Blocks {
		for _, tag := range block.ClaimTags {
			if tag.Status == ClaimTagResolved || (tag.Status == ClaimTagInferred && tag.Resolved) {
				count++
			}
		}
	}
	return count
}

func MustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func ExtractSentences(text string) []string {
	text = NormalizeText(text)
	if text == "" {
		return nil
	}
	re := regexp.MustCompile(`(?m)([^.!?\n]+[.!?]?|[^.!?\n]+$)`)
	matches := re.FindAllString(text, -1)
	var sentences []string
	for _, match := range matches {
		clean := strings.TrimSpace(match)
		if clean != "" {
			sentences = append(sentences, clean)
		}
	}
	return sentences
}

func splitParagraphs(text string) []string {
	parts := strings.Split(text, "\n\n")
	var result []string
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean != "" {
			result = append(result, clean)
		}
	}
	if len(result) == 0 {
		return []string{""}
	}
	return result
}
