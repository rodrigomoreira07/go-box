package files

import (
	"context"
	"strings"
)

// QuotaValidator acts as an interface boundary to users domain quota.
type QuotaValidator interface {
	CheckQuota(ctx context.Context, userID string, fileSize int64) error
	ConsumeQuota(ctx context.Context, userID string, fileSize int64) error
}

// FileService orchestrates metadata recording and S3 presigned url generation.
type FileService struct {
	repo     Repository
	storage  StorageService
	quotaVal QuotaValidator
}

// NewFileService creates a new FileService.
func NewFileService(repo Repository, storage StorageService, quotaVal QuotaValidator) *FileService {
	return &FileService{
		repo:     repo,
		storage:  storage,
		quotaVal: quotaVal,
	}
}

// RequestUploadURL handles validation, quota checks, and requests a presigned upload URL from S3.
func (s *FileService) RequestUploadURL(ctx context.Context, userID string, path string, size int64) (string, error) {
	if userID == "" {
		return "", ErrInvalidUserID
	}

	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
		return "", ErrInvalidPath
	}

	if strings.HasSuffix(path, "/") {
		return "", ErrDirectoryUpload
	}

	if err := s.quotaVal.CheckQuota(ctx, userID, size); err != nil {
		return "", err
	}

	url, err := s.storage.GenerateUploadURL(ctx, userID, path)
	if err != nil {
		return "", err
	}

	file := &File{
		UserID: userID,
		Path:   path,
		Size:   size,
		IsDir:  false,
	}

	if err := s.repo.Save(ctx, file); err != nil {
		return "", err
	}

	if err := s.quotaVal.ConsumeQuota(ctx, userID, size); err != nil {
		return "", err
	}

	return url, nil
}

// RequestDownloadURL verifies permissions and requests a presigned download URL from S3.
func (s *FileService) RequestDownloadURL(ctx context.Context, userID string, path string) (string, error) {
	if userID == "" {
		return "", ErrInvalidUserID
	}

	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") {
		return "", ErrInvalidPath
	}

	file, err := s.repo.FindByPath(ctx, userID, path)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", ErrFileNotFound
	}

	return s.storage.GenerateDownloadURL(ctx, userID, path)
}
