package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/shared"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type Settings struct {
	DefaultProvider shared.ProviderType `json:"defaultProvider"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

type Repository interface {
	EnsureAdmin(ctx context.Context, email string, passwordHash string) error
	FindUserByEmail(ctx context.Context, email string) (User, string, error)
	GetUserBySessionHash(ctx context.Context, tokenHash string) (User, error)
	CreateAuthSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error
	DeleteAuthSession(ctx context.Context, tokenHash string) error
	GetSettings(ctx context.Context) (Settings, error)
	UpdateSettings(ctx context.Context, defaultProvider shared.ProviderType) (Settings, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BootstrapAdmin(ctx context.Context, email string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return shared.WrapAppError(http.StatusInternalServerError, "password_hash_failed", "관리자 계정을 초기화하지 못했습니다.", err)
	}
	return s.repo.EnsureAdmin(ctx, email, string(hash))
}

func (s *Service) Login(ctx context.Context, email string, password string) (User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, passwordHash, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return User{}, "", shared.NewAppError(http.StatusUnauthorized, "login_failed", "이메일 또는 비밀번호가 올바르지 않습니다.")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return User{}, "", shared.NewAppError(http.StatusUnauthorized, "login_failed", "이메일 또는 비밀번호가 올바르지 않습니다.")
	}

	token, err := randomToken()
	if err != nil {
		return User{}, "", shared.WrapAppError(http.StatusInternalServerError, "session_token_failed", "로그인 세션을 만들지 못했습니다.", err)
	}
	if err := s.repo.CreateAuthSession(ctx, user.ID, hashToken(token), time.Now().Add(30*24*time.Hour)); err != nil {
		return User{}, "", shared.WrapAppError(http.StatusInternalServerError, "session_create_failed", "로그인 세션을 저장하지 못했습니다.", err)
	}
	return user, token, nil
}

func (s *Service) GetMe(ctx context.Context, token string) (User, error) {
	if strings.TrimSpace(token) == "" {
		return User{}, shared.NewAppError(http.StatusUnauthorized, "unauthorized", "로그인이 필요합니다.")
	}
	user, err := s.repo.GetUserBySessionHash(ctx, hashToken(token))
	if err != nil {
		return User{}, shared.NewAppError(http.StatusUnauthorized, "unauthorized", "로그인이 필요합니다.")
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	err := s.repo.DeleteAuthSession(ctx, hashToken(token))
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (s *Service) GetSettings(ctx context.Context) (Settings, error) {
	return s.repo.GetSettings(ctx)
}

func (s *Service) UpdateSettings(ctx context.Context, defaultProvider shared.ProviderType) (Settings, error) {
	return s.repo.UpdateSettings(ctx, defaultProvider)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
