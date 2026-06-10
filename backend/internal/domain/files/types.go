package files

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrFileNotFound    = errors.New("file not found")
	ErrInvalidPath     = errors.New("invalid file path or traversal attempt")
	ErrDirectoryUpload = errors.New("cannot upload directly to a directory")
	ErrInvalidUserID   = errors.New("invalid or empty user ID")
)

// File represents a file or folder metadata record.
type File struct {
	ID        string
	UserID    string    // Cognito sub
	Path      string    // Relative path, e.g. "documents/invoice.pdf"
	Size      int64     // Size in bytes (0 for directories)
	SHA256    string    // SHA-256 hash (empty for directories)
	IsDir     bool      // Indicates if metadata represents a directory
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FileRepository defines database access patterns for file metadata.
type FileRepository interface {
	Save(ctx context.Context, file *File) error
	Delete(ctx context.Context, userID string, path string) error
	FindByPath(ctx context.Context, userID string, path string) (*File, error)
	ListByUser(ctx context.Context, userID string) ([]*File, error)
}

// StorageService defines interactions with AWS S3 for url generation.
type StorageService interface {
	GenerateUploadURL(ctx context.Context, userID string, path string) (string, error)
	GenerateDownloadURL(ctx context.Context, userID string, path string) (string, error)
}

// QuotaValidator acts as an interface boundary to users domain quota.
type QuotaValidator interface {
	ValidateAndReserve(ctx context.Context, userID string, fileSize int64) error
}

// FileService orchestrates metadata recording and S3 presigned url generation.
type FileService struct {
	repo     FileRepository
	storage  StorageService
	quotaVal QuotaValidator
}

// NewFileService creates a new FileService.
func NewFileService(repo FileRepository, storage StorageService, quotaVal QuotaValidator) *FileService {
	return &FileService{
		repo:     repo,
		storage:  storage,
		quotaVal: quotaVal,
	}
}

// RequestUploadURL handles validation, quota checks, and requests a presigned upload URL from S3.
func (s *FileService) RequestUploadURL(ctx context.Context, userID string, path string, size int64) (string, error) {
	return "", ErrNotImplemented
}

// RequestDownloadURL verifies permissions and requests a presigned download URL from S3.
func (s *FileService) RequestDownloadURL(ctx context.Context, userID string, path string) (string, error) {
	return "", ErrNotImplemented
}
