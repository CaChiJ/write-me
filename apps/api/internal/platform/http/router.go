package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/modules/application_specs"
	"github.com/cachij/write-me/apps/api/internal/modules/assets"
	"github.com/cachij/write-me/apps/api/internal/modules/identity"
	"github.com/cachij/write-me/apps/api/internal/modules/writing_sessions"
	"github.com/cachij/write-me/apps/api/internal/platform/bridge"
	"github.com/cachij/write-me/apps/api/internal/platform/config"
	"github.com/cachij/write-me/apps/api/internal/shared"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

const sessionCookieName = "write_me_session"

type Server struct {
	cfg      config.Config
	identity *identity.Service
	assets   *assets.Service
	specs    *application_specs.Service
	sessions *writing_sessions.Service
	bridge   *bridge.AssistantProvider
}

func NewRouter(cfg config.Config, identityService *identity.Service, assetService *assets.Service, specService *application_specs.Service, sessionService *writing_sessions.Service, bridgeProvider *bridge.AssistantProvider) http.Handler {
	server := &Server{
		cfg:      cfg,
		identity: identityService,
		assets:   assetService,
		specs:    specService,
		sessions: sessionService,
		bridge:   bridgeProvider,
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AppOrigin},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", server.handleHealth)

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", server.handleLogin)
		api.Post("/auth/logout", server.handleLogout)

		api.Group(func(private chi.Router) {
			private.Use(server.authMiddleware)

			private.Get("/me", server.handleMe)
			private.Get("/settings", server.handleGetSettings)
			private.Put("/settings", server.handleUpdateSettings)

			private.Get("/assets", server.handleListAssets)
			private.Post("/assets", server.handleUploadAsset)
			private.Delete("/assets/{assetID}", server.handleDeleteAsset)

			private.Post("/application-specs/parse", server.handleParsePosting)

			private.Get("/dashboard", server.handleDashboard)

			private.Post("/sessions", server.handleCreateSession)
			private.Get("/sessions/{sessionID}/workspace", server.handleGetWorkspace)
			private.Patch("/sessions/{sessionID}", server.handleUpdateSession)
			private.Patch("/sessions/{sessionID}/questions/{questionID}/draft", server.handleUpdateDraft)
			private.Post("/sessions/{sessionID}/questions/{questionID}/claims/{claimID}/resolve", server.handleResolveClaim)
			private.Post("/sessions/{sessionID}/generate", server.handleGenerateDraft)
			private.Post("/sessions/{sessionID}/chat/turns", server.handleChatTurn)
			private.Post("/sessions/{sessionID}/tools/{toolName}/execute", server.handleToolAction)
			private.Get("/sessions/{sessionID}/compare", server.handleGetCompare)
			private.Post("/sessions/{sessionID}/suggestions/{suggestionID}/apply", server.handleApplySuggestion)
			private.Get("/sessions/{sessionID}/reviews/latest", server.handleGetLatestReview)
			private.Post("/sessions/{sessionID}/reviews", server.handleRunReview)
			private.Post("/sessions/{sessionID}/finalize", server.handleFinalizeSession)
		})
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"timestamp": time.Now().UTC(),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}

	user, token, err := s.identity.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
		MaxAge:   30 * 24 * 60 * 60,
	})
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(sessionCookieName)
	if cookie != nil {
		_ = s.identity.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r.Context())
	if !ok {
		writeError(w, shared.NewAppError(http.StatusUnauthorized, "unauthorized", "로그인이 필요합니다."))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := s.identity.GetSettings(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	health, healthErr := s.bridge.Health(r.Context())
	if healthErr != nil {
		health = map[shared.ProviderType]bool{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"settings":       settings,
		"providerHealth": health,
	})
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		DefaultProvider shared.ProviderType `json:"defaultProvider"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	settings, err := s.identity.UpdateSettings(r.Context(), payload.DefaultProvider)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": settings})
}

func (s *Server) handleListAssets(w http.ResponseWriter, r *http.Request) {
	result, err := s.assets.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"assets": result})
}

func (s *Server) handleUploadAsset(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		writeError(w, shared.WrapAppError(http.StatusBadRequest, "multipart_parse_failed", "업로드 요청을 해석하지 못했습니다.", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, shared.WrapAppError(http.StatusBadRequest, "file_missing", "업로드할 파일이 없습니다.", err))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, shared.WrapAppError(http.StatusBadRequest, "file_read_failed", "업로드 파일을 읽지 못했습니다.", err))
		return
	}

	asset, err := s.assets.Upload(r.Context(), assets.UploadInput{
		AssetType: r.FormValue("assetType"),
		Title:     r.FormValue("title"),
		FileName:  header.Filename,
		MimeType:  header.Header.Get("Content-Type"),
		Data:      data,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
}

func (s *Server) handleDeleteAsset(w http.ResponseWriter, r *http.Request) {
	if err := s.assets.Delete(r.Context(), chi.URLParam(r, "assetID")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleParsePosting(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		SourceText string `json:"sourceText"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	spec, err := s.specs.ParseAndStore(r.Context(), payload.SourceText)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"applicationSpec": spec})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	summary, err := s.sessions.Dashboard(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title             string              `json:"title"`
		Provider          shared.ProviderType `json:"provider"`
		ApplicationSpecID string              `json:"applicationSpecId"`
		SelectedAssetIDs  []string            `json:"selectedAssetIds"`
		InitialDraftText  string              `json:"initialDraftText"`
		ApplicationSpec   struct {
			CompanyName string                       `json:"companyName"`
			RoleName    string                       `json:"roleName"`
			SourceText  string                       `json:"sourceText"`
			Questions   []application_specs.Question `json:"questions"`
		} `json:"applicationSpec"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}

	var spec application_specs.Spec
	var err error
	if strings.TrimSpace(payload.ApplicationSpecID) != "" {
		spec, err = s.specs.Get(r.Context(), payload.ApplicationSpecID)
	} else {
		spec, err = s.specs.CreateManual(r.Context(), payload.ApplicationSpec.CompanyName, payload.ApplicationSpec.RoleName, payload.ApplicationSpec.Questions, payload.ApplicationSpec.SourceText)
	}
	if err != nil {
		writeError(w, err)
		return
	}

	session, err := s.sessions.CreateSession(r.Context(), writing_sessions.CreateSessionInput{
		Title:             payload.Title,
		Provider:          payload.Provider,
		ApplicationSpecID: spec.ID,
		SelectedAssetIDs:  payload.SelectedAssetIDs,
		InitialDraftText:  payload.InitialDraftText,
	}, spec)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	session, err := s.sessions.GetSession(r.Context(), chi.URLParam(r, "sessionID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session})
}

func (s *Server) handleUpdateSession(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		CurrentProvider shared.ProviderType `json:"currentProvider"`
		AutoReview      bool                `json:"autoReview"`
		AutoApply       bool                `json:"autoApply"`
		CountMode       shared.CountMode    `json:"countMode"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	applyMode := shared.ApplyModeManual
	if payload.AutoApply {
		applyMode = shared.ApplyModeAuto
	}
	session, err := s.sessions.UpdateSession(r.Context(), chi.URLParam(r, "sessionID"), writing_sessions.UpdateSessionInput{
		CurrentProvider: payload.CurrentProvider,
		AutoReview:      payload.AutoReview,
		AutoApply:       payload.AutoApply,
		CountMode:       payload.CountMode,
		ApplyMode:       applyMode,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session})
}

