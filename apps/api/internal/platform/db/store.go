package db

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/modules/application_specs"
	"github.com/cachij/write-me/apps/api/internal/modules/assets"
	"github.com/cachij/write-me/apps/api/internal/modules/identity"
	"github.com/cachij/write-me/apps/api/internal/modules/writing_sessions"
	"github.com/cachij/write-me/apps/api/internal/shared"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations.sql
var migrationsSQL string

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, migrationsSQL)
	return err
}

func (s *Store) EnsureAdmin(ctx context.Context, email string, passwordHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM admin_users WHERE email = $1)`, strings.ToLower(email)).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err := tx.Exec(ctx, `
			INSERT INTO admin_users (id, email, password_hash)
			VALUES ($1, $2, $3)
		`, uuid.NewString(), strings.ToLower(email), passwordHash); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO app_settings (id, default_provider, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO NOTHING
	`, string(shared.ProviderCodexCLI)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (identity.User, string, error) {
	var user identity.User
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM admin_users
		WHERE email = $1
	`, strings.ToLower(email)).Scan(&user.ID, &user.Email, &passwordHash, &user.CreatedAt)
	if err != nil {
		return identity.User{}, "", err
	}
	return user, passwordHash, nil
}

func (s *Store) GetUserBySessionHash(ctx context.Context, tokenHash string) (identity.User, error) {
	var user identity.User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.created_at
		FROM auth_sessions s
		JOIN admin_users u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.expires_at > NOW()
	`, tokenHash).Scan(&user.ID, &user.Email, &user.CreatedAt)
	if err != nil {
		return identity.User{}, err
	}
	return user, nil
}

func (s *Store) CreateAuthSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO auth_sessions (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
	`, uuid.NewString(), userID, tokenHash, expiresAt)
	return err
}

func (s *Store) DeleteAuthSession(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

func (s *Store) GetSettings(ctx context.Context) (identity.Settings, error) {
	var settings identity.Settings
	err := s.pool.QueryRow(ctx, `
		SELECT default_provider, updated_at
		FROM app_settings
		WHERE id = 1
	`).Scan(&settings.DefaultProvider, &settings.UpdatedAt)
	if err != nil {
		return identity.Settings{}, err
	}
	return settings, nil
}

func (s *Store) UpdateSettings(ctx context.Context, defaultProvider shared.ProviderType) (identity.Settings, error) {
	var settings identity.Settings
	err := s.pool.QueryRow(ctx, `
		INSERT INTO app_settings (id, default_provider, updated_at)
		VALUES (1, $1, NOW())
		ON CONFLICT (id) DO UPDATE
		SET default_provider = EXCLUDED.default_provider,
			updated_at = NOW()
		RETURNING default_provider, updated_at
	`, string(defaultProvider)).Scan(&settings.DefaultProvider, &settings.UpdatedAt)
	if err != nil {
		return identity.Settings{}, err
	}
	return settings, nil
}

func (s *Store) CreateAsset(ctx context.Context, assetInput assets.StoredAsset) (assets.Asset, error) {
	var asset assets.Asset
	err := s.pool.QueryRow(ctx, `
		INSERT INTO source_assets (id, asset_type, title, file_name, mime_type, binary_data, extraction_status, extracted_text)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, asset_type, title, file_name, mime_type, extraction_status, extracted_text, created_at
	`, assetInput.ID, assetInput.AssetType, assetInput.Title, assetInput.FileName, assetInput.MimeType, assetInput.Data, assetInput.ExtractionStatus, assetInput.ExtractedText).Scan(
		&asset.ID, &asset.AssetType, &asset.Title, &asset.FileName, &asset.MimeType, &asset.ExtractionStatus, &asset.ExtractedText, &asset.CreatedAt,
	)
	return asset, err
}

func (s *Store) ListAssets(ctx context.Context) ([]assets.Asset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_type, title, file_name, mime_type, extraction_status, extracted_text, created_at
		FROM source_assets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []assets.Asset
	for rows.Next() {
		var asset assets.Asset
		if err := rows.Scan(&asset.ID, &asset.AssetType, &asset.Title, &asset.FileName, &asset.MimeType, &asset.ExtractionStatus, &asset.ExtractedText, &asset.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, asset)
	}
	return result, rows.Err()
}

func (s *Store) DeleteAsset(ctx context.Context, assetID string) error {
	commandTag, err := s.pool.Exec(ctx, `DELETE FROM source_assets WHERE id = $1`, assetID)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return shared.NewAppError(http.StatusNotFound, "asset_not_found", "자료를 찾을 수 없습니다.")
	}
	return nil
}

func (s *Store) GetAssetsByIDs(ctx context.Context, assetIDs []string) ([]assets.Asset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, asset_type, title, file_name, mime_type, extraction_status, extracted_text, created_at
		FROM source_assets
		WHERE id = ANY($1)
	`, assetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []assets.Asset
	for rows.Next() {
		var asset assets.Asset
		if err := rows.Scan(&asset.ID, &asset.AssetType, &asset.Title, &asset.FileName, &asset.MimeType, &asset.ExtractionStatus, &asset.ExtractedText, &asset.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, asset)
	}
	return result, rows.Err()
}

