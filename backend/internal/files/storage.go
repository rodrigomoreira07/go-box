package files

import "context"

// StorageService defines interactions with AWS S3 for url generation.
type StorageService interface {
	GenerateUploadURL(ctx context.Context, userID string, path string) (string, error)
	GenerateDownloadURL(ctx context.Context, userID string, path string) (string, error)
}