func (s *Server) handleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PlainText string           `json:"plainText"`
		Document  *shared.Document `json:"document"`
		Reason    string           `json:"reason"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	draft, err := s.sessions.UpdateDraft(r.Context(), writing_sessions.UpdateDraftInput{
		SessionID:  chi.URLParam(r, "sessionID"),
		QuestionID: chi.URLParam(r, "questionID"),
		PlainText:  payload.PlainText,
		Document:   payload.Document,
		Reason:     payload.Reason,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
}

func (s *Server) handleResolveClaim(w http.ResponseWriter, r *http.Request) {
	draft, err := s.sessions.ResolveClaimTag(r.Context(), chi.URLParam(r, "sessionID"), chi.URLParam(r, "questionID"), chi.URLParam(r, "claimID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
}

func (s *Server) handleGenerateDraft(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		QuestionID string `json:"questionId"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	suggestion, err := s.sessions.GenerateDraft(r.Context(), chi.URLParam(r, "sessionID"), payload.QuestionID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestion": suggestion})
}

func (s *Server) handleChatTurn(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		QuestionID string `json:"questionId"`
		Scope      string `json:"scope"`
		Message    string `json:"message"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	message, suggestion, err := s.sessions.SendChatTurn(r.Context(), chi.URLParam(r, "sessionID"), payload.QuestionID, payload.Scope, payload.Message)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":    message,
		"suggestion": suggestion,
	})
}

func (s *Server) handleToolAction(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		QuestionID string `json:"questionId"`
		Scope      string `json:"scope"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	message, suggestion, err := s.sessions.ExecuteToolAction(r.Context(), chi.URLParam(r, "sessionID"), payload.QuestionID, chi.URLParam(r, "toolName"), payload.Scope)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":    message,
		"suggestion": suggestion,
	})
}

func (s *Server) handleGetCompare(w http.ResponseWriter, r *http.Request) {
	questionID := strings.TrimSpace(r.URL.Query().Get("questionId"))
	if questionID == "" {
		writeError(w, shared.NewAppError(http.StatusBadRequest, "question_required", "questionId가 필요합니다."))
		return
	}
	suggestion, err := s.sessions.GetLatestSuggestion(r.Context(), chi.URLParam(r, "sessionID"), questionID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"suggestion": suggestion})
}

func (s *Server) handleApplySuggestion(w http.ResponseWriter, r *http.Request) {
	draft, err := s.sessions.ApplySuggestion(r.Context(), chi.URLParam(r, "suggestionID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
}

func (s *Server) handleGetLatestReview(w http.ResponseWriter, r *http.Request) {
	report, err := s.sessions.GetLatestReview(r.Context(), chi.URLParam(r, "sessionID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"review": report})
}

func (s *Server) handleRunReview(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		QuestionID string `json:"questionId"`
	}
	if err := decodeJSON(r.Body, &payload); err != nil {
		writeError(w, err)
		return
	}
	report, err := s.sessions.RunReview(r.Context(), chi.URLParam(r, "sessionID"), payload.QuestionID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"review": report})
}

func (s *Server) handleFinalizeSession(w http.ResponseWriter, r *http.Request) {
	if err := s.sessions.FinalizeSession(r.Context(), chi.URLParam(r, "sessionID")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, shared.NewAppError(http.StatusUnauthorized, "unauthorized", "로그인이 필요합니다."))
			return
		}
		user, err := s.identity.GetMe(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	})
}

type userContextKey struct{}

func currentUser(ctx context.Context) (identity.User, bool) {
	user, ok := ctx.Value(userContextKey{}).(identity.User)
	return user, ok
}

func decodeJSON(body io.ReadCloser, target any) error {
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return shared.WrapAppError(http.StatusBadRequest, "invalid_json", "요청 본문을 해석하지 못했습니다.", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, map[string]any{
			"error": map[string]any{
				"code":    appErr.Code,
				"message": appErr.Message,
			},
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]any{
		"error": map[string]any{
			"code":    "internal_error",
			"message": "서버 오류가 발생했습니다.",
		},
	})
}
