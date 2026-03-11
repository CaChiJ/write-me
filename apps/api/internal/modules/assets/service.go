package assets

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cachij/write-me/apps/api/internal/shared"
	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
)

type Asset struct {
	ID               string    `json:"id"`
	AssetType        string    `json:"assetType"`
	Title            string    `json:"title"`
	FileName         string    `json:"fileName"`
	MimeType         string    `json:"mimeType"`
	ExtractionStatus string    `json:"extractionStatus"`
	ExtractedText    string    `json:"extractedText"`
	CreatedAt        time.Time `json:"createdAt"`
}

type UploadInput struct {
	AssetType string
	Title     string
	FileName  string
	MimeType  string
	Data      []byte
}

type Repository interface {
	CreateAsset(ctx context.Context, asset StoredAsset) (Asset, error)
	ListAssets(ctx context.Context) ([]Asset, error)
	DeleteAsset(ctx context.Context, assetID string) error
	GetAssetsByIDs(ctx context.Context, assetIDs []string) ([]Asset, error)
}

type StoredAsset struct {
	ID               string
	AssetType        string
	Title            string
	FileName         string
	MimeType         string
	Data             []byte
	ExtractionStatus string
	ExtractedText    string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Upload(ctx context.Context, input UploadInput) (Asset, error) {
	if len(input.Data) == 0 {
		return Asset{}, shared.NewAppError(http.StatusBadRequest, "asset_empty", "빈 파일은 업로드할 수 없습니다.")
	}

	extractedText, extractionStatus := extractText(input.FileName, input.MimeType, input.Data)
	if strings.TrimSpace(input.Title) == "" {
		input.Title = input.FileName
	}

	return s.repo.CreateAsset(ctx, StoredAsset{
		ID:               uuid.NewString(),
		AssetType:        strings.TrimSpace(input.AssetType),
		Title:            strings.TrimSpace(input.Title),
		FileName:         input.FileName,
		MimeType:         input.MimeType,
		Data:             input.Data,
		ExtractionStatus: extractionStatus,
		ExtractedText:    extractedText,
	})
}

func (s *Service) List(ctx context.Context) ([]Asset, error) {
	return s.repo.ListAssets(ctx)
}

func (s *Service) Delete(ctx context.Context, assetID string) error {
	return s.repo.DeleteAsset(ctx, assetID)
}

func (s *Service) GetByIDs(ctx context.Context, assetIDs []string) ([]Asset, error) {
	if len(assetIDs) == 0 {
		return nil, nil
	}
	return s.repo.GetAssetsByIDs(ctx, assetIDs)
}

func extractText(fileName string, mimeType string, data []byte) (string, string) {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch {
	case ext == ".txt" || strings.HasPrefix(mimeType, "text/plain"):
		return shared.NormalizeText(string(data)), "READY"
	case ext == ".md" || strings.Contains(mimeType, "markdown"):
		return shared.NormalizeText(string(data)), "READY"
	case ext == ".docx" || strings.Contains(mimeType, "wordprocessingml"):
		text, err := extractDOCX(data)
		if err != nil {
			return "", "ERROR"
		}
		return shared.NormalizeText(text), "READY"
	case ext == ".pdf" || strings.Contains(mimeType, "pdf"):
		text, err := extractPDF(data)
		if err != nil {
			return "", "ERROR"
		}
		if strings.TrimSpace(text) == "" {
			return "", "SCANNED_PDF_UNSUPPORTED"
		}
		return shared.NormalizeText(text), "READY"
	default:
		return shared.NormalizeText(string(data)), "READY"
	}
}

func extractDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, file := range reader.File {
		if file.Name != "word/document.xml" {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		decoder := xml.NewDecoder(rc)
		var builder strings.Builder
		for {
			token, err := decoder.Token()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return "", err
			}
			switch value := token.(type) {
			case xml.CharData:
				builder.WriteString(string(value))
				builder.WriteString(" ")
			}
		}
		return builder.String(), nil
	}
	return "", fmt.Errorf("word/document.xml not found")
}

func extractPDF(data []byte) (string, error) {
	file, err := os.CreateTemp("", "write-me-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return "", err
	}

	doc, reader, err := pdf.Open(file.Name())
	if err != nil {
		return "", err
	}
	defer doc.Close()

	buf, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}

	plain, err := io.ReadAll(buf)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