func (s *Store) CreateSpec(ctx context.Context, spec application_specs.Spec) (application_specs.Spec, error) {
	if spec.ID == "" {
		spec.ID = uuid.NewString()
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO application_specs (id, company_name, role_name, source_text, warnings, questions)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, spec.ID, spec.CompanyName, spec.RoleName, spec.SourceText, shared.MustJSON(spec.Warnings), shared.MustJSON(spec.Questions)); err != nil {
		return application_specs.Spec{}, err
	}
	return s.GetSpec(ctx, spec.ID)
}

func (s *Store) GetSpec(ctx context.Context, specID string) (application_specs.Spec, error) {
	var spec application_specs.Spec
	var warningsRaw []byte
	var questionsRaw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT id, company_name, role_name, source_text, warnings, questions, created_at
		FROM application_specs
		WHERE id = $1
	`, specID).Scan(&spec.ID, &spec.CompanyName, &spec.RoleName, &spec.SourceText, &warningsRaw, &questionsRaw, &spec.CreatedAt)
	if err != nil {
		return application_specs.Spec{}, err
	}
	if err := json.Unmarshal(warningsRaw, &spec.Warnings); err != nil {
		return application_specs.Spec{}, err
	}
	if err := json.Unmarshal(questionsRaw, &spec.Questions); err != nil {
		return application_specs.Spec{}, err
	}
	return spec, nil
}

func (s *Store) CreateSession(ctx context.Context, input writing_sessions.CreateSessionInput, spec application_specs.Spec) (writing_sessions.Session, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return writing_sessions.Session{}, err
	}
	defer tx.Rollback(ctx)

	sessionID := uuid.NewString()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = fmt.Sprintf("%s · %s", spec.CompanyName, spec.RoleName)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO writing_sessions (
			id, application_spec_id, title, status, current_provider, apply_mode, review_mode,
			auto_review, auto_apply, count_mode, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, FALSE, $8, NOW(), NOW())
	`, sessionID, spec.ID, title, string(shared.SessionStatusDraft), string(input.Provider), string(shared.ApplyModeManual), string(shared.ReviewModeBoth), string(shared.CountModeIncludeSpaces))
	if err != nil {
		return writing_sessions.Session{}, err
	}

	for _, assetID := range input.SelectedAssetIDs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO session_assets (session_id, asset_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, sessionID, assetID); err != nil {
			return writing_sessions.Session{}, err
		}
	}

	for _, question := range spec.Questions {
		document := shared.NewDocumentFromText("")
		if _, err := tx.Exec(ctx, `
			INSERT INTO question_drafts (
				id, session_id, question_id, title, prompt_text, char_limit, document_json, plain_text,
				inferred_count, resolved_inferred_count, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, '', 0, 0, NOW(), NOW())
		`, uuid.NewString(), sessionID, question.ID, question.Title, question.PromptText, question.CharLimit, shared.MustJSON(document)); err != nil {
			return writing_sessions.Session{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return writing_sessions.Session{}, err
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (writing_sessions.Session, error) {
	var session writing_sessions.Session
	var specID string
	err := s.pool.QueryRow(ctx, `
		SELECT id, application_spec_id, title, status, current_provider, apply_mode, review_mode,
		       auto_review, auto_apply, count_mode, created_at, updated_at
		FROM writing_sessions
		WHERE id = $1
	`, sessionID).Scan(
		&session.ID,
		&specID,
		&session.Title,
		&session.Status,
		&session.CurrentProvider,
		&session.ApplyMode,
		&session.ReviewMode,
		&session.AutoReview,
		&session.AutoApply,
		&session.CountMode,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return writing_sessions.Session{}, shared.NewAppError(http.StatusNotFound, "session_not_found", "세션을 찾을 수 없습니다.")
		}
		return writing_sessions.Session{}, err
	}

	spec, err := s.GetSpec(ctx, specID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	session.ApplicationSpec = spec

	assetsList, err := s.getSessionAssets(ctx, session.ID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	session.Assets = assetsList

	drafts, err := s.getDrafts(ctx, session.ID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	session.Drafts = drafts

	messages, err := s.getChatMessages(ctx, session.ID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	session.ChatMessages = messages

	pending, err := s.getPendingSuggestions(ctx, session.ID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	session.Pending = pending

	latestReview, err := s.GetLatestReview(ctx, session.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return writing_sessions.Session{}, err
	}
	session.LatestReview = latestReview

	return session, nil
}

func (s *Store) ListDashboard(ctx context.Context) (writing_sessions.DashboardSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ws.id, ws.title, ap.company_name, ap.role_name, ws.status, ws.updated_at,
		       COALESCE(SUM(GREATEST(qd.inferred_count - qd.resolved_inferred_count, 0)), 0) AS unresolved_count
		FROM writing_sessions ws
		JOIN application_specs ap ON ap.id = ws.application_spec_id
		LEFT JOIN question_drafts qd ON qd.session_id = ws.id
		GROUP BY ws.id, ws.title, ap.company_name, ap.role_name, ws.status, ws.updated_at
		ORDER BY ws.updated_at DESC
		LIMIT 12
	`)
	if err != nil {
		return writing_sessions.DashboardSummary{}, err
	}
	defer rows.Close()

	summary := writing_sessions.DashboardSummary{}
	for rows.Next() {
		var card writing_sessions.SessionCard
		if err := rows.Scan(&card.ID, &card.Title, &card.CompanyName, &card.RoleName, &card.Status, &card.UpdatedAt, &card.UnresolvedCount); err != nil {
			return writing_sessions.DashboardSummary{}, err
		}
		summary.RecentSessions = append(summary.RecentSessions, card)
	}
	if err := rows.Err(); err != nil {
		return writing_sessions.DashboardSummary{}, err
	}

	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM source_assets`).Scan(&summary.AssetCount); err != nil {
		return writing_sessions.DashboardSummary{}, err
	}
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM writing_sessions WHERE status = $1`, string(shared.SessionStatusFinalized)).Scan(&summary.ReadyCount); err != nil {
		return writing_sessions.DashboardSummary{}, err
	}
	return summary, nil
}

func (s *Store) UpdateDraft(ctx context.Context, input writing_sessions.UpdateDraftInput) (writing_sessions.DraftQuestion, error) {
	document := input.Document
	if document == nil {
		doc := shared.NewDocumentFromText(input.PlainText)
		document = &doc
	}
	plainText := document.PlainText()
	inferredCount := document.CountClaimStatus(shared.ClaimTagInferred)
	resolvedCount := document.CountResolvedInferred()

	commandTag, err := s.pool.Exec(ctx, `
		UPDATE question_drafts
		SET document_json = $1,
		    plain_text = $2,
		    inferred_count = $3,
		    resolved_inferred_count = $4,
		    updated_at = NOW()
		WHERE session_id = $5 AND question_id = $6
	`, shared.MustJSON(document), plainText, inferredCount, resolvedCount, input.SessionID, input.QuestionID)
	if err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	if commandTag.RowsAffected() == 0 {
		return writing_sessions.DraftQuestion{}, shared.NewAppError(http.StatusNotFound, "question_not_found", "문항을 찾을 수 없습니다.")
	}
	if _, err := s.pool.Exec(ctx, `
		UPDATE writing_sessions
		SET updated_at = NOW()
		WHERE id = $1
	`, input.SessionID); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	return s.getDraft(ctx, input.SessionID, input.QuestionID)
}

func (s *Store) CreateVersionSnapshot(ctx context.Context, sessionID string, questionID string, reason string, draft writing_sessions.DraftQuestion) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO session_versions (id, session_id, question_id, reason, document_json, plain_text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, uuid.NewString(), sessionID, questionID, reason, shared.MustJSON(draft.Document), draft.PlainText)
	return err
}

func (s *Store) CreateSuggestion(ctx context.Context, sessionID string, suggestion writing_sessions.Suggestion) (writing_sessions.Suggestion, error) {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO suggestions (
			id, session_id, question_id, source, scope, rationale, original_document_json, original_plain_text,
			suggested_document_json, suggested_plain_text, status, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
	`, suggestion.ID, sessionID, suggestion.QuestionID, suggestion.Source, suggestion.Scope, shared.MustJSON(suggestion.Rationale), shared.MustJSON(suggestion.OriginalDocument), suggestion.OriginalPlainText, shared.MustJSON(suggestion.SuggestedDocument), suggestion.SuggestedPlainText, suggestion.Status)
	if err != nil {
		return writing_sessions.Suggestion{}, err
	}
	return s.loadSuggestionByID(ctx, suggestion.ID)
}

func (s *Store) GetLatestSuggestion(ctx context.Context, sessionID string, questionID string) (*writing_sessions.Suggestion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM suggestions
		WHERE session_id = $1 AND question_id = $2 AND status = 'PENDING'
		ORDER BY created_at DESC
		LIMIT 1
	`, sessionID, questionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var suggestionID string
	if err := rows.Scan(&suggestionID); err != nil {
		return nil, err
	}
	suggestion, err := s.loadSuggestionByID(ctx, suggestionID)
	if err != nil {
		return nil, err
	}
	return &suggestion, nil
}

func (s *Store) ApplySuggestion(ctx context.Context, suggestionID string) (writing_sessions.DraftQuestion, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	defer tx.Rollback(ctx)

	suggestion, sessionID, err := s.loadSuggestionTx(ctx, tx, suggestionID)
	if err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	currentDraft, err := s.getDraftTx(ctx, tx, sessionID, suggestion.QuestionID)
	if err != nil {
		return writing_sessions.DraftQuestion{}, err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO session_versions (id, session_id, question_id, reason, document_json, plain_text, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, uuid.NewString(), sessionID, suggestion.QuestionID, "apply_suggestion", shared.MustJSON(currentDraft.Document), currentDraft.PlainText); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}

	inferredCount := suggestion.SuggestedDocument.CountClaimStatus(shared.ClaimTagInferred)
	resolvedCount := suggestion.SuggestedDocument.CountResolvedInferred()
	if _, err := tx.Exec(ctx, `
		UPDATE question_drafts
		SET document_json = $1,
		    plain_text = $2,
		    inferred_count = $3,
		    resolved_inferred_count = $4,
		    updated_at = NOW()
		WHERE session_id = $5 AND question_id = $6
	`, shared.MustJSON(suggestion.SuggestedDocument), suggestion.SuggestedPlainText, inferredCount, resolvedCount, sessionID, suggestion.QuestionID); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE suggestions
		SET status = 'APPLIED', applied_at = NOW()
		WHERE id = $1
	`, suggestionID); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE writing_sessions SET updated_at = NOW() WHERE id = $1
	`, sessionID); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	return s.getDraft(ctx, sessionID, suggestion.QuestionID)
}

func (s *Store) CreateChatMessage(ctx context.Context, sessionID string, message writing_sessions.ChatMessage) (writing_sessions.ChatMessage, error) {
	meta := message.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	err := s.pool.QueryRow(ctx, `
		INSERT INTO chat_messages (id, session_id, question_id, role, message_type, content, meta_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at
	`, message.ID, sessionID, message.QuestionID, message.Role, message.MessageType, message.Content, shared.MustJSON(meta)).Scan(&message.CreatedAt)
	if err != nil {
		return writing_sessions.ChatMessage{}, err
	}
	message.Meta = meta
	if _, err := s.pool.Exec(ctx, `UPDATE writing_sessions SET updated_at = NOW() WHERE id = $1`, sessionID); err != nil {
		return writing_sessions.ChatMessage{}, err
	}
	return message, nil
}

func (s *Store) CreateToolAction(ctx context.Context, sessionID string, questionID string, toolName string, promptText string, chatMessageID string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tool_actions (id, session_id, question_id, tool_name, prompt_text, chat_message_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, uuid.NewString(), sessionID, questionID, toolName, promptText, chatMessageID)
	return err
}

func (s *Store) CreateReview(ctx context.Context, sessionID string, report writing_sessions.ReviewReport) (writing_sessions.ReviewReport, error) {
	payload := shared.MustJSON(report)
	err := s.pool.QueryRow(ctx, `
		INSERT INTO review_reports (id, session_id, question_id, report_json, ready_to_submit, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at
	`, report.ID, sessionID, report.QuestionID, payload, report.ReadyToSubmit).Scan(&report.CreatedAt)
	if err != nil {
		return writing_sessions.ReviewReport{}, err
	}
	if _, err := s.pool.Exec(ctx, `UPDATE writing_sessions SET updated_at = NOW() WHERE id = $1`, sessionID); err != nil {
		return writing_sessions.ReviewReport{}, err
	}
	return report, nil
}

func (s *Store) GetLatestReview(ctx context.Context, sessionID string) (*writing_sessions.ReviewReport, error) {
	var raw []byte
	err := s.pool.QueryRow(ctx, `
		SELECT report_json
		FROM review_reports
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, sessionID).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	report, err := writing_sessions.ParseStoredReview(raw)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (s *Store) ResolveClaimTag(ctx context.Context, sessionID string, questionID string, claimID string) (writing_sessions.DraftQuestion, error) {
	draft, err := s.getDraft(ctx, sessionID, questionID)
	if err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	if !draft.Document.ResolveClaim(claimID) {
		return writing_sessions.DraftQuestion{}, shared.NewAppError(http.StatusNotFound, "claim_not_found", "해당 AI 추론 표시를 찾을 수 없습니다.")
	}
	return s.UpdateDraft(ctx, writing_sessions.UpdateDraftInput{
		SessionID:  sessionID,
		QuestionID: questionID,
		Document:   &draft.Document,
		Reason:     "resolve_claim",
	})
}

func (s *Store) FinalizeSession(ctx context.Context, sessionID string) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT question_id, document_json, plain_text
		FROM question_drafts
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var questionID string
		var documentRaw []byte
		var plainText string
		if err := rows.Scan(&questionID, &documentRaw, &plainText); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO session_versions (id, session_id, question_id, reason, document_json, plain_text, created_at)
			VALUES ($1, $2, $3, 'finalize', $4, $5, NOW())
		`, uuid.NewString(), sessionID, questionID, documentRaw, plainText); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE writing_sessions
		SET status = $1, finalized_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, string(shared.SessionStatusFinalized), sessionID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) UpdateSession(ctx context.Context, sessionID string, input writing_sessions.UpdateSessionInput) (writing_sessions.Session, error) {
	commandTag, err := s.pool.Exec(ctx, `
		UPDATE writing_sessions
		SET current_provider = $1,
		    auto_review = $2,
		    auto_apply = $3,
		    count_mode = $4,
		    apply_mode = $5,
		    updated_at = NOW()
		WHERE id = $6
	`, string(input.CurrentProvider), input.AutoReview, input.AutoApply, string(input.CountMode), string(input.ApplyMode), sessionID)
	if err != nil {
		return writing_sessions.Session{}, err
	}
	if commandTag.RowsAffected() == 0 {
		return writing_sessions.Session{}, shared.NewAppError(http.StatusNotFound, "session_not_found", "세션을 찾을 수 없습니다.")
	}
	return s.GetSession(ctx, sessionID)
}

func (s *Store) getSessionAssets(ctx context.Context, sessionID string) ([]assets.Asset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT a.id, a.asset_type, a.title, a.file_name, a.mime_type, a.extraction_status, a.extracted_text, a.created_at
		FROM session_assets sa
		JOIN source_assets a ON a.id = sa.asset_id
		WHERE sa.session_id = $1
		ORDER BY a.created_at DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []assets.Asset
	for rows.Next() {
		var asset assets.Asset
		if err := rows.Scan(&asset.ID, &asset.AssetType, &asset.Title, &asset.FileName, &asset.MimeType, &asset.ExtractionStatus, &asset.ExtractedText, &asset.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, asset)
	}
	return result, rows.Err()
}

func (s *Store) getDrafts(ctx context.Context, sessionID string) ([]writing_sessions.DraftQuestion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, question_id, title, prompt_text, char_limit, document_json, plain_text, inferred_count, resolved_inferred_count, updated_at
		FROM question_drafts
		WHERE session_id = $1
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []writing_sessions.DraftQuestion
	for rows.Next() {
		draft, err := scanDraft(rows)
		if err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	return drafts, rows.Err()
}

func (s *Store) getDraft(ctx context.Context, sessionID string, questionID string) (writing_sessions.DraftQuestion, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, question_id, title, prompt_text, char_limit, document_json, plain_text, inferred_count, resolved_inferred_count, updated_at
		FROM question_drafts
		WHERE session_id = $1 AND question_id = $2
	`, sessionID, questionID)
	return scanDraft(row)
}

func (s *Store) getDraftTx(ctx context.Context, tx pgx.Tx, sessionID string, questionID string) (writing_sessions.DraftQuestion, error) {
	row := tx.QueryRow(ctx, `
		SELECT id, question_id, title, prompt_text, char_limit, document_json, plain_text, inferred_count, resolved_inferred_count, updated_at
		FROM question_drafts
		WHERE session_id = $1 AND question_id = $2
	`, sessionID, questionID)
	return scanDraft(row)
}

func (s *Store) getChatMessages(ctx context.Context, sessionID string) ([]writing_sessions.ChatMessage, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, question_id, role, message_type, content, meta_json, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []writing_sessions.ChatMessage
	for rows.Next() {
		var message writing_sessions.ChatMessage
		var raw []byte
		if err := rows.Scan(&message.ID, &message.QuestionID, &message.Role, &message.MessageType, &message.Content, &raw, &message.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &message.Meta); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (s *Store) getPendingSuggestions(ctx context.Context, sessionID string) (map[string]*writing_sessions.Suggestion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id
		FROM suggestions
		WHERE session_id = $1 AND status = 'PENDING'
		ORDER BY created_at DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]*writing_sessions.Suggestion{}
	for rows.Next() {
		var suggestionID string
		if err := rows.Scan(&suggestionID); err != nil {
			return nil, err
		}
		suggestion, err := s.loadSuggestionByID(ctx, suggestionID)
		if err != nil {
			return nil, err
		}
		if _, exists := result[suggestion.QuestionID]; !exists {
			copied := suggestion
			result[suggestion.QuestionID] = &copied
		}
	}
	return result, rows.Err()
}

func (s *Store) loadSuggestionByID(ctx context.Context, suggestionID string) (writing_sessions.Suggestion, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT question_id, source, scope, rationale, original_document_json, original_plain_text,
		       suggested_document_json, suggested_plain_text, status, created_at
		FROM suggestions
		WHERE id = $1
	`, suggestionID)
	return scanSuggestion(row, suggestionID)
}

func (s *Store) loadSuggestionTx(ctx context.Context, tx pgx.Tx, suggestionID string) (writing_sessions.Suggestion, string, error) {
	var sessionID string
	row := tx.QueryRow(ctx, `
		SELECT session_id, question_id, source, scope, rationale, original_document_json, original_plain_text,
		       suggested_document_json, suggested_plain_text, status, created_at
		FROM suggestions
		WHERE id = $1
	`, suggestionID)
	suggestion, err := scanSuggestionWithSession(row, suggestionID, &sessionID)
	return suggestion, sessionID, err
}

func scanDraft(scanner interface{ Scan(dest ...any) error }) (writing_sessions.DraftQuestion, error) {
	var draft writing_sessions.DraftQuestion
	var raw []byte
	err := scanner.Scan(&draft.ID, &draft.QuestionID, &draft.Title, &draft.PromptText, &draft.CharLimit, &raw, &draft.PlainText, &draft.InferredCount, &draft.ResolvedInferredCount, &draft.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return writing_sessions.DraftQuestion{}, shared.NewAppError(http.StatusNotFound, "question_not_found", "문항을 찾을 수 없습니다.")
		}
		return writing_sessions.DraftQuestion{}, err
	}
	if err := json.Unmarshal(raw, &draft.Document); err != nil {
		return writing_sessions.DraftQuestion{}, err
	}
	return draft, nil
}

func scanSuggestion(scanner interface{ Scan(dest ...any) error }, suggestionID string) (writing_sessions.Suggestion, error) {
	var suggestion writing_sessions.Suggestion
	var rationaleRaw []byte
	var originalRaw []byte
	var suggestedRaw []byte
	suggestion.ID = suggestionID
	err := scanner.Scan(&suggestion.QuestionID, &suggestion.Source, &suggestion.Scope, &rationaleRaw, &originalRaw, &suggestion.OriginalPlainText, &suggestedRaw, &suggestion.SuggestedPlainText, &suggestion.Status, &suggestion.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return writing_sessions.Suggestion{}, shared.NewAppError(http.StatusNotFound, "suggestion_not_found", "수정 제안을 찾을 수 없습니다.")
		}
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(rationaleRaw, &suggestion.Rationale); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(originalRaw, &suggestion.OriginalDocument); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(suggestedRaw, &suggestion.SuggestedDocument); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	return suggestion, nil
}

func scanSuggestionWithSession(scanner interface{ Scan(dest ...any) error }, suggestionID string, sessionID *string) (writing_sessions.Suggestion, error) {
	var suggestion writing_sessions.Suggestion
	var rationaleRaw []byte
	var originalRaw []byte
	var suggestedRaw []byte
	suggestion.ID = suggestionID
	err := scanner.Scan(sessionID, &suggestion.QuestionID, &suggestion.Source, &suggestion.Scope, &rationaleRaw, &originalRaw, &suggestion.OriginalPlainText, &suggestedRaw, &suggestion.SuggestedPlainText, &suggestion.Status, &suggestion.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return writing_sessions.Suggestion{}, shared.NewAppError(http.StatusNotFound, "suggestion_not_found", "수정 제안을 찾을 수 없습니다.")
		}
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(rationaleRaw, &suggestion.Rationale); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(originalRaw, &suggestion.OriginalDocument); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	if err := json.Unmarshal(suggestedRaw, &suggestion.SuggestedDocument); err != nil {
		return writing_sessions.Suggestion{}, err
	}
	return suggestion, nil
}

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
